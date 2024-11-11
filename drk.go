package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codingconcepts/drk/pkg/model"
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
	}).Level(lo.Ternary(*debug, zerolog.DebugLevel, zerolog.InfoLevel))

	cfg, err := loadConfig(*config)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	printConfig(cfg, &logger)

	if *dryRun {
		return
	}

	runner, err := model.NewRunner(cfg, *url, *driver, *duration, &logger)
	if err != nil {
		log.Fatalf("error creating runner: %v", err)
	}

	if err = runner.Run(); err != nil {
		log.Fatalf("error running config: %v", err)
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
