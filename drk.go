package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codingconcepts/drk/internal/pkg/model"
	"github.com/codingconcepts/drk/internal/pkg/runner"
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
	r := runner.New(db)
	for _, query := range config.Queries {
		if err := runQuery(r, query, t); err != nil {
			return fmt.Errorf("running query for %q: %w", query.Table, err)
		}
	}

	return nil
}

func runQuery(r *runner.Runner, query model.Query, t *model.Tables) error {
	thr := throttle.New(int64(query.Rate.RPS), time.Second)

	executions := make(chan struct{})
	done := make(chan struct{})

	f := func() error {
		executions <- struct{}{}
		return r.Run(query)
	}

	go printer(query, executions, done)

	if err := thr.DoFor(context.Background(), query.Rate.Duration, f); err != nil {
		return fmt.Errorf("running query: %w", err)
	}

	done <- struct{}{}
	return nil
}

func printer(query model.Query, executions, done <-chan struct{}) {
	defer fmt.Println()

	count := 0

	ticker := time.NewTicker(time.Millisecond * 100).C
	for {
		select {
		case <-ticker:
			fmt.Printf("%s: %d\r", query.Table, count)
		case <-executions:
			count++
		case <-done:
			return
		}
	}
}
