package bot_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensority-labs/builder/internal/bot"
	"github.com/sensority-labs/builder/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetBotConfig_Success(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/customers/get-bot-config/testuser/testbot", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	botConfig, err := bot.GetConfig(cfg, userName, botName)

	assert.NoError(t, err)
	assert.NotNil(t, botConfig)
	assert.Equal(t, "value", botConfig.Envs["key"])
}

func TestGetBotConfig_Empty(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/customers/get-bot-config/testuser/testbot", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	botConfig, err := bot.GetConfig(cfg, userName, botName)

	assert.NoError(t, err)
	assert.NotNil(t, botConfig)
	assert.Equal(t, map[string]string{}, botConfig.Envs)
}

func TestGetBotConfig_HttpError(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	botConfig, err := bot.GetConfig(cfg, userName, botName)

	assert.Error(t, err)
	assert.Nil(t, botConfig)
}

func TestGetBotConfig_InvalidJson(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	botConfig, err := bot.GetConfig(cfg, userName, botName)

	assert.Error(t, err)
	assert.Nil(t, botConfig)
}

func TestGetBotConfig_UnexpectedStatusCode(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	botConfig, err := bot.GetConfig(cfg, userName, botName)

	assert.Error(t, err)
	assert.Nil(t, botConfig)
}

func TestGetBotConfig_RequestError(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"

	_, err := bot.GetConfig(cfg, userName, botName)

	assert.Error(t, err)
}

func TestUpdateBotID_Success(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"
	containerID := "container123"
	payloadBytes := []byte(`{"system_user_name":"testuser","bot_name":"testbot","container_id":"container123"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/customers/set-bot-container-id", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		reqBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, payloadBytes, reqBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	err := bot.UpdateID(cfg, userName, botName, containerID)

	assert.NoError(t, err)
}

func TestUpdateBotID_HttpError(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"
	containerID := "container123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	err := bot.UpdateID(cfg, userName, botName, containerID)

	assert.Error(t, err)
}

func TestUpdateBotID_UnexpectedStatusCode(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"
	containerID := "container123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg.CoreURL = server.URL
	err := bot.UpdateID(cfg, userName, botName, containerID)

	assert.Error(t, err)
}

func TestUpdateBotID_RequestError(t *testing.T) {
	cfg := &config.Config{CoreURL: "http://example.com"}
	userName := "testuser"
	botName := "testbot"
	containerID := "container123"

	err := bot.UpdateID(cfg, userName, botName, containerID)

	assert.Error(t, err)
}
