package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sensority-labs/builder/internal/config"
)

type Config struct {
	Envs map[string]string
}

func GetBotConfig(cfg *config.Config, userName, botName string) (*Config, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/customers/get-bot-config/%s/%s", cfg.CoreURL, userName, botName), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode the response body into the Config struct
	var botConfig Config
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&botConfig.Envs); err != nil {
		return nil, err
	}

	return &botConfig, nil
}
