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

	"github.com/codingconcepts/drk/pkg/model"
	"github.com/codingconcepts/drk/pkg/repo"
	"github.com/codingconcepts/ring"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func main() {
	config := flag.String("config", "drk.yaml", "absolute or relative path to config file")
	url := flag.String("url", "", "database connection string")
	driver := flag.String("driver", "pgx", "database driver to use [pgx]")
	dryRun := flag.Bool("dry-run", false, "if specified, prints config and exits")
	debug := flag.Bool("debug", false, "enable verbose logging")
	duration := flag.Duration("duration", time.Minute*10, "total duration of simulation")
	flag.Parse()

	if *url == "" || *driver == "" || *config == "" {
		flag.Usage()
		os.Exit(2)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).Level(lo.Ternary(*debug, zerolog.DebugLevel, zerolog.WarnLevel))

	cfg, err := loadConfig(*config)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	printConfig(cfg, &logger)

	if *dryRun {
		return
	}

	db, err := sql.Open(*driver, *url)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	queryer := repo.NewDBRepo(db)

	runner, err := model.NewRunner(cfg, queryer, *url, *driver, *duration, &logger)
	if err != nil {
		log.Fatalf("error creating runner: %v", err)
	}

	if !*debug {
		go monitor(runner)
	}

	if err = runner.Run(); err != nil {
		log.Fatalf("error running config: %v", err)
	}
}

func monitor(r *model.Runner) {
	events := r.GetEventStream()
	printTicks := time.Tick(time.Second)

	eventCounts := map[string]int{}
	eventLatencies := map[string]*ring.Ring[time.Duration]{}

	for {
		select {
		case event := <-events:
			key := fmt.Sprintf("%s.%s", event.Workflow, event.Name)

			// Add to event count.
			eventCounts[key]++

			// Add to event latencies.
			if _, ok := eventLatencies[key]; !ok {
				eventLatencies[key] = ring.New[time.Duration](1000)
			}
			eventLatencies[key].Add(event.Duration)

		case <-printTicks:
			fmt.Print("\033[H\033[2J")

			w := tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)

			fmt.Fprintln(w, "Setup queries")
			fmt.Fprintf(w, "=============\n\n")
			writeEvent(w, eventCounts, eventLatencies, func(s string, _ int) bool {
				return strings.HasPrefix(s, "*")
			})

			fmt.Fprintf(w, "\n\n")

			fmt.Fprintln(w, "Queries")
			fmt.Fprintf(w, "=======\n\n")
			writeEvent(w, eventCounts, eventLatencies, func(s string, _ int) bool {
				return !strings.HasPrefix(s, "*")
			})

			w.Flush()
		}
	}
}

type filter func(string, int) bool

func writeEvent(w io.Writer, counts map[string]int, latencies map[string]*ring.Ring[time.Duration], f filter) {
	keys := lo.Keys(counts)
	sort.Strings(keys)

	fmt.Fprintln(w, "Query\tRequests\tAverage Latency")
	fmt.Fprintln(w, "-----\t--------\t---------------")

	for _, key := range lo.Filter(keys, f) {
		latencies := latencies[key].Slice()

		fmt.Fprintf(
			w,
			"%s\t%d\t%s\n",
			strings.TrimPrefix(key, "*"),
			counts[key],
			lo.Sum(latencies)/time.Duration(len(latencies)),
		)
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
