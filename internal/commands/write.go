package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"alpineworks.io/wsdot"
	"alpineworks.io/wsdot/ferries"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/cespare/xxhash/v2"
	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/config"
	"github.com/michaelpeterswa/vessellocations-pulsar-demo/internal/core"
	"github.com/urfave/cli/v2"
)

type wrappedFerriesLocation struct {
	vl           ferries.VesselLocation
	locationHash string
}

type PulsarVesselLocation struct {
	VesselID     int         `json:"vessel_id"`
	VesselName   string      `json:"vessel_name"`
	Coordinates  Coordinates `json:"coordinates"`
	LocationHash string      `json:"location_hash"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func WriteCommand() *cli.Command {
	return &cli.Command{
		Name:    "write",
		Aliases: []string{"w"},
		Usage:   "write vessel locations from pulsar",
		Action: func(cCtx *cli.Context) error {
			return Write()
		},
	}
}

func Write() error {
	ctx := context.Background()
	ootelShutdown, err := core.Init(ctx)
	if err != nil {
		slog.Error("unable to initialize core", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		_ = ootelShutdown(ctx)
	}()

	writeConfig, err := config.NewWriteConfig()
	if err != nil {
		slog.Error("unable to gather write config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	wsdotClient, err := wsdot.NewWSDOTClient(
		wsdot.WithAPIKey(writeConfig.WsdotAPIKey),
		wsdot.WithHTTPClient(&http.Client{
			Timeout: 5 * time.Second,
		}),
	)
	if err != nil {
		slog.Error("unable to create wsdot client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ferriesClient, err := ferries.NewFerriesClient(wsdotClient)
	if err != nil {
		slog.Error("unable to create ferries client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pulsarClient, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               writeConfig.PulsarAddr,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		slog.Error("unable to create pulsar client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pulsarClient.Close()

	vesselLocationsProducer, err := pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: writeConfig.PulsarTopic,
	})
	if err != nil {
		slog.Error("unable to create pulsar producer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer vesselLocationsProducer.Close()

	ticker := time.NewTicker(writeConfig.LoopDuration)
	defer ticker.Stop()

	vesselLocationsMap := make(map[string]wrappedFerriesLocation)
	for range ticker.C {
		vesselLocations, err := ferriesClient.GetVesselLocations()
		if err != nil {
			slog.Error("unable to load vessel locations", slog.String("error", err.Error()))
		}

		for _, vesselLocation := range vesselLocations {
			vesselLocationHash := HashCoordinates(vesselLocation)

			existingLocation, exists := vesselLocationsMap[vesselLocation.VesselName]
			if existingLocation.locationHash != vesselLocationHash || !exists {
				slog.Info("vessel location changed", slog.String("vesselname", vesselLocation.VesselName), slog.Time("time", time.Now()))
				vesselLocationsMap[vesselLocation.VesselName] = wrappedFerriesLocation{
					vl:           vesselLocation,
					locationHash: vesselLocationHash,
				}

				pulsarVesselLocation := PulsarVesselLocation{
					VesselID:   vesselLocation.VesselID,
					VesselName: vesselLocation.VesselName,
					Coordinates: Coordinates{
						Latitude:  vesselLocation.Latitude,
						Longitude: vesselLocation.Longitude,
					},
					LocationHash: vesselLocationHash,
				}

				jsonPulsarVesselLocation, err := json.Marshal(pulsarVesselLocation)
				if err != nil {
					slog.Error("failed to marshal vessel location", slog.String("error", err.Error()))
					continue
				}

				_, err = vesselLocationsProducer.Send(ctx, &pulsar.ProducerMessage{
					Payload: jsonPulsarVesselLocation,
				})
				if err != nil {
					slog.Error("failed to send vessel location", slog.String("error", err.Error()))
					continue
				}

				slog.Info("sent new vessel location", slog.String("vesselname", vesselLocation.VesselName), slog.Time("time", time.Now()))
			}
		}
	}

	return nil
}

func HashCoordinates(f ferries.VesselLocation) string {
	return strconv.FormatUint(xxhash.Sum64String(fmt.Sprintf("%f-%f", f.Latitude, f.Longitude)), 16)
}
