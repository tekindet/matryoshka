package orchestrator

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/tekindet/matryoshka/internal/domain"
)

type DockerOrchestrator struct {
	cli *client.Client
}

func NewDockerOrchestrator(cli *client.Client) *DockerOrchestrator {
	return &DockerOrchestrator{cli: cli}
}

func (d *DockerOrchestrator) CreateNetwork(ctx context.Context, name string) (string, error) {

	res, err := d.cli.NetworkCreate(ctx, name, network.CreateOptions{
		Driver:     "bridge",
		Attachable: true,
	})

	if err != nil {
		return "", fmt.Errorf("docker failed to create network %s: %w", name, err)
	}

	return res.ID, nil
}

func (d *DockerOrchestrator) DeployService(ctx context.Context, svc *domain.Service, networkName string) (string, error) {
	var image string
	var env []string
	var exposedPorts nat.PortSet

	switch svc.Type {
	case domain.ServiceTypePostgres:
		image = "postgres:15-alpine"
		env = []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
		}
		exposedPorts = nat.PortSet{"5435/tcp": struct{}{}}

	case domain.ServiceTypeRedis:
		image = "redis:7-alpine"
		exposedPorts = nat.PortSet{"6379/tcp": struct{}{}}

	case domain.ServiceTypeApp:
		return "", fmt.Errorf("service not implemented yet")

	case domain.ServiceTypeQueue:
		return "", fmt.Errorf("service not implemented yet")

	case domain.ServiceTypeWorker:
		return "", fmt.Errorf("service not implemented yet")

	default:
		return "", fmt.Errorf("service not implemented yet")
	}

	config := &container.Config{
		Image:        image,
		Env:          env,
		ExposedPorts: exposedPorts,
		Hostname:     svc.Name,
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: container.NetworkMode(networkName),
	}

	containerName := fmt.Sprintf("%s-%s", svc.Name, svc.ID[:6])

	res, err := d.cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		&network.NetworkingConfig{},
		&v1.Platform{Architecture: "x86_64"},
		containerName,
	)

	if err != nil {
		return "", fmt.Errorf("failed creating contianer configuration")
	}

	err = d.cli.ContainerStart(context.Background(), res.ID, container.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed starting container infrastructure : %w", err)
	}

	return res.ID, nil
}
