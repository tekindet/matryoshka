package manager

import (
	"context"
	"errors"

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
