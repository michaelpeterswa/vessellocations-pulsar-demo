package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"error"`

	MetricsEnabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsPort    int  `env:"METRICS_PORT" envDefault:"8081"`

	TracingEnabled    bool    `env:"TRACING_ENABLED" envDefault:"true"`
	TracingSampleRate float64 `env:"TRACING_SAMPLERATE" envDefault:"1"`
	TracingService    string  `env:"TRACING_SERVICE" envDefault:"vessellocations-pulsar-demo"`
	TracingVersion    string  `env:"TRACING_VERSION"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

type WriteConfig struct {
	WsdotAPIKey  string        `env:"WSDOT_API_KEY"`
	PulsarAddr   string        `env:"PULSAR_ADDR"`
	PulsarTopic  string        `env:"PULSAR_TOPIC"`
	LoopDuration time.Duration `env:"LOOP_DURATION" envDefault:"15s"`
}

func NewWriteConfig() (*WriteConfig, error) {
	var cfg WriteConfig

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

type ReadConfig struct {
	PulsarAddr         string `env:"PULSAR_ADDR"`
	PulsarTopic        string `env:"PULSAR_TOPIC"`
	PulsarSubscription string `env:"PULSAR_SUBSCRIPTION" envDefault:"vlpd-reader"`
}

func NewReadConfig() (*ReadConfig, error) {
	var cfg ReadConfig

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
