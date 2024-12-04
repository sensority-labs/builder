package config

import (
	"fmt"
	"os"
)

type Config struct {
	Debug       bool
	GithubToken string
	Port        string
	Bot         BotConfig
	Stream      StreamConfig
}

type StreamConfig struct {
	NetworkName        string
	NatsURL            string
	EventStreamName    string
	FindingsStreamName string
}

type BotConfig struct {
	SentryDSN string
}

func GetConfig() (*Config, error) {
	debug := os.Getenv("DEBUG") == "true"

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}
	networkName := os.Getenv("NETWORK_NAME")
	if networkName == "" {
		networkName = "sensority-labs"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "5005"
	}
	botsSentryDSN := os.Getenv("BOTS_SENTRY_DSN")

	eventStreamName := os.Getenv("EVENT_STREAM_NAME")
	if eventStreamName == "" {
		eventStreamName = "ethereum_events"
	}
	findingsStreamName := os.Getenv("FINDINGS_STREAM_NAME")
	if findingsStreamName == "" {
		findingsStreamName = "findings"
	}

	return &Config{
		Debug:       debug,
		GithubToken: os.Getenv("GITHUB_TOKEN"),
		Port:        port,
		Bot: BotConfig{
			SentryDSN: botsSentryDSN,
		},
		Stream: StreamConfig{
			NetworkName:        networkName,
			NatsURL:            natsURL,
			EventStreamName:    eventStreamName,
			FindingsStreamName: findingsStreamName,
		},
	}, nil
}
