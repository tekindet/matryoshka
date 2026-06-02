package orchestrator_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/tekindet/matryoshka/internal/domain"
	"github.com/tekindet/matryoshka/internal/orchestrator"
)

func TestOrchestrator_CreateNetwork(t *testing.T) {

	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		t.Fatalf("could not create a docker client %v", err)
	}
	defer cli.Close()

	orch := orchestrator.NewDockerOrchestrator(cli)
	ctx := context.Background()
	networkName := "test-paas-network-integration"

	defer cli.NetworkRemove(ctx, networkName)

	netID, err := orch.CreateNetwork(ctx, networkName)
	if err != nil {
		t.Fatalf("expected network creation to succeed, got : %v", err)
	}

	if netID == "" {
		t.Error("expected a valid network ID string, got empty string")
	}

	netResource, err := cli.NetworkInspect(ctx, netID, network.InspectOptions{})
	if err != nil {
		t.Fatalf("failed to inspect created network: %v", err)
	}

	if netResource.Name != networkName {
		t.Errorf("wanted network name to be %s got %s", networkName, netResource.Name)
	}
}

func TestOrchestrator_DeployService_Redis(t *testing.T) {

	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		t.Fatalf("could not create a docker client %v", err)
	}
	defer cli.Close()

	orch := orchestrator.NewDockerOrchestrator(cli)
	ctx := context.Background()

	networkName := "test-project-isolated-net"

	_, err = orch.CreateNetwork(ctx, networkName)
	if err != nil {
		t.Skip("failed to create network, might already exist")
	}

	defer cli.NetworkRemove(ctx, networkName)

	mockSvc := &domain.Service{
		ID:        "abc-123-xyz-789-longstring",
		ProjectID: "proj-123",
		Name:      "test-redis-cache",
		Type:      domain.ServiceTypeRedis,
	}

	containerID, err := orch.DeployService(ctx, mockSvc, networkName)
	if err != nil {
		t.Fatalf("expected service to deploy onto docker cleanly, got: %v", err)
	}

	if containerID == "" {
		t.Error("expected a functional system container ID handle")
	}

	defer cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})

	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		t.Fatalf("failed to inspect launched container infrastructure: %v", err)
	}

	if !inspect.State.Running {
		t.Errorf("expected container to be state:RUNNING, instead found status: %s", inspect.State.Status)
	}

}
