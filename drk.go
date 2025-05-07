package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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

func main() {
	var e model.EnvironmentVariables

	flag.StringVar(&e.Config, "config", "drk.yaml", "absolute or relative path to config file")
	flag.StringVar(&e.URL, "url", "", "database connection string")
	flag.StringVar(&e.Driver, "driver", "pgx", "database driver to use [mysql, spanner, pgx]")
	flag.DurationVar(&e.Duration, "duration", time.Minute*10, "total duration of simulation")
	flag.DurationVar(&e.ConnectionLifetime, "connection-lifetime", time.Minute*1, "amount of time a connection can be reused")
	flag.IntVar(&e.Retries, "retries", 1, "number of request retries")
	flag.DurationVar(&e.QueryTimeout, "query-timeout", time.Second*5, "timeout for database queries")
	flag.BoolVar(&e.Debug, "debug", false, "show debugging logs")
	flag.BoolVar(&e.Sensitive, "sensitive", false, "show sensitive logs")
	flag.IntVar(&e.AverageWindowSize, "average-window-size", 1000, "number of request to derive an average latency for")
	flag.BoolVar(&e.NoColor, "no-color", false, "print logs without color")

	dryRun := flag.Bool("dry-run", false, "if specified, prints config and exits")
	showVersion := flag.Bool("version", false, "display the application version")
	mode := flag.String("output", "log", "type of metrics output to print [log, table]")
	clear := flag.Bool("clear", false, "clear the terminal before printing metrics")
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
		NoColor: e.NoColor,
	}).Level(lo.Ternary(e.Debug, zerolog.DebugLevel, zerolog.InfoLevel))

	if *showVersion {
		logger.Info().Str("version", version).Msg("application info")
		return
	}

	if e.Sensitive {
		logger.Info().Any("env args", e).Msg("sensitive")

		for _, env := range os.Environ() {
			logger.Info().Str("env", env).Msg("sensitive")
		}
	}

	if e.URL == "" || e.Driver == "" || e.Config == "" {
		flag.Usage()
		os.Exit(2)
	}

	if _, ok := monitoring.ValidPrintModes[*mode]; !ok {
		log.Fatalf("invalid output type: %q (should be one of: %v)", *mode, monitoring.ValidPrintModes)
	}

	cfg, err := loadConfig(e.Config)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	printer := monitoring.NewPrinter(monitoring.PrintMode(*mode), *clear, &logger)

	printer.PrintConfig(cfg)

	if *dryRun {
		return
	}

	db, err := sql.Open(e.Driver, e.URL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	db.SetConnMaxLifetime(e.ConnectionLifetime)
	logger.Debug().Msg("db connection established")

	queryer := repo.NewDBRepo(db, e.QueryTimeout, e.Retries)

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err = db.PingContext(timeout); err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	logger.Debug().Msg("db connection tested")

	runner, err := model.NewRunner(cfg, queryer, e, &logger)
	if err != nil {
		log.Fatalf("error creating runner: %v", err)
	}

	summaryC := make(chan struct{})
	go monitor(runner, e, printer, summaryC)

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)

	if err = runner.Run(); err != nil {
		log.Fatalf("error running config: %v", err)
	}

	// Tell the monitor function to print a summary, then wait
	// for it to finish using the same channel.
	summaryC <- struct{}{}
	<-summaryC
}

func monitor(r *model.Runner, e model.EnvironmentVariables, printer *monitoring.Printer, summary chan struct{}) {
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
				latencies[key] = ring.New[time.Duration](e.AverageWindowSize)
			}
			latencies[key].Add(event.Duration)

		case <-printTicks:
			printer.Print(counts, errors, latencies)

		case <-summary:
			printer.Print(counts, errors, latencies)

			// Allow the app to finish (the caller will be waiting on this).
			summary <- struct{}{}
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
