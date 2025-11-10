package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

// AuthUsecase coordinates user onboarding and authentication.
type AuthUsecase struct {
	users repository.UserRepository
}

func NewAuthUsecase(users repository.UserRepository) *AuthUsecase {
	return &AuthUsecase{users: users}
}

// SignupParams captures the data required to create a user.
type SignupParams struct {
	Email       string
	Password    string
	DisplayName string
	TimeZone    string
}

func (u *AuthUsecase) Signup(ctx context.Context, params SignupParams) (*entity.User, error) {
	if params.Email == "" || params.Password == "" {
		return nil, errors.New("email and password are required")
	}
	_, err := u.users.GetByEmail(ctx, params.Email)
	if err == nil {
		return nil, fmt.Errorf("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        params.Email,
		PasswordHash: string(hash),
		DisplayName:  params.DisplayName,
		TimeZone:     params.TimeZone,
	}
	user.Normalize()
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := u.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login validates a user's credentials and returns the entity on success.
func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (u *AuthUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return u.users.GetByID(ctx, userID)
}
