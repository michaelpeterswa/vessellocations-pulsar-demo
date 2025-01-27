package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"alpineworks.io/ootel"
	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/config"
	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/logging"
)

func Init(ctx context.Context) (func(context.Context) error, error) {
	slogHandler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(slogHandler))

	c, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load main config: %w", err)
	}

	slogLevel, err := logging.LogLevelToSlogLevel(c.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert slog level: %w", err)
	}

	slog.SetLogLoggerLevel(slogLevel)

	ootelClient := ootel.NewOotelClient(
		ootel.WithMetricConfig(
			ootel.NewMetricConfig(
				c.MetricsEnabled,
				c.MetricsPort,
			),
		),
		ootel.WithTraceConfig(
			ootel.NewTraceConfig(
				c.TracingEnabled,
				c.TracingSampleRate,
				c.TracingService,
				c.TracingVersion,
			),
		),
	)

	shutdown, err := ootelClient.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ootel: %w", err)
	}

	return shutdown, nil
}
