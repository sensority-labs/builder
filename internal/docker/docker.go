package docker

import (
	"context"
	"io"
	"log"
	"slices"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
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

func (c *Client) BuildImage(srcCodePath, imageName string) error {
	dockerContext, err := getDockerContext(srcCodePath)
	if err != nil {
		return err
	}
	log.Default().Printf("Building image %s\n", imageName)
	if _, err := c.cl.ImageBuild(context.Background(), dockerContext, types.ImageBuildOptions{Tags: []string{imageName}}); err != nil {
		return err
	}
	return nil

}

func (c *Client) RunContainer(imageName, containerName, networkName, natsURL string) (string, error) {
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
		Env:   []string{"NATS_URL=" + natsURL},
	}
	// Attach the container to the network
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}
	log.Default().Printf("Creating container %s from image: %s\n", containerName, imageName)
	cnt, err := c.cl.ContainerCreate(context.Background(), containerConfig, nil, networkConfig, nil, containerName)
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
