package fakes

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

// FakeUserRepository はテスト用に repository.UserRepository を実装する。
type FakeUserRepository struct {
	CreateFn     func(context.Context, *entity.User) error
	GetByEmailFn func(context.Context, string) (*entity.User, error)
	GetByIDFn    func(context.Context, uuid.UUID) (*entity.User, error)
}

func (f *FakeUserRepository) Create(ctx context.Context, user *entity.User) error {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, user)
	}
	return nil
}

func (f *FakeUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if f.GetByEmailFn != nil {
		return f.GetByEmailFn(ctx, email)
	}
	return nil, errors.New("GetByEmail not implemented")
}

func (f *FakeUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if f.GetByIDFn != nil {
		return f.GetByIDFn(ctx, id)
	}
	return nil, errors.New("GetByID not implemented")
}

// FakeProjectRepository はテスト用に repository.ProjectRepository を実装する。
type FakeProjectRepository struct {
	CreateFn  func(context.Context, *entity.Project) error
	ListFn    func(context.Context, uuid.UUID) ([]entity.Project, error)
	GetByIDFn func(context.Context, uuid.UUID, uuid.UUID) (*entity.Project, error)
	UpdateFn  func(context.Context, *entity.Project) error
	DeleteFn  func(context.Context, uuid.UUID, uuid.UUID) error
}

func (f *FakeProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, project)
	}
	return nil
}

func (f *FakeProjectRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Project, error) {
	if f.ListFn != nil {
		return f.ListFn(ctx, userID)
	}
	return nil, nil
}

func (f *FakeProjectRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Project, error) {
	if f.GetByIDFn != nil {
		return f.GetByIDFn(ctx, userID, id)
	}
	return nil, errors.New("GetByID not implemented")
}

func (f *FakeProjectRepository) Update(ctx context.Context, project *entity.Project) error {
	if f.UpdateFn != nil {
		return f.UpdateFn(ctx, project)
	}
	return nil
}

func (f *FakeProjectRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if f.DeleteFn != nil {
		return f.DeleteFn(ctx, userID, id)
	}
	return nil
}

// FakeEntryRepository はテスト用に repository.EntryRepository を実装する。
type FakeEntryRepository struct {
	CreateFn      func(context.Context, *entity.Entry) error
	ListFn        func(context.Context, uuid.UUID, repository.EntryFilter) ([]entity.Entry, error)
	GetByIDFn     func(context.Context, uuid.UUID, uuid.UUID) (*entity.Entry, error)
	UpdateFn      func(context.Context, *entity.Entry) error
	DeleteFn      func(context.Context, uuid.UUID, uuid.UUID) error
	ReplaceTagsFn func(context.Context, *entity.Entry, []uuid.UUID) error
}

func (f *FakeEntryRepository) Create(ctx context.Context, entry *entity.Entry) error {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, entry)
	}
	return nil
}

func (f *FakeEntryRepository) ListByUser(ctx context.Context, userID uuid.UUID, filter repository.EntryFilter) ([]entity.Entry, error) {
	if f.ListFn != nil {
		return f.ListFn(ctx, userID, filter)
	}
	return nil, nil
}

func (f *FakeEntryRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Entry, error) {
	if f.GetByIDFn != nil {
		return f.GetByIDFn(ctx, userID, id)
	}
	return nil, errors.New("GetByID not implemented")
}

func (f *FakeEntryRepository) Update(ctx context.Context, entry *entity.Entry) error {
	if f.UpdateFn != nil {
		return f.UpdateFn(ctx, entry)
	}
	return nil
}

func (f *FakeEntryRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if f.DeleteFn != nil {
		return f.DeleteFn(ctx, userID, id)
	}
	return nil
}

func (f *FakeEntryRepository) ReplaceTags(ctx context.Context, entry *entity.Entry, tagIDs []uuid.UUID) error {
	if f.ReplaceTagsFn != nil {
		return f.ReplaceTagsFn(ctx, entry, tagIDs)
	}
	return nil
}

// FakeTagRepository はテスト用に repository.TagRepository を実装する。
type FakeTagRepository struct {
	CreateFn  func(context.Context, *entity.Tag) error
	ListFn    func(context.Context, uuid.UUID) ([]entity.Tag, error)
	GetByIDFn func(context.Context, uuid.UUID, uuid.UUID) (*entity.Tag, error)
	UpdateFn  func(context.Context, *entity.Tag) error
	DeleteFn  func(context.Context, uuid.UUID, uuid.UUID) error
}

func (f *FakeTagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	if f.CreateFn != nil {
		return f.CreateFn(ctx, tag)
	}
	return nil
}

func (f *FakeTagRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Tag, error) {
	if f.ListFn != nil {
		return f.ListFn(ctx, userID)
	}
	return nil, nil
}

func (f *FakeTagRepository) GetByID(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Tag, error) {
	if f.GetByIDFn != nil {
		return f.GetByIDFn(ctx, userID, id)
	}
	return nil, errors.New("GetByID not implemented")
}

func (f *FakeTagRepository) Update(ctx context.Context, tag *entity.Tag) error {
	if f.UpdateFn != nil {
		return f.UpdateFn(ctx, tag)
	}
	return nil
}

func (f *FakeTagRepository) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	if f.DeleteFn != nil {
		return f.DeleteFn(ctx, userID, id)
	}
	return nil
}
