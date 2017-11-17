/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2017 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package cloud

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	null "gopkg.in/guregu/null.v3"
)

type Executor struct {
	runLock sync.Mutex
	Logger  *log.Logger

	Client  *Client
	Archive *lib.Archive
	Name    string

	lock    sync.RWMutex
	running bool
}

func New(conf Config, r lib.Runner, src *lib.SourceData, opts lib.Options, version string) (*Executor, error) {
	if val, ok := opts.External["loadimpact"]; ok {
		if err := mapstructure.Decode(val, &conf); err != nil {
			log.Warn("Malformed loadimpact settings in script options")
		}
	}

	if conf.Name == "" {
		conf.Name = filepath.Base(src.Filename)
	}

	if conf.Token == "" && conf.DeprecatedToken != "" {
		log.Warn("K6CLOUD_TOKEN is deprecated and will be removed. Use K6_CLOUD_TOKEN instead.")
		conf.Token = conf.DeprecatedToken
	}

	return &Executor{
		Logger:  log.StandardLogger(),
		Client:  NewClient(conf.Token, conf.Host, version),
		Archive: r.MakeArchive(),
		Name:    conf.Name,
	}, nil
}

func (e *Executor) Init() error {
	return e.Client.ValidateConfig(e.Archive.Options)
}

func (e *Executor) Run(ctx context.Context, out chan<- []stats.Sample) error {
	e.runLock.Lock()
	defer e.runLock.Unlock()

	e.lock.Lock()
	e.running = true
	e.lock.Unlock()

	refID, err := e.Client.ArchiveUpload(e.Name, e.Archive)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			resp, err := e.Client.TestProgress(refID)
			if err != nil {
				e.Logger.WithError(err).Error("Couldn't get cloud test status")
				continue
				// return err
			}
			e.Logger.WithFields(log.Fields{
				"progress": resp.Progress,
				"status":   resp.Status,
			}).Debug("Received cloud execution status")

			if resp.Progress != 0 {
				e.Logger.WithField("progress", resp.Progress).Debug("-> Cloud execution ended")
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *Executor) IsRunning() bool {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.running
}

func (e *Executor) GetRootGroup() *lib.Group {
	return nil
}

func (e *Executor) SetLogger(l *log.Logger) {
	e.Logger = l
}

func (e *Executor) GetLogger() *log.Logger {
	return e.Logger
}

func (e *Executor) GetStages() []lib.Stage {
	return nil
}

func (e *Executor) SetStages(s []lib.Stage) {
}

func (e *Executor) GetIterations() int64 {
	return 0
}

func (e *Executor) GetEndIterations() null.Int {
	return null.Int{}
}

func (e *Executor) SetEndIterations(i null.Int) {
}

func (e *Executor) GetTime() time.Duration {
	return 0
}

func (e *Executor) GetEndTime() lib.NullDuration {
	return lib.NullDuration{}
}

func (e *Executor) SetEndTime(t lib.NullDuration) {
}

func (e *Executor) IsPaused() bool {
	return false
}

func (e *Executor) SetPaused(paused bool) {
}

func (e *Executor) GetVUs() int64 {
	return 0
}

func (e *Executor) SetVUs(vus int64) error {
	return nil
}

func (e *Executor) GetVUsMax() int64 {
	return 0
}

func (e *Executor) SetVUsMax(max int64) error {
	return nil
}
