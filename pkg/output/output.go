package output

import (
	"fmt"

	"github.com/CycodeLabs/cicd-leak-scanner/pkg/output/file"
	"github.com/CycodeLabs/cicd-leak-scanner/pkg/output/std"
)

type Output interface {
	Write(org string, repo string, workflow string, secret string) error
}

type OutputConfig struct {
	Filename string
}

type Opts func(*OutputConfig)

func WithFilename(filename string) Opts {
	return func(o *OutputConfig) {
		o.Filename = filename
	}
}

func New(method string, opts ...Opts) (Output, error) {
	cfg := &OutputConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	switch method {
	case "stdout":
		return std.New()
	case "file":
		return file.New(cfg.Filename)
	default:
		return nil, fmt.Errorf("invalid output method: %s", method)
	}
}
