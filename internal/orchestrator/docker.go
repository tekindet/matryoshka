package orchestrator

import (
	"context"
	"fmt"
	"net"
	"strconv"

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

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (d *DockerOrchestrator) DeployService(ctx context.Context, svc *domain.Service, networkName string) (string, error) {
	var image string
	var env []string
	var exposedPorts nat.PortSet
	var portBindings nat.PortMap

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
		image = "ealen/echo-server:latest"
		env = []string{"PORT=8080"}

		exposedPorts = nat.PortSet{"8080/tcp": struct{}{}}

		hostPort, err := getFreePort()
		if err != nil {
			return "", fmt.Errorf("failed to allocate a free host port: %w", err)
		}

		portBindings = nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(hostPort),
				},
			},
		}
		svc.ExternalPort = hostPort

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
		AutoRemove:   true,
		NetworkMode:  container.NetworkMode(networkName),
		PortBindings: portBindings,
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {
				Aliases: []string{svc.Name},
			},
		},
	}

	containerName := fmt.Sprintf("%s-%s", svc.Name, svc.ID[:6])

	res, err := d.cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkingConfig,
		&v1.Platform{Architecture: "x86_64"},
		containerName,
	)

	if err != nil {
		return "", fmt.Errorf("failed creating container configuration: %w", err)
	}

	err = d.cli.ContainerStart(context.Background(), res.ID, container.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("failed starting container infrastructure : %w", err)
	}

	return res.ID, nil
}
