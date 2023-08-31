package model

import (
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the entire contents of a config file.
type Config struct {
	Queries []Query `yaml:"queries"`
}

// Query represents a statement to run against the database.
type Query struct {
	Table     string     `yaml:"table"`
	Group     string     `yaml:"group"`
	Type      string     `yaml:"type"`
	Rate      Rate       `yaml:"rate"`
	Statement string     `yaml:"statement"`
	Arguments []Argument `yaml:"arguments"`
}

type Rate struct {
	RPS      int           `yaml:"rps"`
	Duration time.Duration `yaml:"duration"`
}

// Argument represents an argument to use in a query.
type Argument struct {
	Name      string     `yaml:"name"`
	Type      string     `yaml:"type"`
	Processor RawMessage `yaml:"processor"`
}

// Load config from a file
func LoadConfig(r io.Reader) (Config, error) {
	var c Config
	if err := yaml.NewDecoder(r).Decode(&c); err != nil {
		return Config{}, fmt.Errorf("parsing file: %w", err)
	}

	return c, nil
}
