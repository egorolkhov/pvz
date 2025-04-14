package pvz

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
	storageMocks "avito/internal/storage/mocks"
	// Если требуется для авторизации, хотя в этих тестах не используется
	authMocks "avito/pkg/auth/mocks"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreatePvz_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	reqObj := models.PVZ{
		City: "Москва",
	}
	bodyBytes, _ := json.Marshal(reqObj)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "moderator")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	storageMock.
		EXPECT().
		CreatePvz(gomock.Any(), gomock.AssignableToTypeOf(models.PVZ{})).
		DoAndReturn(func(ctx context.Context, pvz models.PVZ) error {
			pvz.ID = uuid.New()
			pvz.RegistrationDate = time.Now()
			return nil
		})

	metricsMock.
		EXPECT().
		IncPVZCreated()

	handler.CreatePvz(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp models.PVZ
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Москва", resp.City)
}

func TestCreatePvz_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader([]byte(`{"City": "Moscow"`)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "moderator")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreatePvz(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreatePvz_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	reqObj := models.PVZ{
		City: "Moscow",
	}
	bodyBytes, _ := json.Marshal(reqObj)
	req := httptest.NewRequest("POST", "/pvz", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreatePvz(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestPvzList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки.
	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)

	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	reqObj := models.PvzListRequest{
		Page:  1,
		Limit: 10,
	}
	bodyBytes, _ := json.Marshal(reqObj)
	req := httptest.NewRequest("GET", "/pvz", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	pvz1 := models.PVZ{ID: uuid.New(), City: "Moscow"}
	pvz2 := models.PVZ{ID: uuid.New(), City: "Saint Petersburg"}
	expectedList := []models.PVZ{pvz1, pvz2}

	storageMock.
		EXPECT().
		GetPvzList(gomock.Any(), gomock.AssignableToTypeOf(models.PvzListRequest{})).
		Return(expectedList, nil)

	storageMock.
		EXPECT().
		GetPVZReceptions(gomock.Any(), pvz1.ID, gomock.Any(), gomock.Any()).
		Return([]models.Reception{}, nil)
	storageMock.
		EXPECT().
		GetPVZReceptions(gomock.Any(), pvz2.ID, gomock.Any(), gomock.Any()).
		Return([]models.Reception{}, nil)

	handler.PvzList(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []models.PVZDetail
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, pvz1, resp[0].PVZ)
	assert.Equal(t, pvz2, resp[1].PVZ)
}

func TestPvzList_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	req := httptest.NewRequest("GET", "/pvz", bytes.NewReader([]byte(`{"page": "one"`)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.PvzList(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestPvzList_NoPermission проверяет, что если роль не позволяет получение списка, возвращается 403.
func TestPvzList_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	reqObj := models.PvzListRequest{Page: 1, Limit: 10}
	bodyBytes, _ := json.Marshal(reqObj)
	req := httptest.NewRequest("GET", "/pvz", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "client")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.PvzList(rec, req)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestPvzList_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	reqObj := models.PvzListRequest{Page: 1, Limit: 10}
	bodyBytes, _ := json.Marshal(reqObj)
	req := httptest.NewRequest("GET", "/pvz", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	storageMock.
		EXPECT().
		GetPvzList(gomock.Any(), gomock.AssignableToTypeOf(models.PvzListRequest{})).
		Return(nil, errors.New("db error"))

	handler.PvzList(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPvzList_GetReceptionsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storageMock := storageMocks.NewMockStorage(ctrl)
	tokenManagerMock := authMocks.NewMockTokenManager(ctrl)
	metricsMock := metricsMocks.NewMockMetrics(ctrl)
	mgr := manager.NewManager(storageMock, tokenManagerMock, metricsMock)
	handler := NewPvzHandler(mgr)

	reqObj := models.PvzListRequest{Page: 1, Limit: 10}
	bodyBytes, _ := json.Marshal(reqObj)
	req := httptest.NewRequest("GET", "/pvz", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(context.Background(), "role", "employee")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	pvz1 := models.PVZ{ID: uuid.New(), City: "Moscow"}
	expectedList := []models.PVZ{pvz1}

	storageMock.
		EXPECT().
		GetPvzList(gomock.Any(), gomock.AssignableToTypeOf(models.PvzListRequest{})).
		Return(expectedList, nil)

	storageMock.
		EXPECT().
		GetPVZReceptions(gomock.Any(), pvz1.ID, gomock.Any(), gomock.Any()).
		Return(nil, errors.New("receptions error"))

	handler.PvzList(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
