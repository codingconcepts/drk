package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type genFunc func(*VU) (any, error)

type dependencyFunc func(*VU) bool

func dependencyFuncNoop(*VU) bool { return true }

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

	generator       genFunc
	dependencyCheck dependencyFunc
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
		if a.generator, a.dependencyCheck, err = parseArgTypeGen(raw); err != nil {
			return fmt.Errorf("parsing gen arg type: %w", err)
		}

	case "ref":
		if a.generator, a.dependencyCheck, err = parseArgTypeRef(raw); err != nil {
			return fmt.Errorf("parsing ref arg type: %w", err)
		}

	case "set":
		if a.generator, a.dependencyCheck, err = parseArgTypeSet(raw); err != nil {
			return fmt.Errorf("parsing set arg type: %w", err)
		}

	case "const":
		if a.generator, a.dependencyCheck, err = parseArgTypeConst(raw); err != nil {
			return fmt.Errorf("parsing const arg type: %w", err)
		}

	default:
		if a.generator, a.dependencyCheck, err = parseArgTypeScalar(argType, raw); err != nil {
			return fmt.Errorf("parsing scalar arg type: %w", err)
		}
	}

	return nil
}
