package monitoring

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codingconcepts/drk/pkg/model"
	"github.com/codingconcepts/ring"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type PrintMode string

const (
	PrintModeTable PrintMode = "table"
	PrintModeLog   PrintMode = "log"
)

var (
	ValidPrintModes = map[string]struct{}{
		string(PrintModeTable): {},
		string(PrintModeLog):   {},
	}
)

type Printer struct {
	logger *zerolog.Logger
	mode   PrintMode
	clear  bool
}

func NewPrinter(mode PrintMode, clear bool, logger *zerolog.Logger) *Printer {
	return &Printer{
		logger: logger,
		mode:   mode,
		clear:  clear,
	}
}

func (p *Printer) Print(counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration]) {
	if p.clear {
		fmt.Print("\033[H\033[2J")
	}

	switch p.mode {
	case PrintModeLog:
		p.PrintLine(counts, errors, latencies)

	case PrintModeTable:
		p.PrintTable(counts, errors, latencies)
	}
}

// PrintTable clears the terminal and prints a summary of requests.
func (p *Printer) PrintTable(counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration]) {
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

// PrintLine adds new lines to the terminal containing a summary of requests.
func (p *Printer) PrintLine(counts, errors map[string]int, latencies map[string]*ring.Ring[time.Duration]) {
	keys := lo.Uniq(append(lo.Keys(counts), lo.Keys(errors)...))
	sort.Strings(keys)

	f := func(s string, _ int) bool {
		return !strings.HasPrefix(s, "*")
	}

	for _, key := range lo.Filter(keys, f) {
		latencies := latencies[key].Slice()
		errors := errors[key]
		counts := counts[key]

		p.logger.Info().
			Str("key", key).
			Int("counts", counts).
			Int("errors", errors).
			Dur("avg_latency", lo.Sum(latencies)/time.Duration(len(latencies))).
			Msg("")
	}
}

// PrintConfig displays the applications configuration in the terminal.
func (p *Printer) PrintConfig(cfg *model.Drk) {
	if p.clear {
		fmt.Print("\033[H\033[2J")
	}

	for name, workflow := range cfg.Workflows {
		p.logger.Info().Msgf("workflow: %s...", name)
		p.logger.Info().Msgf("\tvus: %d", workflow.Vus)

		p.logger.Info().Msgf("\tsetup queries:")
		for _, query := range workflow.SetupQueries {
			p.logger.Info().Msgf("\t\t- %s", query)
		}

		p.logger.Info().Msgf("\tworkflow queries:")
		for _, query := range workflow.Queries {
			p.logger.Info().Msgf("\t\t- %s (%s)", query.Name, query.Rate)
		}
	}
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
