package main

import (
	"log/slog"
	"os"

	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			commands.ReadCommand(),
			commands.WriteCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("failed to run vessellocations-pulsar-demo", slog.String("error", err.Error()))
	}
}
