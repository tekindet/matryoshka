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

	err := s.db.WithContext(ctx).Find(&project, "id = ?", projectID).Error
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (s *PostgresStore) CreateService(ctx context.Context, svc *domain.Service) error {
	return s.db.WithContext(ctx).Create(svc).Error
}
