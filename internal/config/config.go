package config

import (
	"github.com/cristalhq/aconfig"
)

type Config struct {
	Debug          bool   `default:"false"`
	GithubToken    string `required:"true"`
	Port           string `default:"5005"`
	CoreURL        string `default:"http://core:8000"`
	ApiAccessToken string `required:"true"`
	NetworkName    string `default:"sensority-labs"`
	Bot            BotConfig
	Stream         StreamConfig
}

type StreamConfig struct {
	NatsURL            string `default:"nats://nats:4222"`
	EventStreamName    string `default:"ethereum_events"`
	FindingsStreamName string `default:"findings"`
}

type BotConfig struct {
	SentryDSN string
}

func GetConfig() (*Config, error) {
	var cfg Config
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		AllowUnknownFlags: true,
		SkipFlags:         true,
	})

	if err := loader.Load(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
