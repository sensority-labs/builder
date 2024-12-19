package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sensority-labs/builder/internal/config"
)

type Config struct {
	Envs map[string]string
}

func GetConfig(cfg *config.Config, userName, botName string) (*Config, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/customers/get-bot-config/%s/%s", cfg.CoreURL, userName, botName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", cfg.ApiAccessToken)
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

func UpdateID(cfg *config.Config, userName, botName, containerID string) error {
	payload := struct {
		UserName    string `json:"system_user_name"`
		BotName     string `json:"bot_name"`
		ContainerID string `json:"container_id"`
	}{
		UserName:    userName,
		BotName:     botName,
		ContainerID: containerID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/customers/set-bot-container-id", cfg.CoreURL), bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", cfg.ApiAccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
