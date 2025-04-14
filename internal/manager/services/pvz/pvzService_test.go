package pvz

import (
	metricsMocks "avito/internal/metrics/mocks"
	"avito/internal/models"
	storageMocks "avito/internal/storage/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPvzService_Create(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "moderator")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storageMocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := PvzService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	p := models.PVZ{
		City: "Москва",
	}

	mockStorage.
		EXPECT().
		CreatePvz(gomock.Any(), gomock.AssignableToTypeOf(models.PVZ{})).
		DoAndReturn(func(ctx context.Context, pvz models.PVZ) error {
			pvz.ID = uuid.New()
			pvz.RegistrationDate = time.Now()
			return nil
		})

	mockMetrics.
		EXPECT().
		IncPVZCreated()

	created, err := service.Create(ctx, p)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, created.ID)
	assert.Equal(t, "Москва", created.City)
	assert.False(t, created.RegistrationDate.IsZero())
}

func TestPvzService_Create_NoPermission(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storageMocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := PvzService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	_, err := service.Create(ctx, models.PVZ{City: "Moscow"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoPermission)
}

func TestPvzService_Create_ForbiddenCity(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "moderator")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storageMocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := PvzService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	_, err := service.Create(ctx, models.PVZ{City: "SPB"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbiddenCity)
}

func TestPvzService_GetList(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storageMocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := PvzService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	req := models.PvzListRequest{
		Page:  1,
		Limit: 10,
	}

	pvzList := []models.PVZ{
		{ID: uuid.New(), City: "Moscow"},
	}

	mockStorage.
		EXPECT().
		GetPvzList(gomock.Any(), req).
		Return(pvzList, nil)

	recID := uuid.New()
	receptions := []models.Reception{
		{ID: recID, PVZID: pvzList[0].ID, Status: models.StatusClosed},
	}
	mockStorage.
		EXPECT().
		GetPVZReceptions(gomock.Any(), pvzList[0].ID, (*time.Time)(nil), (*time.Time)(nil)).
		Return(receptions, nil)

	productList := []models.Product{
		{ID: uuid.New(), Type: "box", ReceptionID: recID},
	}
	mockStorage.
		EXPECT().
		GetReceptionProducts(gomock.Any(), recID).
		Return(productList, nil)

	result, err := service.GetList(ctx, req)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Receptions, 1)
	assert.Len(t, result[0].Receptions[0].Products, 1)
}

func TestPvzService_GetList_NoPermission(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "client")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storageMocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := PvzService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	req := models.PvzListRequest{}
	_, err := service.GetList(ctx, req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoPermission)
}

func TestPvzService_GetList_StorageError(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storageMocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := PvzService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	req := models.PvzListRequest{Page: 1, Limit: 10}

	mockStorage.
		EXPECT().
		GetPvzList(gomock.Any(), req).
		Return(nil, errors.New("some db error"))

	_, err := service.GetList(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some db error")
}
