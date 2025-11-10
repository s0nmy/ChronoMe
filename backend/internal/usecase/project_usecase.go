package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
	"chronome/internal/usecase/dto"
	"chronome/internal/usecase/provider"
)

// ProjectUsecase holds business logic for project CRUD.
type ProjectUsecase struct {
	projects repository.ProjectRepository
	cfg      provider.AppConfig
}

func NewProjectUsecase(projects repository.ProjectRepository, cfg provider.AppConfig) *ProjectUsecase {
	return &ProjectUsecase{projects: projects, cfg: cfg}
}

func (u *ProjectUsecase) Create(ctx context.Context, userID uuid.UUID, input dto.ProjectCreateRequest) (*entity.Project, error) {
	data, err := input.Normalize(u.cfg.DefaultProjectColor())
	if err != nil {
		return nil, err
	}
	project := &entity.Project{
		ID:     uuid.New(),
		UserID: userID,
		Name:   data.Name,
		Color:  data.Color,
	}
	if err := project.Validate(); err != nil {
		return nil, err
	}
	if err := u.projects.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (u *ProjectUsecase) List(ctx context.Context, userID uuid.UUID) ([]entity.Project, error) {
	return u.projects.ListByUser(ctx, userID)
}

func (u *ProjectUsecase) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input dto.ProjectUpdateRequest) (*entity.Project, error) {
	data, err := input.Normalize()
	if err != nil {
		return nil, err
	}
	project, err := u.projects.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if data.Name != nil {
		project.Name = *data.Name
	}
	if data.Color != nil {
		project.Color = *data.Color
	}
	if data.IsArchived != nil {
		project.IsArchived = *data.IsArchived
	}
	if err := project.Validate(); err != nil {
		return nil, err
	}
	if err := u.projects.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (u *ProjectUsecase) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("id is required")
	}
	return u.projects.Delete(ctx, userID, id)
}
