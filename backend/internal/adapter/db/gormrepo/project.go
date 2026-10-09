package gormrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

// ProjectRepository は GORM で repository.ProjectRepository を実装する。
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

func (r *ProjectRepository) Update(ctx context.Context, project *entity.Project, columns []string) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&entity.Project{}).Where("id = ? AND user_id = ? AND version = ?", project.ID, project.UserID, project.Version).
			Select(append(append([]string{}, columns...), "version", "updated_at")).Omit("Tags").Updates(updateValues(map[string]any{"name": project.Name, "description": project.Description, "color": project.Color, "is_archived": project.IsArchived}, columns, now))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return repository.ErrConflict
		}
		return nil
	})
	if err == nil {
		project.Version++
		project.UpdatedAt = now
	}
	return err
}

func (r *ProjectRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Project{}).Error
}
