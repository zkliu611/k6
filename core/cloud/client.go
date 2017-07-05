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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/loadimpact/k6/lib"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	RequestTimeout = 10 * time.Second
)

type LoadImpactConfig struct {
	ProjectId int    `mapstructure:"project_id"`
	Name      string `mapstructure:"name"`
}

func (cfg LoadImpactConfig) GetProjectId() int {
	env := os.Getenv("K6CLOUD_PROJECTID")
	if env != "" {
		id, err := strconv.Atoi(env)
		if err == nil && id > 0 {
			return id
		}
	}
	return cfg.ProjectId
}

func (cfg LoadImpactConfig) GetName(src *lib.SourceData) string {
	envName := os.Getenv("K6CLOUD_NAME")
	if envName != "" {
		return envName
	}

	if cfg.Name != "" {
		return cfg.Name
	}

	if src.Filename != "" && src.Filename != "-" {
		name := filepath.Base(src.Filename)
		if name != "" && name != "." {
			return name
		}
	}

	return "k6 test"
}

// Client handles communication with Load Impact cloud API.
type Client struct {
	client  *http.Client
	token   string
	baseURL string
	version string
}

func NewClient(token, host, version string) *Client {
	client := &http.Client{
		Timeout: RequestTimeout,
	}

	hostEnv := os.Getenv("K6CLOUD_HOST")
	if hostEnv != "" {
		host = hostEnv
	}
	if host == "" {
		host = "https://ingest.loadimpact.com"
	}

	baseURL := fmt.Sprintf("%s/v1", host)

	c := &Client{
		client:  client,
		token:   token,
		baseURL: baseURL,
		version: version,
	}
	return c
}

func (c *Client) NewRequest(method, url string, data interface{}) (*http.Request, error) {
	var buf io.Reader

	if data != nil {
		b, err := json.Marshal(&data)
		if err != nil {
			return nil, err
		}

		buf = bytes.NewBuffer(b)
	}

	return http.NewRequest(method, url, buf)
}

func (c *Client) Do(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))
	req.Header.Set("User-Agent", "k6cloud/"+c.version)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorln(err)
		}
	}()

	if err = checkResponse(resp); err != nil {
		return err
	}

	if v != nil {
		if err = json.NewDecoder(resp.Body).Decode(v); err == io.EOF {
			err = nil // Ignore EOF from empty body
		}
	}

	return err
}

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	if r.StatusCode == 401 {
		return ErrNotAuthenticated
	} else if r.StatusCode == 403 {
		return ErrNotAuthorized
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var payload ErrorResponsePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return errors.Wrap(err, "Non-standard API error response: "+string(data))
	}
	return payload.Error
}
