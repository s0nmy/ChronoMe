package gormrepo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"chronome/internal/domain/entity"
)

// ProjectRepository implements repository.ProjectRepository via GORM.
type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Project, error) {
	var res []entity.Project
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Project, error) {
	var project entity.Project
	err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Project{}).Error
}
