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

	"github.com/sensority-labs/builder/internal/docker"
)

func buildWatchman(cradlePath, networkName, natsURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repoName")

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

		log.Default().Println("Uploaded File: ", handler.Filename)
		log.Default().Println("File Size: ", handler.Size)

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

		dc, err := docker.NewClient()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(dockerClient *docker.Client) {
			if err := dockerClient.Close(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(dc)

		// Build the cradle with a docker client
		imageName := repoName + ":latest"
		if err := dc.BuildImage(cradlePath, imageName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		containerName := repoName // We'll define a proper container name later
		containerId, err := dc.RunContainer(imageName, containerName, networkName, natsURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container started with ID: ", containerId)

		// Return the container ID
		if _, err := fmt.Fprintf(w, containerId); err != nil {
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

func getCradle(cradlePath string) error {
	clearCmd := exec.Command("rm", "-rf", cradlePath)
	if err := clearCmd.Run(); err != nil {
		return err
	}
	log.Printf("Cloning cradle to the path: %s\n", cradlePath)
	cmd := exec.Command("git", "clone", "git@github.com:sensority-labs/cradle-ts.git", cradlePath)
	if _, err := cmd.Output(); err != nil {
		fmt.Println(cmd.Stderr)
		return err
	}
	return nil
}

func Run() error {
	tmpDir := os.TempDir()
	cradlePath := tmpDir + "cradle-ts"
	if err := getCradle(cradlePath); err != nil {
		return err
	}

	networkName := os.Getenv("NETWORK_NAME")
	if networkName == "" {
		networkName = "sensority-labs"
	}
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	// Setup server
	http.HandleFunc("/build/{repoName}", buildWatchman(cradlePath, networkName, natsURL))

	// Start the server
	log.Default().Println("Server started at :8000")
	return http.ListenAndServe(":8000", nil)
}
