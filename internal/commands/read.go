package commands

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/config"
	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/core"
	"github.com/urfave/cli/v2"
)

func ReadCommand() *cli.Command {
	return &cli.Command{
		Name:    "read",
		Aliases: []string{"r"},
		Usage:   "read from pulsar",
		Action: func(cCtx *cli.Context) error {
			return Read()
		},
	}
}

func Read() error {
	ctx := context.Background()
	ootelShutdown, err := core.Init(ctx)
	if err != nil {
		slog.Error("unable to initialize core", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = ootelShutdown(ctx)
	}()

	readConfig, err := config.NewReadConfig()
	if err != nil {
		slog.Error("unable to gather read config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: readConfig.PulsarAddr})
	if err != nil {
		slog.Error("unable to create pulsar client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer client.Close()

	channel := make(chan pulsar.ConsumerMessage, 100)

	options := pulsar.ConsumerOptions{
		Topic:            readConfig.PulsarTopic,
		SubscriptionName: readConfig.PulsarSubscription,
		Type:             pulsar.Shared,
	}

	options.MessageChannel = channel

	consumer, err := client.Subscribe(options)
	if err != nil {
		slog.Error("unable to subscribe", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer consumer.Close()

	for cm := range channel {
		msg := cm.Message

		var parsedMsg *PulsarVesselLocation
		err := json.Unmarshal(msg.Payload(), &parsedMsg)
		if err != nil {
			slog.Error("failed to unmarshal vessel location", slog.String("error", err.Error()))
			continue
		}

		slog.Info("received new location", slog.Any("message", parsedMsg))

		err = consumer.Ack(msg)
		if err != nil {
			slog.Error("failed to acknowledge message", slog.String("error", err.Error()))
			continue
		}
	}

	return nil
}
