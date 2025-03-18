package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codingconcepts/drk/pkg/model"
	"github.com/codingconcepts/drk/pkg/monitoring"
	"github.com/codingconcepts/drk/pkg/repo"
	"github.com/codingconcepts/env"
	"github.com/codingconcepts/ring"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/googleapis/go-sql-spanner"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	_ "github.com/sijms/go-ora/v2"
	"gopkg.in/yaml.v3"
)

var (
	version string
)

type envs struct {
	Config   string        `env:"CONFIG"`
	URL      string        `env:"URL"`
	Driver   string        `env:"DRIVER"`
	Duration time.Duration `env:"DURATION"`
	Retries  int           `env:"RETRIES"`
}

func main() {
	var e envs

	flag.StringVar(&e.Config, "config", "drk.yaml", "absolute or relative path to config file")
	flag.StringVar(&e.URL, "url", "", "database connection string")
	flag.StringVar(&e.Driver, "driver", "pgx", "database driver to use [mysql, spanner, pgx]")
	flag.DurationVar(&e.Duration, "duration", time.Minute*10, "total duration of simulation")
	flag.IntVar(&e.Retries, "retries", 1, "number of request retries")

	dryRun := flag.Bool("dry-run", false, "if specified, prints config and exits")
	debug := flag.Bool("debug", false, "enable verbose logging")
	showVersion := flag.Bool("version", false, "display the application version")
	pretty := flag.Bool("pretty", false, "print results to the terminal in a table")
	queryTimeout := flag.Duration("query-timeout", time.Second*5, "timeout for database queries")
	flag.Parse()

	// Override settings with values from the environment if provided.
	if err := env.Set(&e); err != nil {
		log.Fatalf("error setting environment variables: %v", err)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).Level(lo.Ternary(*debug, zerolog.DebugLevel, zerolog.InfoLevel))

	if *showVersion {
		logger.Info().Str("version", version).Msg("application info")
		return
	}

	if e.URL == "" || e.Driver == "" || e.Config == "" {
		flag.Usage()
		os.Exit(2)
	}

	cfg, err := loadConfig(e.Config)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	printConfig(cfg, &logger)

	if *dryRun {
		return
	}

	db, err := sql.Open(e.Driver, e.URL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	queryer := repo.NewDBRepo(db, *queryTimeout, e.Retries)

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err = db.PingContext(timeout); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	runner, err := model.NewRunner(cfg, queryer, e.URL, e.Driver, e.Duration, &logger)
	if err != nil {
		log.Fatalf("error creating runner: %v", err)
	}

	if !*debug {
		go monitor(runner, &logger, *pretty)
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)

	if err = runner.Run(); err != nil {
		log.Fatalf("error running config: %v", err)
	}
}

func monitor(r *model.Runner, logger *zerolog.Logger, pretty bool) {
	events := r.GetEventStream()
	printTicks := time.Tick(time.Second)

	errors := map[string]int{}
	counts := map[string]int{}
	latencies := map[string]*ring.Ring[time.Duration]{}

	for {
		select {
		case event := <-events:
			key := fmt.Sprintf("%s.%s", event.Workflow, event.Name)

			// Increment counts.
			if event.Err != nil {
				errors[key]++

				monitoring.MetricErrorCount.
					With(prometheus.Labels{"workflow": event.Workflow, "query": event.Name}).Inc()

				monitoring.MetricErrorDuration.
					With(prometheus.Labels{"workflow": event.Workflow, "query": event.Name}).
					Observe(event.Duration.Seconds())
			} else {
				counts[key]++

				monitoring.MetricRequestCount.
					With(prometheus.Labels{"workflow": event.Workflow, "query": event.Name}).Inc()

				monitoring.MetricRequestDuration.
					With(prometheus.Labels{"workflow": event.Workflow, "query": event.Name}).
					Observe(event.Duration.Seconds())
			}

			// Add to event latencies.
			if _, ok := latencies[key]; !ok {
				latencies[key] = ring.New[time.Duration](1000)
			}
			latencies[key].Add(event.Duration)

		case <-printTicks:
			if pretty {
				printTable(counts, errors, latencies)
			} else {
				printLine(counts, errors, latencies, logger)
			}
		}
	}
}

func printTable(counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration]) {
	fmt.Print("\033[H\033[2J")

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)

	fmt.Fprintln(w, "Setup queries")
	fmt.Fprintf(w, "=============\n\n")
	writeEvent(w, counts, errors, latencies, func(s string, _ int) bool {
		return strings.HasPrefix(s, "*")
	})

	fmt.Fprintf(w, "\n\n")

	fmt.Fprintln(w, "Queries")
	fmt.Fprintf(w, "=======\n\n")
	writeEvent(w, counts, errors, latencies, func(s string, _ int) bool {
		return !strings.HasPrefix(s, "*")
	})

	w.Flush()
}

type filter func(string, int) bool

func writeEvent(w io.Writer, counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration], f filter) {
	keys := lo.Uniq(append(lo.Keys(counts), lo.Keys(errors)...))
	sort.Strings(keys)

	fmt.Fprintln(w, "Query\tRequests\tErrors\tAverage Latency")
	fmt.Fprintln(w, "-----\t--------\t------\t---------------")

	for _, key := range lo.Filter(keys, f) {
		latencies := latencies[key].Slice()
		errors, hasErrors := errors[key]
		counts, hasCount := counts[key]

		fmt.Fprintf(
			w,
			"%s\t%d\t%d\t%s\n",
			strings.TrimPrefix(key, "*"),
			lo.Ternary(hasCount, counts, 0),
			lo.Ternary(hasErrors, errors, 0),
			lo.Sum(latencies)/time.Duration(len(latencies)),
		)
	}
}

func printLine(counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration], logger *zerolog.Logger) {
	keys := lo.Uniq(append(lo.Keys(counts), lo.Keys(errors)...))
	sort.Strings(keys)

	f := func(s string, _ int) bool {
		return !strings.HasPrefix(s, "*")
	}

	for _, key := range lo.Filter(keys, f) {
		latencies := latencies[key].Slice()
		errors, hasErrors := errors[key]
		counts, hasCount := counts[key]

		logger.Info().
			Str("key", key).
			Int("counts", lo.Ternary(hasCount, counts, 0)).
			Int("errors", lo.Ternary(hasErrors, errors, 0)).
			Dur("avg_latency", lo.Sum(latencies)/time.Duration(len(latencies))).
			Msg("")
	}
}

func printConfig(cfg *model.Drk, logger *zerolog.Logger) {
	for name, workflow := range cfg.Workflows {
		logger.Info().Msgf("workflow: %s...", name)
		logger.Info().Msgf("\tvus: %d", workflow.Vus)

		logger.Info().Msgf("\tsetup queries:")
		for _, query := range workflow.SetupQueries {
			logger.Info().Msgf("\t\t- %s", query)
		}

		logger.Info().Msgf("\tworkflow queries:")
		for _, query := range workflow.Queries {
			logger.Info().Msgf("\t\t- %s (%s)", query.Name, query.Rate)
		}
	}
}

func loadConfig(path string) (*model.Drk, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var cfg model.Drk
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	return &cfg, nil
}
