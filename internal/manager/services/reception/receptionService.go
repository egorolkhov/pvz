package reception

import (
	"avito/internal/manager/services/pvz"
	"avito/internal/models"
	"avito/internal/storage"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"time"
)

type Storage interface {
	CreateReception(ctx context.Context, reception models.Reception) error
	GetLastReceptionInProgress(ctx context.Context, pvzId uuid.UUID) (models.Reception, error)
	UpdateReceptionStatus(ctx context.Context, receptionID uuid.UUID, status string) error

	GetPVZByID(ctx context.Context, id uuid.UUID) (models.PVZ, error)
	AddProduct(ctx context.Context, product models.Product) error
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	GetLastProductByReception(ctx context.Context, receptionID uuid.UUID) (models.Product, error)
}

type Metrics interface {
	IncReceptionsCreated()
	IncProductsCreated()
}

type TransactionManager interface {
	WriteTXReadCommitted(ctx context.Context, fn func(*sql.Tx) error) error
}

type ReceptionService struct {
	Storage Storage
	Metrics Metrics
}

func (rs *ReceptionService) Create(ctx context.Context, pvzId uuid.UUID) (models.Reception, error) {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "employee" {
		return models.Reception{}, pvz.ErrNoPermission
	}

	_, err := rs.Storage.GetLastReceptionInProgress(ctx, pvzId)
	if err == nil {
		return models.Reception{}, ErrOpenReception
	} else if !errors.Is(err, storage.ErrNotFound) {
		return models.Reception{}, err
	}

	reception := models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzId,
		Status:   models.StatusInProgress,
	}
	if err = rs.Storage.CreateReception(ctx, reception); err != nil {
		return models.Reception{}, err
	}
	rs.Metrics.IncReceptionsCreated()
	return reception, nil
}

func (rs *ReceptionService) Close(ctx context.Context, pvzId uuid.UUID) (models.Reception, error) {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "employee" {
		return models.Reception{}, pvz.ErrNoPermission
	}

	reception, err := rs.Storage.GetLastReceptionInProgress(ctx, pvzId)
	if err != nil {
		return models.Reception{}, ErrNoOpenReception
	}
	if err = rs.Storage.UpdateReceptionStatus(ctx, reception.ID, models.StatusClosed); err != nil {
		return models.Reception{}, err
	}
	reception.Status = models.StatusClosed
	return reception, nil
}

func (rs *ReceptionService) AddProduct(ctx context.Context, productType string, pvzId uuid.UUID) (models.Product, error) {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "employee" {
		return models.Product{}, pvz.ErrNoPermission
	}

	if _, allowed := productTypes[productType]; !allowed {
		return models.Product{}, ErrForbiddenProductType
	}

	reception, err := rs.Storage.GetLastReceptionInProgress(ctx, pvzId)
	if err != nil {
		return models.Product{}, ErrNoOpenReception
	}

	product := models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: reception.ID,
	}

	if err = rs.Storage.AddProduct(ctx, product); err != nil {
		return models.Product{}, err
	}
	rs.Metrics.IncProductsCreated()
	return product, nil
}

func (rs *ReceptionService) DeleteLastProduct(ctx context.Context, pvzId uuid.UUID) error {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "employee" {
		return pvz.ErrNoPermission
	}

	reception, err := rs.Storage.GetLastReceptionInProgress(ctx, pvzId)
	if err != nil {
		return ErrNoOpenReception
	}
	product, err := rs.Storage.GetLastProductByReception(ctx, reception.ID)
	if err != nil {
		return ErrNoProductToDelete
	}
	return rs.Storage.DeleteProduct(ctx, product.ID)
	return err
}
