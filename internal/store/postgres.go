package store

import (
	"context"

	"github.com/tekindet/matryoshka/internal/domain"
	"gorm.io/gorm"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(db *gorm.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateProject(ctx context.Context, project *domain.Project) error {
	return s.db.WithContext(ctx).Create(project).Error
}

func (s *PostgresStore) GetProject(ctx context.Context, projectID string) (*domain.Project, error) {
	var project domain.Project

	err := s.db.WithContext(ctx).First(&project, "id = ?", projectID).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *PostgresStore) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	var projects []*domain.Project
	err := s.db.WithContext(ctx).Find(&projects).Error
	return projects, err
}

func (s *PostgresStore) CreateService(ctx context.Context, svc *domain.Service) error {
	return s.db.WithContext(ctx).Create(svc).Error
}

func (s *PostgresStore) ListServices(ctx context.Context, projectID string) ([]*domain.Service, error) {
	var services []*domain.Service
	err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&services).Error
	return services, err
}
