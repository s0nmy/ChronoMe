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

// AuthUsecase はユーザー登録と認証を調整する。
type AuthUsecase struct {
	users repository.UserRepository
}

func NewAuthUsecase(users repository.UserRepository) *AuthUsecase {
	return &AuthUsecase{users: users}
}

// SignupParams はユーザー作成に必要なデータをまとめる。
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
	// 既存メールを先に確認して、同一メールのユーザー作成を避ける。
	_, err := u.users.GetByEmail(ctx, params.Email)
	if err == nil {
		return nil, fmt.Errorf("user already exists")
	}

	// パスワードは平文を保存せず、bcrypt hash だけを User に持たせる。
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
	// entity 側で正規化と不変条件の検証を行ってから永続化する。
	user.Normalize()
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := u.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login はユーザーの認証情報を検証し、成功時にエンティティを返す。
func (u *AuthUsecase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	// メール未存在とパスワード不一致は同じエラーにして、認証情報の推測を防ぐ。
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}

func (u *AuthUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return u.users.GetByID(ctx, userID)
}
