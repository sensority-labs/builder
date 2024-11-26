package config

import (
	"fmt"
	"os"
)

type Config struct {
	GithubToken string
	NetworkName string
	NatsURL     string
	Port        string
}

func GetConfig() (*Config, error) {
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

	return &Config{
		GithubToken: os.Getenv("GITHUB_TOKEN"),
		NetworkName: networkName,
		NatsURL:     natsURL,
		Port:        port,
	}, nil
}
