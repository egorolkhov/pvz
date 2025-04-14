package reception

import (
	"avito/internal/manager/services/pvz"
	metricsMocks "avito/internal/metrics/mocks"
	"avito/internal/models"
	"avito/internal/storage"
	"avito/internal/storage/mocks"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReceptionService_Create(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(ctx, pvzID).
		Return(models.Reception{}, storage.ErrNotFound)

	mockStorage.
		EXPECT().
		CreateReception(gomock.Any(), gomock.AssignableToTypeOf(models.Reception{})).
		DoAndReturn(func(_ context.Context, rec models.Reception) error {
			return nil
		})

	mockMetrics.
		EXPECT().
		IncReceptionsCreated()

	rec, err := service.Create(ctx, pvzID)
	assert.NoError(t, err)
	assert.Equal(t, pvzID, rec.PVZID)
	assert.Equal(t, models.StatusInProgress, rec.Status)
}

func TestReceptionService_Create_NoPermission(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "client")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	_, err := service.Create(ctx, pvzID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, pvz.ErrNoPermission)
}

func TestReceptionService_Create_OpenReceptionExists(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	openRec := models.Reception{
		ID:     uuid.New(),
		Status: models.StatusInProgress,
	}
	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(ctx, pvzID).
		Return(openRec, nil)

	_, err := service.Create(ctx, pvzID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOpenReception)
}

func TestReceptionService_Close(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	openRec := models.Reception{
		ID:     uuid.New(),
		PVZID:  pvzID,
		Status: models.StatusInProgress,
	}

	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(ctx, pvzID).
		Return(openRec, nil)

	mockStorage.
		EXPECT().
		UpdateReceptionStatus(ctx, openRec.ID, models.StatusClosed).
		Return(nil)

	rec, err := service.Close(ctx, pvzID)
	assert.NoError(t, err)
	assert.Equal(t, models.StatusClosed, rec.Status)
}

func TestReceptionService_Close_NoOpen(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(ctx, pvzID).
		Return(models.Reception{}, errors.New("not found"))

	_, err := service.Close(ctx, pvzID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoOpenReception)
}

func TestReceptionService_AddProduct(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	// Разрешаем тип продукта "box".
	productTypes["box"] = struct{}{}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	openRec := models.Reception{
		ID:     uuid.New(),
		PVZID:  pvzID,
		Status: models.StatusInProgress,
	}

	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(gomock.Any(), pvzID).
		Return(openRec, nil)

	mockStorage.
		EXPECT().
		AddProduct(gomock.Any(), gomock.AssignableToTypeOf(models.Product{})).
		DoAndReturn(func(_ context.Context, prod models.Product) error {
			// Для теста возвращаем nil, имитируя успешное добавление продукта.
			return nil
		})

	mockMetrics.
		EXPECT().
		IncProductsCreated()

	product, err := service.AddProduct(ctx, "box", pvzID)
	assert.NoError(t, err)
	assert.Equal(t, openRec.ID, product.ReceptionID)
	assert.Equal(t, "box", product.Type)
}

func TestReceptionService_DeleteLastProduct(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	openRec := models.Reception{
		ID:     uuid.New(),
		PVZID:  pvzID,
		Status: models.StatusInProgress,
	}
	lastProduct := models.Product{
		ID:          uuid.New(),
		ReceptionID: openRec.ID,
	}

	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(ctx, pvzID).
		Return(openRec, nil)

	mockStorage.
		EXPECT().
		GetLastProductByReception(ctx, openRec.ID).
		Return(lastProduct, nil)

	mockStorage.
		EXPECT().
		DeleteProduct(ctx, lastProduct.ID).
		Return(nil)

	err := service.DeleteLastProduct(ctx, pvzID)
	assert.NoError(t, err)
}

func TestReceptionService_DeleteLastProduct_NoOpenReception(t *testing.T) {
	ctx := context.WithValue(context.Background(), "role", "employee")
	pvzID := uuid.New()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockMetrics := metricsMocks.NewMockMetrics(ctrl)

	service := ReceptionService{
		Storage: mockStorage,
		Metrics: mockMetrics,
	}

	mockStorage.
		EXPECT().
		GetLastReceptionInProgress(ctx, pvzID).
		Return(models.Reception{}, errors.New("not found"))

	err := service.DeleteLastProduct(ctx, pvzID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoOpenReception)
}
