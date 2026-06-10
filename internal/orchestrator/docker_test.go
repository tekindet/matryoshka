package orchestrator_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/tekindet/matryoshka/internal/domain"
	"github.com/tekindet/matryoshka/internal/orchestrator"
)

func TestMain(m *testing.M) {
	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		os.Exit(1)
	}
	defer cli.Close()

	images := []string{
		"redis:7-alpine",
		"ealen/echo-server:latest",
	}
	ctx := context.Background()
	for _, img := range images {
		reader, err := cli.ImagePull(ctx, img, image.PullOptions{})
		if err != nil {
			os.Exit(1)
		}
		io.Copy(io.Discard, reader)
		reader.Close()
	}

	os.Exit(m.Run())
}

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

func TestOrchestrator_InterServiceNetworking(t *testing.T) {

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

	networkName := "net-shared-project-test"

	_, err = orch.CreateNetwork(ctx, networkName)
	if err != nil {
		t.Skip("failed to create network, might already exist")
	}

	defer cli.NetworkRemove(ctx, networkName)

	redisSvc := &domain.Service{
		ID:        "redis-111",
		ProjectID: "proj-shared",
		Type:      domain.ServiceTypeRedis,
		Name:      "cache-service",
	}

	appSvc := &domain.Service{
		ID:        "app-222",
		ProjectID: "proj-shared",
		Type:      domain.ServiceTypeApp,
		Name:      "web-app-service",
	}

	redisID, err := orch.DeployService(ctx, redisSvc, networkName)
	if err != nil {
		t.Fatalf("failed to deploy redis service %v", err)
	}

	defer cli.ContainerRemove(ctx, redisID, container.RemoveOptions{Force: true})

	appID, err := orch.DeployService(ctx, appSvc, networkName)
	if err != nil {
		t.Fatalf("failed to deploy web-app-service %v", err)
	}

	defer cli.ContainerRemove(ctx, appID, container.RemoveOptions{Force: true})

	execConfig := container.ExecOptions{
		Cmd:          []string{"nc", "-z", "-v", "cache-service", "6379"},
		AttachStderr: true,
		AttachStdout: true,
	}

	execCreateRes, err := cli.ContainerExecCreate(ctx, appID, execConfig)
	if err != nil {
		t.Fatalf("failed to create exec config : %v", err)
	}

	err = cli.ContainerExecStart(ctx, execCreateRes.ID, container.ExecStartOptions{})
	if err != nil {
		t.Fatalf("Failed to start exec: %v", err)
	}

	for {
		inspectRes, err := cli.ContainerExecInspect(ctx, execCreateRes.ID)
		if err != nil {
			t.Fatalf("Failed to inspect exec outcome: %v", err)
		}
		if !inspectRes.Running {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	inspectRes, err := cli.ContainerExecInspect(ctx, execCreateRes.ID)
	if err != nil {
		t.Fatalf("Failed to inspect exec outcome: %v", err)
	}

	if inspectRes.ExitCode != 0 {
		t.Errorf("Networking isolation test failed. Web app container could not resolve or reach 'cache-service:6379'. Exit code: %d", inspectRes.ExitCode)
	}

}

func TestOrchestrator_ExposeServiceToPublic(t *testing.T) {

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

	networkName := "net-ingress-test"

	_, err = orch.CreateNetwork(ctx, networkName)
	if err != nil {
		t.Skip("failed to create network, might already exist")
	}

	defer cli.NetworkRemove(ctx, networkName)

	publicSvc := &domain.Service{
		ID:        "app-ingress-123",
		ProjectID: "project-ingress",
		Name:      "public-web-app",
		Type:      domain.ServiceTypeApp,
	}

	containerID, err := orch.DeployService(ctx, publicSvc, networkName)
	if err != nil {
		t.Fatalf("failed to deploy service : %v", err)
	}

	defer cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})

	if publicSvc.ExternalPort == 0 {
		t.Fatal("Expected Orchestrator to assign an ExternalPort to the app service, got 0")
	}

	targetURL := "http://localhost:" + strconv.Itoa(publicSvc.ExternalPort) + "/"

	time.Sleep(10 * time.Second)
	resp, err := http.Get(targetURL)

	if err != nil {
		t.Fatalf("Failed to reach exposed application on host port %d after retries: %v", publicSvc.ExternalPort, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200 from exposed service, got %d", resp.StatusCode)
	}

}
