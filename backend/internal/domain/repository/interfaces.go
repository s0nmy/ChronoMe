package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
)

// UserRepository abstracts persistence for the user model.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

// ProjectRepository handles CRUD for projects.
type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Project, error)
	GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Project, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

// EntryFilter captures query constraints.
type EntryFilter struct {
	From      *time.Time
	To        *time.Time
	ProjectID *uuid.UUID
	TagID     *uuid.UUID
}

// EntryRepository provides CRUD for entries.
type EntryRepository interface {
	Create(ctx context.Context, entry *entity.Entry) error
	ListByUser(ctx context.Context, userID uuid.UUID, filter EntryFilter) ([]entity.Entry, error)
	GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Entry, error)
	Update(ctx context.Context, entry *entity.Entry) error
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	ReplaceTags(ctx context.Context, entry *entity.Entry, tagIDs []uuid.UUID) error
}

// TagRepository is currently unused but reserved for future work.
type TagRepository interface {
	Create(ctx context.Context, tag *entity.Tag) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Tag, error)
	GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Tag, error)
	Update(ctx context.Context, tag *entity.Tag) error
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}
