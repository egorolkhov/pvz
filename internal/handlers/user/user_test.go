package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito/internal/manager"
	metricsMocks "avito/internal/metrics/mocks"
	"avito/internal/models"
	storageMocks "avito/internal/storage/mocks"
	authMocks "avito/pkg/auth/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestDummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	tokenManagerMock.
		EXPECT().
		BuildToken(gomock.Any(), "employee").
		Return("fake-token", nil)

	payload := []byte(`{"role": "employee"}`)
	req, err := http.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.DummyLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var token string
	if err := json.Unmarshal(rec.Body.Bytes(), &token); err != nil {
		t.Errorf("error unmarshalling response: %v", err)
	}
	if token != "fake-token" {
		t.Errorf("expected token 'fake-token', got '%s'", token)
	}
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	hashed, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	user := models.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		Email:        "user@example.com",
		Role:         "employee",
		PasswordHash: string(hashed),
	}

	storageMock.
		EXPECT().
		CheckUser(gomock.Any(), "user@example.com").
		Return(user, nil)
	tokenManagerMock.
		EXPECT().
		BuildToken(gomock.Any(), "employee").
		Return("fake-token", nil)

	validPayload := map[string]string{
		"email":    "user@example.com",
		"password": "secret",
	}
	body, _ := json.Marshal(validPayload)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var token string
	if err := json.Unmarshal(rec.Body.Bytes(), &token); err != nil {
		t.Errorf("error decoding response: %v", err)
	}
	if token != "fake-token" {
		t.Errorf("expected token 'fake-token', got '%s'", token)
	}

	storageMock.
		EXPECT().
		CheckUser(gomock.Any(), "user@example.com").
		Return(user, nil)

	invalidPayload := map[string]string{
		"email":    "user@example.com",
		"password": "wrongpassword",
	}
	body, _ = json.Marshal(invalidPayload)
	req, err = http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	storageMock.
		EXPECT().
		CreateUser(gomock.Any(), gomock.AssignableToTypeOf(models.User{})).
		Return(nil)

	payload := map[string]string{
		"email":    "new@example.com",
		"password": "newpass",
		"role":     "employee",
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var userResp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &userResp); err != nil {
		t.Errorf("error unmarshalling response: %v", err)
	}
	if userResp["email"] != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%v'", userResp["email"])
	}
	if userResp["role"] != "employee" {
		t.Errorf("expected role 'employee', got '%v'", userResp["role"])
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte(`{"email": 123}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_CheckUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	storageMock.
		EXPECT().
		CheckUser(gomock.Any(), "nonexistent@example.com").
		Return(models.User{}, errors.New("user not found"))

	payload := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "secret",
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	hashed, err := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	assert.NoError(t, err)
	user := models.User{
		ID:           uuid.New(),
		CreatedAt:    time.Now(),
		Email:        "user@example.com",
		Role:         "employee",
		PasswordHash: string(hashed),
	}

	storageMock.
		EXPECT().
		CheckUser(gomock.Any(), "user@example.com").
		Return(user, nil)

	payload := map[string]string{
		"email":    "user@example.com",
		"password": "wrong",
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRegister_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte(`{"email":123}`))) // некорректный email
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegister_CreateUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewUserHandler(mgr)

	storageMock.
		EXPECT().
		CreateUser(gomock.Any(), gomock.AssignableToTypeOf(models.User{})).
		Return(errors.New("create error"))

	payload := map[string]string{
		"email":    "fail@example.com",
		"password": "pass",
		"role":     "employee",
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
