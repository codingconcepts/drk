package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codingconcepts/drk/internal/pkg/model"
	"github.com/codingconcepts/drk/internal/pkg/random"
	"github.com/codingconcepts/throttle"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	requestsMade int64
)

func main() {
	log.SetFlags(0)

	configPath := flag.String("c", "", "the absolute or relative path to the config file")
	url := flag.String("u", "", "url of the database (i.e. the connection string)")
	flag.Parse()

	if *configPath == "" || *url == "" {
		flag.Usage()
		os.Exit(2)
	}

	c, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, err := pgxpool.New(context.Background(), *url)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}
	defer db.Close()

	tables := model.Tables{
		Tables: make(map[string]model.Table),
	}

	if err = runQueries(db, c, &tables); err != nil {
		log.Fatalf("error running queries: %v", err)
	}
}

func loadConfig(filename string) (model.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return model.Config{}, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	return model.LoadConfig(file)
}

func runQueries(db *pgxpool.Pool, config model.Config, t *model.Tables) error {
	for _, query := range config.Queries {
		if err := runQuery(db, query, t); err != nil {
			return fmt.Errorf("running query for %q: %w", query.Name, err)
		}
	}

	return nil
}

func runQuery(db *pgxpool.Pool, query model.Query, t *model.Tables) error {
	thr := throttle.New(int64(query.Rate.RPS), time.Second)

	f := func() error {
		args, err := generateArguments(query.Arguments)
		if err != nil {
			return fmt.Errorf("generating arguments: %w", err)
		}

		if _, err = db.Exec(context.Background(), query.Statement, args...); err != nil {
			return fmt.Errorf("making query: %w", err)
		}

		return nil
	}

	if err := thr.DoFor(context.Background(), query.Rate.Duration, f); err != nil {
		return fmt.Errorf("running query: %w", err)
	}

	return nil
}

func generateArguments(args []model.Argument) ([]any, error) {
	out := make([]any, len(args))

	for i, arg := range args {
		switch arg.Type {
		case "gen":
			var g random.GenGenerator
			if err := arg.Processor.UnmarshalFunc(&g); err != nil {
				return nil, fmt.Errorf("parsing gen processor for %s: %w", arg.Name, err)
			}
			out[i] = g.Generate()
		case "set":
			var g random.SetGenerator
			if err := arg.Processor.UnmarshalFunc(&g); err != nil {
				return nil, fmt.Errorf("parsing gen processor for %s: %w", arg.Name, err)
			}
			out[i] = g.Generate()
		}
	}

	return out, nil
}
