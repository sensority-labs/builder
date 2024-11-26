package service

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/sensority-labs/builder/internal/config"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

const cradleRepoURL = "https://github.com/sensority-labs/cradle-ts.git"

func getCradle(cradlePath, ghToken string) error {
	clearCmd := exec.Command("rm", "-rf", cradlePath)
	if err := clearCmd.Run(); err != nil {
		return err
	}
	log.Printf("Cloning cradle to the path: %s\n", cradlePath)
	auth := &githttp.BasicAuth{
		Username: "username", // Can be anything except an empty string
		Password: ghToken,
	}

	_, err := git.PlainClone(cradlePath, false, &git.CloneOptions{
		URL:      cradleRepoURL,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func Run(cfg *config.Config) error {
	tmpDir := os.TempDir()
	cradlePath := path.Join(tmpDir, "cradle-ts")
	if err := getCradle(cradlePath, cfg.GithubToken); err != nil {
		return err
	}

	// Setup server
	http.HandleFunc("/build/{repoName}", buildWatchman(cradlePath, cfg.NetworkName, cfg.NatsURL))

	// Start the server
	log.Default().Println("Server started at :" + cfg.Port)
	return http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil)
}
