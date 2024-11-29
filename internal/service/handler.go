package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	"github.com/sensority-labs/builder/internal/config"
	"github.com/sensority-labs/builder/internal/docker"
)

func buildWatchman(cradlePath string, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// Extract the tar.gz file
		if err := extractWatchmanSourceCode(cradlePath, tempFile.Name()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Bot code extracted. Building docker image...")

		dc, err := docker.NewClient()
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(dockerClient *docker.Client) {
			if err := dockerClient.Close(); err != nil {
				log.Default().Println(fmt.Sprintf("Error: %+v", err))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(dc)

		// Build the cradle with a docker client
		imageName := botName + ":latest"
		if err := dc.BuildImage(cradlePath, imageName); err != nil {
			log.Default().Println(fmt.Sprintf("Error building image: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		containerName := botName // We'll define a proper container name later
		containerId, err := dc.RunContainer(cfg, imageName, containerName, customerName, botName)
		if err != nil {
			log.Default().Println(fmt.Sprintf("Error running container: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container started with ID: ", containerId)

		// Return the container ID
		if _, err := fmt.Fprintf(w, containerId); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func extractWatchmanSourceCode(cradlePath, tempFile string) error {
	var errb bytes.Buffer
	// Create a directory for the watchman
	watchmanPath := cradlePath + "/watchman"
	if _, err := os.Stat(watchmanPath); os.IsNotExist(err) {
		if err := os.Mkdir(watchmanPath, 0755); err != nil {
			return err
		}
	} else {
		if err := os.RemoveAll(watchmanPath); err != nil {
			return err
		}
		if err := os.Mkdir(watchmanPath, 0755); err != nil {
			return err
		}
	}

	// Extract the tar.gz file
	extractCmd := exec.Command("tar", "-xvzf", tempFile, "-C", watchmanPath)
	extractCmd.Stderr = &errb
	if _, err := extractCmd.Output(); err != nil {
		fmt.Println(errb.String())
		return fmt.Errorf("failed to extract the tar.gz file: %s", errb.String())
	}
	return nil
}
