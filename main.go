package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
)

func buildWatchman(cradlePath, networkName, natsURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var errb bytes.Buffer
		repoName := r.PathValue("repoName")

		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
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

		// Create a temporary file within our temp-images directory that follows
		// a particular naming pattern
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

		// read all the contents of our uploaded file into a
		// byte array
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// write this byte array to our temporary file
		if _, err := tempFile.Write(fileBytes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Extract the tar.gz file
		// Create a directory for the watchman
		watchmanPath := cradlePath + "/watchman"
		if _, err := os.Stat(watchmanPath); os.IsNotExist(err) {
			if err := os.Mkdir(watchmanPath, 0755); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			if err := os.RemoveAll(watchmanPath); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := os.Mkdir(watchmanPath, 0755); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Extract the tar.gz file
		extractCmd := exec.Command("tar", "-xvzf", tempFile.Name(), "-C", watchmanPath)
		extractCmd.Stderr = &errb
		if _, err := extractCmd.Output(); err != nil {
			fmt.Println(errb.String())
			http.Error(w, errb.String(), http.StatusInternalServerError)
			return
		}

		// Build the cradle with watchman
		imageName := repoName + ":latest"
		containerName := repoName
		buildCmd := exec.Command("docker", "build", "-t", imageName, ".")
		buildCmd.Stderr = &errb
		buildCmd.Dir = cradlePath
		if _, err := buildCmd.Output(); err != nil {
			fmt.Println(errb.String())
			http.Error(w, errb.String(), http.StatusInternalServerError)
			return
		}

		// Check that the container is exist
		checkCmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
		checkCmd.Stderr = &errb
		output, err := checkCmd.Output()
		if err != nil {
			fmt.Println(errb.String())
			http.Error(w, errb.String(), http.StatusInternalServerError)
			return
		}
		if bytes.Contains(output, []byte(containerName)) {
			// Remove old and run the new container
			stopCmd := exec.Command("docker", "stop", containerName)
			stopCmd.Stderr = &errb
			if _, err := stopCmd.Output(); err != nil {
				fmt.Println(errb.String())
			}
			removeCmd := exec.Command("docker", "rm", containerName)
			removeCmd.Stderr = &errb
			if _, err := removeCmd.Output(); err != nil {
				fmt.Println(errb.String())
			}
		}

		runCmd := exec.Command("docker", "run", "-d", "--network", networkName, "-e", "NATS_URL="+natsURL, "--name", containerName, imageName)
		runCmd.Stderr = &errb
		output, err = runCmd.Output()
		if err != nil {
			fmt.Println(errb.String())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Default().Println("Container started with ID: ", string(output))

		if _, err := fmt.Fprintf(w, string(output)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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

func main() {
	tmpDir := os.TempDir()
	cradlePath := tmpDir + "cradle-ts"
	if err := getCradle(cradlePath); err != nil {
		log.Fatal(err)
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
	log.Default().Println("Server started at :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
