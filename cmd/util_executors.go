package cmd

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/loadimpact/k6/core/cloud"
	"github.com/loadimpact/k6/core/local"
	"github.com/loadimpact/k6/lib"
	"github.com/pkg/errors"
)

func newExecutor(t string, r lib.Runner, src *lib.SourceData, conf Config) (lib.Executor, error) {
	switch t {
	case "", execLocal:
		return local.New(r), nil
	case execCloud:
		config := conf.Collectors.Cloud
		if err := envconfig.Process("k6", &config); err != nil {
			return nil, err
		}
		return cloud.New(config, r, src, conf.Options, Version)
	default:
		return nil, errors.Errorf("unknown executor type: %s", t)
	}
}
