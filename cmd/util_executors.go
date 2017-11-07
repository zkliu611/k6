package cmd

import (
	"github.com/loadimpact/k6/core/cloud"
	"github.com/loadimpact/k6/core/local"
	"github.com/loadimpact/k6/lib"
	"github.com/pkg/errors"
)

func newExecutor(t string, r lib.Runner, src *lib.SourceData) (lib.Executor, error) {
	switch t {
	case "", execLocal:
		return local.New(r), nil
	case execCloud:
		return cloud.New(r, src, Version), nil
	default:
		return nil, errors.Errorf("unknown executor type: %s", t)
	}
}
