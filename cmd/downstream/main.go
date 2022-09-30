package main

import (
	"context"

	"opentracing-playground/app"
	"opentracing-playground/app/downstream"

	"github.com/urfave/cli/v2"
)

// Main starts process in cli.
func Main(ctx context.Context, c *cli.Context) {
	server := downstream.Server{}
	server.Start(ctx, c.String("listen-addr"))
}

func main() {
	app := app.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "listen-addr",
				Value: ":8787",
			},
		},
		Main: Main,
	}

	app.Run()
}
