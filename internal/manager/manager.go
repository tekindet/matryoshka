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

func (m *PaaSManager) CreateService(ctx context.Context, projectID, name, svcType string) (*domain.Service, error) {
	proj, err := m.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign service : %v", err)
	}

	serviceID := uuid.New().String()

	svc := &domain.Service{
		ID:        serviceID,
		ProjectID: proj.ID,
		Name:      name,
		Type:      svcType,
		Status:    domain.ServiceStatusPending,
	}

	if err := m.store.CreateService(ctx, svc); err != nil {
		return nil, fmt.Errorf("failed to create service : %v", err)
	}

	networkName := fmt.Sprintf("net-%s", proj.ID)
	_, err = m.orch.CreateNetwork(ctx, networkName)
	if err != nil {
		svc.Status = domain.ServiceStatusFailed

		// todo : need to update the state here
		return nil, fmt.Errorf("failed to deploy infrastructure : %v", err)
	}

	svc.Status = domain.ServiceStatusRunning
	// todo : same here, need to update the state in the store

	return svc, nil
}
