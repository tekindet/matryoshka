package manager_test

import (
	"context"
	"testing"

	"github.com/tekindet/matryoshka/internal/domain"
	"github.com/tekindet/matryoshka/internal/manager"
)

type MockStore struct {
	projects map[string]*domain.Project
}

func (m *MockStore) CreateProject(ctx context.Context, proj *domain.Project) error {
	m.projects[proj.ID] = proj
	return nil
}

func (m *MockStore) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	proj, exists := m.projects[id]
	if !exists {
		return nil, manager.ErrProjectNotFound
	}
	return proj, nil
}

func (m *MockStore) CreateService(ctx context.Context, svc *domain.Service) error {
	return nil
}

type MockOrchestrator struct {
	CreatedNetworks []string
}

func (m *MockOrchestrator) CreateNetwork(ctx context.Context, name string) (string, error) {
	m.CreatedNetworks = append(m.CreatedNetworks, name)
	return "mock-net-id", nil
}

func (m *MockOrchestrator) DeployService(ctx context.Context, svc *domain.Service, networkName string) (string, error) {
	return "mock-container-id", nil
}

func TestCreateProject_Success(t *testing.T) {
	store := &MockStore{projects: make(map[string]*domain.Project)}
	orch := &MockOrchestrator{}

	m := manager.NewPaaSManager(store, orch)

	projectName := "e-commerce"
	projectDesc := "Production deployment for online store"

	proj, err := m.CreateProject(context.Background(), projectName, projectDesc)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if proj.Name != projectName {
		t.Errorf("expected name %s, got %s", projectName, proj.Name)
	}

	if proj.Description != projectDesc {
		t.Errorf("expected description %s, got %s", projectDesc, proj.Description)
	}

	if proj.ID == "" {
		t.Error("expected a generated project ID, got empty string")
	}

	expectedNetworkName := "net-" + proj.ID
	if len(orch.CreatedNetworks) != 1 || orch.CreatedNetworks[0] != expectedNetworkName {
		t.Errorf("expected network %s to be provisioned, got %v", expectedNetworkName, orch.CreatedNetworks)
	}
}

func TestMannager_CreateService(t *testing.T) {
	store := &MockStore{projects: make(map[string]*domain.Project)}
	orch := &MockOrchestrator{}
	m := manager.NewPaaSManager(store, orch)

	ctx := context.Background()

	proj, err := m.CreateProject(ctx, "data-pipeline", "Processes large streams")
	if err != nil {
		t.Fatalf("wanted project to be created but got : %v", err)
	}

	svc, err := m.CreateService(ctx, proj.ID, "cache-layer", domain.ServiceTypeRedis)
	if err != nil {
		t.Fatalf("wanted service to be provisioned but got : %v", err)
	}

	if svc.Name != "cache-layer" {
		t.Errorf("wanted svc name to be %s got %s", "cache-layer", svc.Name)
	}

	if svc.Type != domain.ServiceTypeRedis {
		t.Errorf("wanted svc type to be %s got %s", domain.ServiceTypeRedis, svc.Type)
	}

	if svc.Status != domain.ServiceStatusRunning {
		t.Errorf("wanted svc status to be %s got %s", domain.ServiceStatusRunning, svc.Status)
	}

}
