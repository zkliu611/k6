package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/loadimpact/speedboat/lib"
	"github.com/loadimpact/speedboat/stats"
	"github.com/loadimpact/speedboat/stats/accumulate"
	"github.com/loadimpact/speedboat/stats/writer"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"os"
	"os/signal"
	"time"
)

var cmdRun = cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "once",
			Usage: "Run only a single test iteration, with one VU",
		},
		cli.StringFlag{
			Name:  "type, t",
			Usage: "Input file type, if not evident (url, js or postman)",
		},
		cli.StringSliceFlag{
			Name:  "vus, u",
			Usage: "Number of VUs to simulate",
		},
		cli.DurationFlag{
			Name:  "duration, d",
			Usage: "Test duration",
			Value: time.Duration(10) * time.Second,
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "Suppress the summary at the end of a test",
		},
		cli.StringFlag{
			Name:  "format, f",
			Usage: "Format for printed metrics (yaml, json, prettyjson)",
			Value: "yaml",
		},
		cli.DurationFlag{
			Name:  "interval, i",
			Usage: "Write periodic summaries",
		},
		cli.StringSliceFlag{
			Name:  "out, o",
			Usage: "Write metrics to a database",
		},
		cli.BoolFlag{
			Name:  "raw",
			Usage: "Instead of summaries, dump raw samples to stdout",
		},
		cli.StringSliceFlag{
			Name:  "select, s",
			Usage: "Include only named metrics",
		},
		cli.StringSliceFlag{
			Name:  "exclude, e",
			Usage: "Exclude named metrics",
		},
		cli.StringSliceFlag{
			Name:  "group-by, g",
			Usage: "Group metrics by tags",
		},
		cli.StringSliceFlag{
			Name:  "tag",
			Usage: "Additional metric tags",
		},
	},
	Action: actionRun,
}

func pollVURamping(ctx context.Context, t lib.Test) <-chan int {
	ch := make(chan int)
	startTime := time.Now()

	go func() {
		defer close(ch)

		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				ch <- t.VUsAt(time.Since(startTime))
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func actionRun(cc *cli.Context) error {
	once := cc.Bool("once")

	for _, out := range cc.StringSlice("out") {
		backend, err := parseBackend(out)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		stats.DefaultRegistry.Backends = append(stats.DefaultRegistry.Backends, backend)
	}

	var formatter writer.Formatter
	switch cc.String("format") {
	case "":
	case "json":
		formatter = writer.JSONFormatter{}
	case "prettyjson":
		formatter = writer.PrettyJSONFormatter{}
	case "yaml":
		formatter = writer.YAMLFormatter{}
	default:
		return cli.NewExitError("Unknown output format", 1)
	}

	stats.DefaultRegistry.ExtraTags = parseTags(cc.StringSlice("tag"))

	var summarizer *Summarizer
	if formatter != nil {
		filter := stats.MakeFilter(cc.StringSlice("exclude"), cc.StringSlice("select"))
		if cc.Bool("raw") {
			backend := &writer.Backend{
				Writer:    os.Stdout,
				Formatter: formatter,
			}
			backend.Filter = filter
			stats.DefaultRegistry.Backends = append(stats.DefaultRegistry.Backends, backend)
		} else {
			accumulator := accumulate.New()
			accumulator.Filter = filter
			accumulator.GroupBy = cc.StringSlice("group-by")
			stats.DefaultRegistry.Backends = append(stats.DefaultRegistry.Backends, accumulator)

			summarizer = &Summarizer{
				Accumulator: accumulator,
				Formatter:   formatter,
			}
		}
	}

	stages, err := parseStages(cc.StringSlice("vus"), cc.Duration("duration"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if once {
		stages = []lib.TestStage{
			lib.TestStage{Duration: 0, StartVUs: 1, EndVUs: 1},
		}
	}
	t := lib.Test{Stages: stages}

	var r lib.Runner
	switch len(cc.Args()) {
	case 0:
		cli.ShowAppHelp(cc)
		return nil
	case 1:
		filename := cc.Args()[0]
		typ := cc.String("type")
		if typ == "" {
			typ = guessType(filename)
		}

		if filename == "-" && typ == "" {
			typ = typeJS
		}

		runner, err := makeRunner(t, filename, typ)
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		r = runner
	default:
		return cli.NewExitError("Too many arguments!", 1)
	}

	if cc.Bool("plan") {
		data, err := yaml.Marshal(map[string]interface{}{
			"stages": stages,
		})
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		os.Stdout.Write(data)
		return nil
	}

	vus := lib.VUGroup{
		Pool: lib.VUPool{
			New: r.NewVU,
		},
		RunOnce: func(ctx context.Context, vu lib.VU) {
			if err := vu.RunOnce(ctx); err != nil {
				log.WithError(err).Error("Uncaught Error")
			}
		},
	}

	for i := 0; i < t.MaxVUs(); i++ {
		vu, err := vus.Pool.New()
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		vus.Pool.Put(vu)
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.TotalDuration())
	if once {
		ctx, cancel = context.WithCancel(context.Background())
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				if err := stats.Submit(); err != nil {
					log.WithError(err).Error("[Couldn't submit stats]")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	interval := cc.Duration("interval")
	if interval > 0 && summarizer != nil {
		go func() {
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ticker.C:
					if err := summarizer.Print(os.Stdout); err != nil {
						log.WithError(err).Error("Couldn't print statistics!")
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit)

		select {
		case <-quit:
			cancel()
		case <-ctx.Done():
		}
	}()

	if !cc.Bool("quiet") {
		log.WithFields(log.Fields{
			"at":     time.Now(),
			"length": t.TotalDuration(),
		}).Info("Starting test...")
	}

	if once {
		stats.Add(stats.Sample{Stat: &mVUs, Values: stats.Value(1)})

		vu, _ := vus.Pool.Get()
		if err := vu.RunOnce(ctx); err != nil {
			log.WithError(err).Error("Uncaught Error")
		}
	} else {
		vus.Start(ctx)
		scaleTo := pollVURamping(ctx, t)
	mainLoop:
		for {
			select {
			case num := <-scaleTo:
				vus.Scale(num)
				stats.Add(stats.Sample{
					Stat:   &mVUs,
					Values: stats.Value(float64(num)),
				})
			case <-ctx.Done():
				break mainLoop
			}
		}

		vus.Stop()
	}

	stats.Add(stats.Sample{Stat: &mVUs, Values: stats.Value(0)})
	stats.Submit()

	if summarizer != nil {
		if err := summarizer.Print(os.Stdout); err != nil {
			log.WithError(err).Error("Couldn't print statistics!")
		}
	}

	return nil
}
