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

// TagUsecase はタグの CRUD を扱う。
type TagUsecase struct {
	tags repository.TagRepository
	cfg  provider.AppConfig
}

func NewTagUsecase(tags repository.TagRepository, cfg provider.AppConfig) *TagUsecase {
	return &TagUsecase{tags: tags, cfg: cfg}
}

func (u *TagUsecase) List(ctx context.Context, userID uuid.UUID) ([]entity.Tag, error) {
	return u.tags.ListByUser(ctx, userID)
}

func (u *TagUsecase) Create(ctx context.Context, userID uuid.UUID, input dto.TagCreateRequest) (*entity.Tag, error) {
	data, err := input.Normalize(u.cfg.DefaultProjectColor())
	if err != nil {
		return nil, err
	}
	tag := &entity.Tag{
		ID:     uuid.New(),
		UserID: userID,
		Name:   data.Name,
		Color:  data.Color,
	}
	if err := tag.Validate(); err != nil {
		return nil, err
	}
	if err := u.tags.Create(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (u *TagUsecase) Update(ctx context.Context, userID uuid.UUID, id uuid.UUID, input dto.TagUpdateRequest) (*entity.Tag, error) {
	updates, err := input.Normalize()
	if err != nil {
		return nil, err
	}
	tag, err := u.tags.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if updates.Name != nil {
		tag.Name = *updates.Name
	}
	if updates.Color != nil {
		tag.Color = *updates.Color
	}
	if err := tag.Validate(); err != nil {
		return nil, err
	}
	if err := u.tags.Update(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (u *TagUsecase) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("id is required")
	}
	return u.tags.Delete(ctx, userID, id)
}
