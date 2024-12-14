package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/sensority-labs/builder/internal/config"
	"github.com/sensority-labs/builder/internal/docker"
)

func startBot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")

		bc, err := docker.GetBotContainer(containerId)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(bc *docker.BotContainer) {
			if err := bc.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(bc)

		if err := bc.Start(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container started with ID: ", containerId)
	}
}

func stopBot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")

		bc, err := docker.GetBotContainer(containerId)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(bc *docker.BotContainer) {
			if err := bc.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(bc)

		if err := bc.Stop(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container stopped with ID: ", containerId)
	}
}

func removeBot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")

		bc, err := docker.GetBotContainer(containerId)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(bc *docker.BotContainer) {
			if err := bc.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(bc)

		if err := bc.Remove(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container removed with ID: ", containerId)
	}
}

func recreateBot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")

		bc, err := docker.GetBotContainer(containerId)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(bc *docker.BotContainer) {
			if err := bc.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(bc)

		if err := bc.Recreate(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Printf("Container recreated\n oldID: %s\n New ID: %s", containerId, bc.ID)

		response := struct {
			ContainerID string `json:"containerId"`
		}{
			ContainerID: bc.ID,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func botStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")

		bc, err := docker.GetBotContainer(containerId)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(bc *docker.BotContainer) {
			if err := bc.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(bc)

		status, err := bc.Status()
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		statusResponse := struct {
			Status string `json:"status"`
		}{
			Status: status,
		}

		if err := json.NewEncoder(w).Encode(statusResponse); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func makeBot(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpDir := os.TempDir()
		cradlePath := path.Join(tmpDir, "cradle-ts")

		if err := getCradle(cradlePath, cfg.GithubToken); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		botName := r.PathValue("botName")
		customerName := r.PathValue("customerName")

		// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(file)

		log.Default().Println("Received new bot code:")
		log.Default().Println("Received file: ", handler.Filename)
		log.Default().Println("File size: ", handler.Size)
		log.Default().Println("Extracting...")

		// Create a temporary file within our temp-images directory that follows a particular pattern.
		tempFile, err := os.CreateTemp(cradlePath, "*.tar.gz")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(tempFile *os.File) {
			err := tempFile.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(tempFile)

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := tempFile.Write(fileBytes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Extract the tar.gz file to the cradle directory
		if err := extractBotSourceCode(cradlePath, tempFile.Name()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Bot code extracted. Building docker image...")

		bc, err := docker.NewBotContainer(cfg, botName, customerName)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(bc *docker.BotContainer) {
			if err := bc.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(bc)

		log.Default().Println("Building the bot image...")
		if err := bc.Build(cradlePath); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Bot image built. Creating container...")
		if err := bc.Create(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Printf("Container created with ID: %s\nStarting...", bc.ID)
		if err := bc.Start(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container started")

		// Return the container ID
		if _, err := fmt.Fprintf(w, bc.ID); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func extractBotSourceCode(cradlePath, tempFile string) error {
	var errb bytes.Buffer
	// Create a directory for the bot
	botPath := cradlePath + "/bot"
	if _, err := os.Stat(botPath); os.IsNotExist(err) {
		if err := os.Mkdir(botPath, 0755); err != nil {
			return err
		}
	} else {
		if err := os.RemoveAll(botPath); err != nil {
			return err
		}
		if err := os.Mkdir(botPath, 0755); err != nil {
			return err
		}
	}

	// Extract the tar.gz file
	extractCmd := exec.Command("tar", "-xvzf", tempFile, "-C", botPath)
	extractCmd.Stderr = &errb
	if _, err := extractCmd.Output(); err != nil {
		fmt.Println(errb.String())
		return fmt.Errorf("failed to extract the tar.gz file: %s", errb.String())
	}
	return nil
}
