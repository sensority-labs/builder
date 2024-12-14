package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/sensority-labs/builder/internal/config"
)

type Client struct {
	cl *client.Client
}

func NewClient() (*Client, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &Client{cl: apiClient}, nil
}

func (c *Client) Close() error {
	if err := c.cl.Close(); err != nil {
		return err
	}
	return nil
}

type BotContainer struct {
	docker *Client
	ID     string
}

func NewBotContainer(id string) (*BotContainer, error) {
	cl, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &BotContainer{
		docker: cl,
		ID:     id,
	}, nil

}

func (bc *BotContainer) Close() error {
	if err := bc.docker.Close(); err != nil {
		return err
	}
	return nil
}

func (bc *BotContainer) Start() error {
	if err := bc.docker.cl.ContainerStart(context.Background(), bc.ID, container.StartOptions{}); err != nil {
		return err
	}
	return nil
}

func (bc *BotContainer) Stop() error {
	if err := bc.docker.cl.ContainerStop(context.Background(), bc.ID, container.StopOptions{}); err != nil {
		return err
	}
	return nil
}

func (bc *BotContainer) Remove() error {
	if err := bc.docker.cl.ContainerRemove(context.Background(), bc.ID, container.RemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func (c *Client) BuildImage(srcCodePath, imageName string) error {
	log.Default().Printf("Building image %s\n", imageName)

	dockerContext, err := getDockerContext(srcCodePath)
	if err != nil {
		return err
	}

	// Build the image
	buildResponse, err := c.cl.ImageBuild(context.Background(), dockerContext, types.ImageBuildOptions{
		Tags: []string{imageName},
	})
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Default().Println(fmt.Sprintf("Error: %+v", err))
		}
	}(buildResponse.Body)

	// Read the build output and print it to the console
	decoder := json.NewDecoder(buildResponse.Body)
	for {
		var message map[string]interface{}
		if err := decoder.Decode(&message); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if stream, ok := message["stream"]; ok {
			fmt.Print(stream)
		}
	}
	return nil
}

func (c *Client) RunContainer(cfg *config.Config, imageName, containerName, customerName, botName string) (string, error) {
	// Check if a container already exists
	containers, err := c.cl.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	if slices.ContainsFunc(containers, func(container types.Container) bool {
		return container.Names[0] == "/"+containerName
	}) {
		// Remove old container
		log.Default().Printf("Removing old container %s\n", containerName)
		if err := c.cl.ContainerRemove(context.Background(), containerName, container.RemoveOptions{Force: true}); err != nil {
			return "", err
		}
	}

	// Create a new container
	containerConfig := &container.Config{
		Image: imageName,
		Env: []string{
			"NATS_URL=" + cfg.Stream.NatsURL,
			"EVENTS_STREAM_NAME=" + cfg.Stream.EventStreamName,
			"FINDINGS_STREAM_NAME=" + cfg.Stream.FindingsStreamName,
			"SENTRY_DSN=" + cfg.Bot.SentryDSN,
			"CUSTOMER_NAME=" + customerName,
			"BOT_NAME=" + botName,
		},
	}
	// HostConfig is used to configure the container to be attached to the network
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyOnFailure,
		},
	}
	// Attach the container to the network
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			cfg.Stream.NetworkName: {},
		},
	}
	log.Default().Printf("Creating container %s from image: %s\n", containerName, imageName)
	cnt, err := c.cl.ContainerCreate(context.Background(), containerConfig, hostConfig, networkConfig, nil, containerName)
	if err != nil {
		return "", err
	}

	// Start the container
	log.Default().Printf("Starting container %s\n", containerName)
	if err := c.cl.ContainerStart(context.Background(), cnt.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return cnt.ID, nil
}

func getDockerContext(filePath string) (io.Reader, error) {
	ctx, err := archive.TarWithOptions(filePath, &archive.TarOptions{})
	if err != nil {
		return nil, err
	}
	return ctx, nil
}
