package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/loadimpact/speedboat/stats"
	"github.com/loadimpact/speedboat/stats/accumulate"
	"github.com/loadimpact/speedboat/stats/writer"
	"github.com/urfave/cli"
	"os"
)

var summarizer *Summarizer

func setupLogging(cc *cli.Context) {
	if cc.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
}

func setupStats(cc *cli.Context) error {
	for _, out := range cc.GlobalStringSlice("out") {
		backend, err := parseBackend(out)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		stats.DefaultRegistry.Backends = append(stats.DefaultRegistry.Backends, backend)
	}

	var formatter writer.Formatter
	switch cc.GlobalString("format") {
	case "":
	case "json":
		formatter = writer.JSONFormatter{}
	case "prettyjson":
		formatter = writer.PrettyJSONFormatter{}
	case "yaml":
		formatter = writer.YAMLFormatter{}
	default:
		return errors.New("Unknown output format")
	}

	stats.DefaultRegistry.ExtraTags = parseTags(cc.GlobalStringSlice("tag"))

	if formatter != nil {
		filter := stats.MakeFilter(cc.GlobalStringSlice("exclude"), cc.GlobalStringSlice("select"))
		if cc.GlobalBool("raw") {
			backend := &writer.Backend{
				Writer:    os.Stdout,
				Formatter: formatter,
			}
			backend.Filter = filter
			stats.DefaultRegistry.Backends = append(stats.DefaultRegistry.Backends, backend)
		} else {
			accumulator := accumulate.New()
			accumulator.Filter = filter
			accumulator.GroupBy = cc.GlobalStringSlice("group-by")
			stats.DefaultRegistry.Backends = append(stats.DefaultRegistry.Backends, accumulator)

			summarizer = &Summarizer{
				Accumulator: accumulator,
				Formatter:   formatter,
			}
		}
	}

	return nil
}
