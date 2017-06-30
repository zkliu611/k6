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
	"os"
	"time"

	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	null "gopkg.in/guregu/null.v3"
)

type Executor struct {
	Client  *Client
	Archive *lib.Archive
}

func New(r lib.Runner, src *lib.SourceData, version string) (*Executor, error) {
	token := os.Getenv("K6CLOUD_TOKEN")
	opts := r.GetOptions()

	var extConfig LoadImpactConfig
	if val, ok := opts.External["loadimpact"]; ok {
		err := mapstructure.Decode(val, &extConfig)
		if err != nil {
			return nil, err
		}
	}

	return &Executor{
		Client:  NewClient(token, "", version),
		Archive: r.MakeArchive(),
	}, nil
}

func (e *Executor) Init() error {
	return nil
}

func (e *Executor) Run(ctx context.Context, out chan<- []stats.Sample) error {
	return nil
}

func (e *Executor) IsRunning() bool {
	return true
}

func (e *Executor) GetRootGroup() *lib.Group {
	return nil
}

func (e *Executor) SetLogger(l *log.Logger) {
}

func (e *Executor) GetLogger() *log.Logger {
	return nil
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
