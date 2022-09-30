package jaeger

import (
	"opentracing-playground/cliflag"

	"github.com/urfave/cli/v2"
)

var defaultOption ConnectOption

func init() {
	cliflag.Register(&defaultOption)
}

// ConnectOption defines a generic connect option for all dialects.
type ConnectOption struct {
	CollectorHost string
}

// CliFlags returns cli flag list.
func (opt *ConnectOption) CliFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, &cli.StringFlag{
		Name:        "jaeger-collector-host",
		EnvVars:     []string{"JAEGER_COLLECTOR_HOST"},
		Required:    true,
		Destination: &opt.CollectorHost,
	})

	return flags
}
