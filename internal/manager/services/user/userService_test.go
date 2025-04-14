package user

import (
	"avito/internal/manager/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"avito/internal/models"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	mockStorage "avito/internal/storage/mocks"
	mocsAuth "avito/pkg/auth/mocks"
)

func TestUserService_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := mocsAuth.NewMockTokenManager(ctrl)
	mockStorage := mockStorage.NewMockStorage(ctrl)

	us := UserService{
		Storage:      mockStorage,
		TokenManager: mockTM,
	}

	mockTM.
		EXPECT().
		BuildToken(gomock.Any(), "employee").
		Return("dummy-token", nil)

	token, err := us.DummyLogin(context.Background(), "employee")
	assert.NoError(t, err)
	assert.Equal(t, "dummy-token", token)
}

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := mocsAuth.NewMockTokenManager(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	us := UserService{
		Storage:      mockStorage,
		TokenManager: mockTM,
	}

	mockStorage.
		EXPECT().
		CreateUser(gomock.Any(), gomock.AssignableToTypeOf(models.User{})).
		Return(nil)

	email := "test@example.com"
	role := "employee"
	pass := "password"

	usr, err := us.Register(context.Background(), email, role, pass)
	assert.NoError(t, err)
	assert.Equal(t, email, usr.Email)
	assert.Equal(t, role, usr.Role)
	assert.NotEmpty(t, usr.PasswordHash)
	assert.NotEqual(t, pass, usr.PasswordHash)
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := mocsAuth.NewMockTokenManager(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	us := UserService{
		Storage:      mockStorage,
		TokenManager: mockTM,
	}

	email := "test@example.com"
	password := "123456"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	userVal := models.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		Email:        email,
		Role:         "employee",
		PasswordHash: string(hashed),
	}

	mockStorage.
		EXPECT().
		CheckUser(gomock.Any(), email).
		Return(userVal, nil)

	mockTM.
		EXPECT().
		BuildToken(gomock.Any(), "employee").
		Return("login-token", nil)

	token, err := us.Login(context.Background(), email, password)
	assert.NoError(t, err)
	assert.Equal(t, "login-token", token)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := mocsAuth.NewMockTokenManager(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	us := UserService{
		Storage:      mockStorage,
		TokenManager: mockTM,
	}

	email := "test@example.com"
	wrongPassword := "wrongpass"
	correctPassword := "CorrectPassword"

	hashed, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	userVal := models.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		Email:        email,
		Role:         "employee",
		PasswordHash: string(hashed),
	}

	mockStorage.
		EXPECT().
		CheckUser(gomock.Any(), email).
		Return(userVal, nil)

	token, err := us.Login(context.Background(), email, wrongPassword)
	assert.Empty(t, token)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWrongPassword)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTM := mocsAuth.NewMockTokenManager(ctrl)
	mockStorage := mocks.NewMockStorage(ctrl)

	us := UserService{
		Storage:      mockStorage,
		TokenManager: mockTM,
	}

	email := "nonexistent@example.com"
	password := "123456"

	mockStorage.
		EXPECT().
		CheckUser(gomock.Any(), email).
		Return(models.User{}, errors.New("not found"))

	token, err := us.Login(context.Background(), email, password)
	assert.Empty(t, token)
	assert.Error(t, err)
}
