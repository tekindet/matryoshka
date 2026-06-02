package orchestrator

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
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
	return "not-yet-implemented", nil
}
