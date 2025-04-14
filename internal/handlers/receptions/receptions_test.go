package receptions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"avito/internal/manager"
	metricsMocks "avito/internal/metrics/mocks"
	"avito/internal/models"
	"avito/internal/storage"
	storageMocks "avito/internal/storage/mocks"
	authMocks "avito/pkg/auth/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateReception_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createReception", bytes.NewReader(bodyBytes))
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	storageMock.
		EXPECT().
		GetLastReceptionInProgress(gomock.Any(), pvzID).
		Return(models.Reception{}, storage.ErrNotFound)

	storageMock.
		EXPECT().
		CreateReception(gomock.Any(), gomock.AssignableToTypeOf(models.Reception{})).
		DoAndReturn(func(ctx context.Context, rec models.Reception) error {
			rec.ID = uuid.New()
			rec.Status = models.StatusInProgress
			rec.PVZID = pvzID
			rec.DateTime = time.Now()
			return nil
		})

	metricsMock.
		EXPECT().
		IncReceptionsCreated()

	handler.CreateReception(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp models.Reception
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, pvzID, resp.PVZID)
	assert.Equal(t, models.StatusInProgress, resp.Status)
}

func TestCreateProduct_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"type":  "электроника",
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createProduct", bytes.NewReader(bodyBytes))
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	expectedProduct := models.Product{
		ID:          uuid.New(),
		Type:        "электроника",
		ReceptionID: uuid.New(),
	}

	storageMock.
		EXPECT().
		GetLastReceptionInProgress(gomock.Any(), pvzID).
		Return(models.Reception{
			ID:     expectedProduct.ReceptionID,
			PVZID:  pvzID,
			Status: models.StatusInProgress,
		}, nil)

	storageMock.
		EXPECT().
		AddProduct(gomock.Any(), gomock.AssignableToTypeOf(models.Product{})).
		DoAndReturn(func(ctx context.Context, prod models.Product) error {
			prod.ID = expectedProduct.ID
			return nil
		})

	metricsMock.
		EXPECT().
		IncProductsCreated()

	handler.CreateProduct(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp models.Product
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "электроника", resp.Type)
	assert.Equal(t, expectedProduct.ReceptionID, resp.ReceptionID)
}

func TestCloseReception_BadParams(t *testing.T) {
	mgr := manager.NewManager(nil, nil, nil)
	handler := NewReceptionsHandler(mgr)

	req := httptest.NewRequest("POST", "/closeReception", nil)
	rec := httptest.NewRecorder()

	handler.CloseReception(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateReception_BadRequest(t *testing.T) {
	mgr := manager.NewManager(nil, nil, nil)
	handler := NewReceptionsHandler(mgr)

	req := httptest.NewRequest("POST", "/createReception", bytes.NewBuffer([]byte(`{"wrongField": "value"}`)))
	rec := httptest.NewRecorder()

	handler.CreateReception(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_BadRequest(t *testing.T) {
	mgr := manager.NewManager(nil, nil, nil)
	handler := NewReceptionsHandler(mgr)

	req := httptest.NewRequest("POST", "/createProduct", bytes.NewBuffer([]byte(`{"type": "", "pvzId": "not-a-uuid"}`)))
	rec := httptest.NewRecorder()

	handler.CreateProduct(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateReception_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createReception", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "client")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateReception(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreateReception_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createReception", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	storageMock.
		EXPECT().
		GetLastReceptionInProgress(gomock.Any(), pvzID).
		Return(models.Reception{}, errors.New("db error"))

	handler.CreateReception(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"type":  "электроника",
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createProduct", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "client")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateProduct(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreateProduct_InvalidType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"type":  "furniture",
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createProduct", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateProduct(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateProduct_NoOpenReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewReceptionsHandler(mgr)

	pvzID := uuid.New()
	reqBody := map[string]string{
		"type":  "электроника",
		"pvzId": pvzID.String(),
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/createProduct", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	storageMock.
		EXPECT().
		GetLastReceptionInProgress(gomock.Any(), pvzID).
		Return(models.Reception{}, errors.New("no open reception"))

	handler.CreateProduct(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
