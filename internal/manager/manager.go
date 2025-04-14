package manager

import (
	"avito/internal/manager/services/pvz"
	"avito/internal/manager/services/reception"
	"avito/internal/manager/services/user"
	"avito/internal/metrics"
	"avito/internal/models"
	"avito/pkg/auth"
	"context"
	"github.com/google/uuid"
	"time"
)

//go:generate mockgen -source=manager.go -destination=mocks/manager_mock.go -package=mocks

type Storage interface {
	CreateUser(ctx context.Context, user models.User) error
	CheckUser(ctx context.Context, email string) (models.User, error)

	CreatePvz(ctx context.Context, pvz models.PVZ) error
	GetPvzList(ctx context.Context, pvzList models.PvzListRequest) ([]models.PVZ, error)
	GetPVZReceptions(ctx context.Context, pvzId uuid.UUID, startDate, endDate *time.Time) ([]models.Reception, error)
	GetReceptionProducts(ctx context.Context, receptionId uuid.UUID) ([]models.Product, error)

	CreateReception(ctx context.Context, reception models.Reception) error
	GetLastReceptionInProgress(ctx context.Context, pvzId uuid.UUID) (models.Reception, error)
	UpdateReceptionStatus(ctx context.Context, receptionID uuid.UUID, status string) error

	GetPVZByID(ctx context.Context, id uuid.UUID) (models.PVZ, error)
	AddProduct(ctx context.Context, product models.Product) error
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	GetLastProductByReception(ctx context.Context, receptionID uuid.UUID) (models.Product, error)
}

type Manager struct {
	user.UserService
	pvz.PvzService
	reception.ReceptionService
	TokenManager auth.TokenManager
	Metrics      metrics.Metrics
}

func NewManager(storage Storage, tokenManager auth.TokenManager, pm metrics.Metrics) *Manager {
	return &Manager{
		user.UserService{Storage: storage, TokenManager: tokenManager},
		pvz.PvzService{Storage: storage, Metrics: pm},
		reception.ReceptionService{Storage: storage, Metrics: pm},
		tokenManager,
		pm,
	}
}
