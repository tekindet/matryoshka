package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/tekindet/matryoshka/internal/domain"
)

var ErrProjectNotFound = errors.New("project not found")

type Manager interface {
	CreateProject(ctx context.Context, name, description string) (*domain.Project, error)
	CreateService(ctx context.Context, projectID, svcType string) (*domain.Service, error)
}

type Store interface {
	CreateProject(ctx context.Context, proj *domain.Project) error
	GetProject(ctx context.Context, id string) (*domain.Project, error)
	CreateService(ctx context.Context, svc *domain.Service) error
}

type Orchestrator interface {
	CreateNetwork(ctx context.Context, name string) (string, error)
	DeployService(ctx context.Context, svc *domain.Service, networkName string) (string, error)
}

type PaaSManager struct {
	store Store
	orch  Orchestrator
}

func NewPaaSManager(store Store, orch Orchestrator) *PaaSManager {
	return &PaaSManager{store: store, orch: orch}
}

func (m *PaaSManager) CreateProject(ctx context.Context, name, description string) (*domain.Project, error) {
	projectID := uuid.New().String()

	project := &domain.Project{
		ID:          projectID,
		Name:        name,
		Description: description,
	}

	if err := m.store.CreateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to save new project %v", err)
	}

	networkName := fmt.Sprintf("net-%s", project.ID)
	_, err := m.orch.CreateNetwork(ctx, networkName)
	if err != nil {
		return nil, fmt.Errorf("failed to provision project network %v", err)
	}

	return project, nil
}
