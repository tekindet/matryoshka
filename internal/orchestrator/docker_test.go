package orchestrator_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
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
