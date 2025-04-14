package user

import (
	"avito/internal/models"
	"avito/pkg/auth"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Storage interface {
	CreateUser(ctx context.Context, user models.User) error
	CheckUser(ctx context.Context, email string) (models.User, error)
}

type TransactionManager interface {
	WriteTXSerialize(ctx context.Context, fx func(*sql.Tx) error) error
	WriteTXRepeatableRead(ctx context.Context, fx func(*sql.Tx) error) error
}

type UserService struct {
	Storage Storage
	auth.TokenManager
}

func (us *UserService) DummyLogin(ctx context.Context, role string) (string, error) {
	token, err := us.TokenManager.BuildToken(uuid.New().String(), role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (us *UserService) Register(ctx context.Context, email, role, password string) (models.User, error) {
	const MinCost = 4
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), MinCost)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		Email:        email,
		Role:         role,
		PasswordHash: string(hashedPassword),
	}
	err = us.Storage.CreateUser(ctx, user)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (us *UserService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := us.Storage.CheckUser(ctx, email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrWrongPassword
	}

	token, err := us.TokenManager.BuildToken(uuid.New().String(), user.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}
