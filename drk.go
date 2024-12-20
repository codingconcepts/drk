package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	_ "github.com/codingconcepts/drk/pkg/driver"
	"github.com/codingconcepts/drk/pkg/model"
	"github.com/codingconcepts/drk/pkg/repo"
	"github.com/codingconcepts/env"
	"github.com/codingconcepts/ring"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

var version string

type envs struct {
	Config   string        `env:"CONFIG"`
	URL      string        `env:"URL"`
	Driver   string        `env:"DRIVER"`
	Duration time.Duration `env:"DURATION"`
}

func main() {
	var e envs

	flag.StringVar(&e.Config, "config", "drk.yaml", "absolute or relative path to config file")
	flag.StringVar(&e.URL, "url", "", "database connection string")
	flag.StringVar(&e.Driver, "driver", "pgx", "database driver to use [pgx, mysql, dsql]")
	flag.DurationVar(&e.Duration, "duration", time.Minute*10, "total duration of simulation")
	dryRun := flag.Bool("dry-run", false, "if specified, prints config and exits")
	debug := flag.Bool("debug", false, "enable verbose logging")
	showVersion := flag.Bool("version", false, "display the application version")
	pretty := flag.Bool("pretty", false, "print results to the terminal in a table")
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
	queryer := repo.NewDBRepo(db)

	runner, err := model.NewRunner(cfg, queryer, e.URL, e.Driver, e.Duration, &logger)
	if err != nil {
		log.Fatalf("error creating runner: %v", err)
	}

	if !*debug {
		go monitor(runner, &logger, *pretty)
	}

	if err = runner.Run(); err != nil {
		log.Fatalf("error running config: %v", err)
	}
}

func monitor(r *model.Runner, logger *zerolog.Logger, pretty bool) {
	events := r.GetEventStream()
	printTicks := time.Tick(time.Second)

	eventCounts := map[string]int{}
	errorCounts := map[string]int{}
	eventLatencies := map[string]*ring.Ring[time.Duration]{}

	for {
		select {
		case event := <-events:
			key := fmt.Sprintf("%s.%s", event.Workflow, event.Name)

			// Increment counts.
			if event.Err != nil {
				errorCounts[key]++
			} else {
				eventCounts[key]++
			}

			// Add to event latencies.
			if _, ok := eventLatencies[key]; !ok {
				eventLatencies[key] = ring.New[time.Duration](1000)
			}
			eventLatencies[key].Add(event.Duration)

		case <-printTicks:
			if pretty {
				printTable(eventCounts, errorCounts, eventLatencies)
			} else {
				printLine(eventCounts, errorCounts, eventLatencies, logger)
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
	keys := lo.Keys(counts)
	sort.Strings(keys)

	fmt.Fprintln(w, "Query\tRequests\tErrors\tAverage Latency")
	fmt.Fprintln(w, "-----\t--------\t------\t---------------")

	for _, key := range lo.Filter(keys, f) {
		latencies := latencies[key].Slice()

		fmt.Fprintf(
			w,
			"%s\t%d\t%d\t%s\n",
			strings.TrimPrefix(key, "*"),
			counts[key],
			errors[key],
			lo.Sum(latencies)/time.Duration(len(latencies)),
		)
	}
}

func printLine(counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration], logger *zerolog.Logger) {
	keys := lo.Keys(counts)
	sort.Strings(keys)

	f := func(s string, _ int) bool {
		return !strings.HasPrefix(s, "*")
	}

	for _, key := range lo.Filter(keys, f) {
		latencies := latencies[key].Slice()

		logger.Info().
			Str("key", key).
			Int("count", counts[key]).
			Int("errors", errors[key]).
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
