package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codingconcepts/drk/pkg/random"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var (
	logger zerolog.Logger
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

	logger = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stdout,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).Level(lo.Ternary(*debug, zerolog.DebugLevel, zerolog.InfoLevel))

	cfg, err := loadConfig(*config)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	printConfig(cfg)

	if *dryRun {
		return
	}

	runner, err := newRunner(cfg, *url, *driver, *duration)
	if err != nil {
		log.Fatalf("error creating runner: %v", err)
	}

	if err = runner.run(); err != nil {
		log.Fatalf("error running config: %v", err)
	}
}

type Runner struct {
	db       *sql.DB
	cfg      *Drk
	duration time.Duration
}

func newRunner(cfg *Drk, url, driver string, duration time.Duration) (*Runner, error) {
	db, err := sql.Open(driver, url)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &Runner{
		db:       db,
		cfg:      cfg,
		duration: duration,
	}, nil
}

func (r *Runner) run() error {
	var eg errgroup.Group

	for _, workflow := range r.cfg.Workflows {
		eg.Go(func() error {
			return r.runWorkflow(workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) runWorkflow(workflow Workflow) error {
	var eg errgroup.Group

	for vu := 0; vu < workflow.Vus; vu++ {
		eg.Go(func() error {
			return r.runVU(workflow)
		})
	}

	return eg.Wait()
}

func (r *Runner) runVU(workflow Workflow) error {
	// Prepare VU.
	vu := newVU()

	for _, query := range workflow.SetupQueries {
		act, ok := r.cfg.Activities[query]
		if !ok {
			return fmt.Errorf("missing activity: %q", query)
		}

		data, err := r.runQuery(vu, act)
		if err != nil {
			return fmt.Errorf("running query %q: %w", query, err)
		}

		vu.applyData(query, data)
	}

	// Stagger VU.
	vu.stagger(workflow.Queries)

	// Start VU.
	var eg errgroup.Group

	deadline := time.After(r.duration)

	for _, query := range workflow.Queries {
		act, ok := r.cfg.Activities[query.Name]
		if !ok {
			return fmt.Errorf("missing activity: %q", query)
		}

		eg.Go(func() error {
			return r.runActivity(vu, query.Name, act, query.Rate, deadline)
		})
	}

	return eg.Wait()
}

func (r *Runner) runActivity(vu *VU, name string, query Query, rate Rate, fin <-chan time.Time) error {
	ticks := rate.Ticker()

	for {
		select {
		case <-ticks:
			logger.Debug().Msgf("[QUERY] %s", name)

			data, err := r.runQuery(vu, query)
			if err != nil {
				logger.Error().Msgf("running query %q: %v", name, err)
			}
			logger.Debug().Msgf("[DATA] %+v", data)

			vu.applyData(name, data)

		case <-fin:
			return nil
		}
	}
}

func (r *Runner) runQuery(vu *VU, query Query) ([]map[string]any, error) {
	args, err := vu.generateArgs(query.Args)
	if err != nil {
		return nil, fmt.Errorf("generating args: %w", err)
	}

	logger.Debug().Msgf("[STMT] %s", query.Query)
	logger.Debug().Msgf("\t[ARGS] %v", args)

	switch query.Type {
	case "query":
		rows, err := r.db.Query(query.Query, args...)
		if err != nil {
			return nil, fmt.Errorf("running query: %w", err)
		}

		data, err := readRows(rows)
		if err != nil {
			return nil, fmt.Errorf("reading rows: %w", err)
		}
		return data, nil

	case "exec":
		_, err = r.db.Exec(query.Query, args...)
		if err != nil {
			return nil, fmt.Errorf("running query: %w", err)
		}
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported query type: %q", query.Type)
	}
}

func (vu *VU) stagger(queries []WorkflowQuery) {
	// Stagger using any time between now and the average query tick.
	sumTicks := lo.SumBy(queries, func(a WorkflowQuery) time.Duration {
		return a.Rate.tickerInterval
	})

	avgTicks := sumTicks / time.Duration(len(queries))

	staggerDuration := random.Interval(0, avgTicks)
	time.Sleep(staggerDuration)
}

func (vu *VU) applyData(query string, data []map[string]any) {
	vu.dataMu.Lock()
	defer vu.dataMu.Unlock()

	vu.data[query] = data
}

func (vu *VU) generateArgs(args []Arg) ([]any, error) {
	var values []any

	for _, arg := range args {
		v, err := arg.generator(vu)
		if err != nil {
			return nil, fmt.Errorf("generating value for arg: %w", err)
		}

		values = append(values, v)
	}

	return values, nil
}

func printConfig(cfg *Drk) {
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

func loadConfig(path string) (*Drk, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var cfg Drk
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	return &cfg, nil
}

type VU struct {
	// Map of query names to columns to rows.
	dataMu sync.RWMutex
	data   map[string][]map[string]any
}

func newVU() *VU {
	return &VU{
		data: map[string][]map[string]any{},
	}
}

type genFunc func(*VU) (any, error)

type Drk struct {
	Workflows  map[string]Workflow `yaml:"workflows"`
	Activities map[string]Query    `yaml:"activities"`
}

type WorkflowQuery struct {
	Name string `yaml:"name"`
	Rate Rate   `yaml:"rate"`
}

type Query struct {
	Type  string `yaml:"type"`
	Args  []Arg  `yaml:"args"`
	Query string `yaml:"query"`
}

type Rate struct {
	Times    int
	Interval time.Duration

	tickerInterval time.Duration
}

func (r Rate) Ticker() <-chan time.Time {
	return time.Tick(r.tickerInterval)
}

func (r *Rate) UnmarshalYAML(node *yaml.Node) error {
	parts := strings.Split(node.Value, "/")

	var err error
	if r.Times, err = strconv.Atoi(parts[0]); err != nil {
		return fmt.Errorf("parsing times: %w", err)
	}

	if r.Interval, err = time.ParseDuration(parts[1]); err != nil {
		return fmt.Errorf("parsing interval: %w", err)
	}

	r.tickerInterval = r.Interval / time.Duration(r.Times)

	return nil
}

func (r Rate) String() string {
	return fmt.Sprintf("%d/%s", r.Times, r.Interval)
}

type Workflow struct {
	Vus          int             `yaml:"vus"`
	SetupQueries []string        `yaml:"setup_queries"`
	Queries      []WorkflowQuery `yaml:"queries"`
}

type Arg struct {
	Type string `yaml:"type"`

	generator genFunc
}

func (a *Arg) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	argType, err := parseField[string](raw, "type")
	if err != nil {
		return fmt.Errorf("parsing type: %w", err)
	}

	switch argType {
	case "gen":
		if a.generator, err = parseArgTypeGen(raw); err != nil {
			return fmt.Errorf("parsing gen arg type: %w", err)
		}

	case "ref":
		if a.generator, err = parseArgTypeRef(raw); err != nil {
			return fmt.Errorf("parsing ref arg type: %w", err)
		}

	default:
		if a.generator, err = parseArgTypeScalar(argType, raw); err != nil {
			return fmt.Errorf("parsing scalar arg type: %w", err)
		}
	}

	return nil
}

func parseArgTypeGen(raw map[string]any) (genFunc, error) {
	value, err := parseField[string](raw, "value")
	if err != nil {
		return nil, fmt.Errorf("parsing value: %w", err)
	}

	return func(vu *VU) (any, error) {
		g, ok := random.Replacements[value]
		if !ok {
			return nil, fmt.Errorf("missing generator: %q", value)
		}
		return g(), nil
	}, nil
}

func parseArgTypeScalar(argType string, raw map[string]any) (genFunc, error) {
	return func(vu *VU) (any, error) {
		switch strings.ToLower(argType) {
		case "int":
			min, err := parseField[int](raw, "min")
			if err != nil {
				return nil, fmt.Errorf("parsing min: %w", err)
			}

			max, err := parseField[int](raw, "max")
			if err != nil {
				return nil, fmt.Errorf("parsing max: %w", err)
			}

			return random.Int(min, max), nil

		case "float":
			min, err := parseField[float64](raw, "min")
			if err != nil {
				return nil, fmt.Errorf("parsing min: %w", err)
			}

			max, err := parseField[float64](raw, "max")
			if err != nil {
				return nil, fmt.Errorf("parsing max: %w", err)
			}

			return random.Float(min, max), nil

		case "timestamp":
			minStr, err := parseField[string](raw, "min")
			if err != nil {
				return nil, fmt.Errorf("parsing min: %w", err)
			}

			maxStr, err := parseField[string](raw, "max")
			if err != nil {
				return nil, fmt.Errorf("parsing max: %w", err)
			}

			min, err := time.Parse(time.RFC3339, minStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as timestamp: %w", err)
			}

			max, err := time.Parse(time.RFC3339, maxStr)
			if err != nil {
				return nil, fmt.Errorf("parsing max as timestamp: %w", err)
			}

			return random.Timestamp(min, max), nil

		default:
			return nil, fmt.Errorf("invalid scalar generator: %q", argType)
		}
	}, nil
}

func parseArgTypeRef(raw map[string]any) (genFunc, error) {
	queryRef, err := parseField[string](raw, "query")
	if err != nil {
		return nil, fmt.Errorf("parsing table: %w", err)
	}

	columnRef, err := parseField[string](raw, "column")
	if err != nil {
		return nil, fmt.Errorf("parsing column: %w", err)
	}

	genFunc := func(vu *VU) (any, error) {
		logger.Debug().Msgf("[REF] gen %s - %s", queryRef, columnRef)

		vu.dataMu.RLock()
		defer vu.dataMu.RUnlock()
		query, ok := vu.data[queryRef]
		if !ok {
			return nil, fmt.Errorf("missing query: %q", query)
		}

		if len(query) == 0 {
			return nil, fmt.Errorf("no data found for %s - %s", queryRef, columnRef)
		}

		row := rand.IntN(len(query))
		cell, ok := query[row][columnRef]
		if !ok {
			return nil, fmt.Errorf("missing column: %q", cell)
		}

		return cell, nil
	}

	return genFunc, err
}

func parseField[T any](m map[string]any, key string) (T, error) {
	valueRaw, ok := m[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("missing value")
	}

	value, ok := valueRaw.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("value is not of expected type")
	}

	return value, nil
}

func readRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting column names: %w", err)
	}

	values := make([]any, len(columns))
	scanArgs := make([]any, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var results []map[string]any

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scaning row: %w", err)
		}

		result := map[string]any{}

		for i, c := range columns {
			cellPtr := scanArgs[i]
			result[c] = *cellPtr.(*any)
		}

		results = append(results, result)
	}

	return results, nil
}
