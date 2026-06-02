package manager_test

import (
	"context"

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
