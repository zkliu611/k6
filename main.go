package main

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/loadimpact/speedboat/js"
	"github.com/loadimpact/speedboat/lib"
	"github.com/loadimpact/speedboat/postman"
	"github.com/loadimpact/speedboat/simple"
	"github.com/loadimpact/speedboat/stats"
	"github.com/loadimpact/speedboat/stats/influxdb"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	typeURL     = "url"
	typeJS      = "js"
	typePostman = "postman"
)

var mVUs = stats.Stat{Name: "vus", Type: stats.GaugeType}

func parseBackend(out string) (stats.Backend, error) {
	switch {
	case strings.HasPrefix(out, "influxdb+"):
		url := strings.TrimPrefix(out, "influxdb+")
		return influxdb.NewFromURL(url)
	default:
		return nil, errors.New("Unknown output destination")
	}
}

func parseStages(vus []string, total time.Duration) (stages []lib.TestStage, err error) {
	if len(vus) == 0 {
		return []lib.TestStage{
			lib.TestStage{Duration: total, StartVUs: 10, EndVUs: 10},
		}, nil
	}

	accountedTime := time.Duration(0)
	fluidStages := []int{}
	for i, spec := range vus {
		parts := strings.SplitN(spec, ":", 2)
		countParts := strings.SplitN(parts[0], "-", 2)

		stage := lib.TestStage{}

		// An absent first part means keep going from the last stage's end
		// If it's the first stage, just start with 0
		if countParts[0] == "" {
			if i > 0 {
				stage.StartVUs = stages[i-1].EndVUs
			}
		} else {
			start, err := strconv.ParseInt(countParts[0], 10, 64)
			if err != nil {
				return stages, err
			}
			stage.StartVUs = int(start)
		}

		// If an end is specified, use that, otherwise keep the VU level constant
		if len(countParts) > 1 && countParts[1] != "" {
			end, err := strconv.ParseInt(countParts[1], 10, 64)
			if err != nil {
				return stages, err
			}
			stage.EndVUs = int(end)
		} else {
			stage.EndVUs = stage.StartVUs
		}

		// If a time is specified, use that, otherwise mark the stage as "fluid", allotting it an
		// even slice of what time remains after all fixed stages are accounted for (may be 0)
		if len(parts) > 1 {
			duration, err := time.ParseDuration(parts[1])
			if err != nil {
				return stages, err
			}
			stage.Duration = duration
			accountedTime += duration
		} else {
			fluidStages = append(fluidStages, i)
		}

		stages = append(stages, stage)
	}

	// We're ignoring fluid stages if the fixed stages already take up all the allotted time
	// Otherwise, evenly divide the test's remaining time between all fluid stages
	if len(fluidStages) > 0 && accountedTime < total {
		fluidDuration := (total - accountedTime) / time.Duration(len(fluidStages))
		for _, i := range fluidStages {
			stage := stages[i]
			stage.Duration = fluidDuration
			stages[i] = stage
		}
	}

	return stages, nil
}

func parseTags(lines []string) stats.Tags {
	tags := make(stats.Tags)
	for _, line := range lines {
		idx := strings.IndexAny(line, ":=")
		if idx == -1 {
			tags[line] = line
			continue
		}

		key := line[:idx]
		val := line[idx+1:]
		if key == "" {
			key = val
		}
		tags[key] = val
	}
	return tags
}

func guessType(arg string) string {
	switch {
	case strings.Contains(arg, "://"):
		return typeURL
	case strings.HasSuffix(arg, ".js"):
		return typeJS
	case strings.HasSuffix(arg, ".postman_collection.json"):
		return typePostman
	}
	return ""
}

func readAll(filename string) ([]byte, error) {
	if filename == "-" {
		return ioutil.ReadAll(os.Stdin)
	}

	return ioutil.ReadFile(filename)
}

func makeRunner(t lib.Test, filename, typ string) (lib.Runner, error) {
	if typ == typeURL {
		return simple.New(filename), nil
	}

	bytes, err := readAll(filename)
	if err != nil {
		return nil, err
	}

	switch typ {
	case typeJS:
		return js.New(filename, string(bytes)), nil
	case typePostman:
		return postman.New(bytes)
	default:
		return nil, errors.New("Type ambiguous, please specify -t/--type")
	}
}

func main() {
	// Submit usage statistics for the closed beta
	invocation := Invocation{}
	invocationError := make(chan error, 1)

	go func() {
		// Set SUBMIT=false to prevent stat collection
		submitURL := os.Getenv("SB_SUBMIT")
		switch submitURL {
		case "false", "no":
			return
		case "":
			submitURL = "http://52.209.216.227:8080"
		}

		// Wait at most 2s for an invocation error to be reported
		select {
		case err := <-invocationError:
			invocation.Error = err.Error()
		case <-time.After(2 * time.Second):
		}

		// Submit stats to a specified server
		if err := invocation.Submit(submitURL); err != nil {
			log.WithError(err).Debug("Couldn't submit statistics")
		}
	}()

	// Free up -v and -h for our own flags
	cli.VersionFlag.Name = "version"
	cli.HelpFlag.Name = "help, ?"

	// Bootstrap the app from commandline flags
	app := cli.NewApp()
	app.Name = "speedboat"
	app.Usage = "A next-generation load generator"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		cmdRun,
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "More verbose output",
		},
	}
	app.Before = func(cc *cli.Context) error {
		if cc.GlobalBool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		invocation.PopulateWithContext(cc)
		return nil
	}
	invocationError <- app.Run(os.Args)
}
