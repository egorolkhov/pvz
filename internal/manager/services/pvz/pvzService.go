package pvz

import (
	"avito/internal/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

type Storage interface {
	CreatePvz(ctx context.Context, pvz models.PVZ) error
	GetPvzList(ctx context.Context, pvzList models.PvzListRequest) ([]models.PVZ, error)
	GetPVZReceptions(ctx context.Context, pvzId uuid.UUID, startDate, endDate *time.Time) ([]models.Reception, error)
	GetReceptionProducts(ctx context.Context, receptionId uuid.UUID) ([]models.Product, error)
}

type Metrics interface {
	IncPVZCreated()
}

type TransactionManager interface {
	WriteTXSerialize(ctx context.Context, fx func(*sql.Tx) error) error
	WriteTXRepeatableRead(ctx context.Context, fx func(*sql.Tx) error) error
}

type PvzService struct {
	Storage Storage
	Metrics Metrics
}

func (us *PvzService) Create(ctx context.Context, pvz models.PVZ) (models.PVZ, error) {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "moderator" {
		return models.PVZ{}, ErrNoPermission
	}

	if _, allowed := AllowedCities[pvz.City]; !allowed {
		return models.PVZ{}, ErrForbiddenCity
	}
	if pvz.ID == uuid.Nil {
		pvz.ID = uuid.New()
	}
	if pvz.RegistrationDate.IsZero() {
		pvz.RegistrationDate = time.Now()
	}
	err := us.Storage.CreatePvz(ctx, pvz)
	if err != nil {
		return models.PVZ{}, err
	}
	us.Metrics.IncPVZCreated()
	return pvz, nil
}

func (us *PvzService) GetList(ctx context.Context, req models.PvzListRequest) ([]models.PVZDetail, error) {
	role, ok := ctx.Value("role").(string)
	if !ok || role != "employee" && role != "moderator" {
		return nil, ErrNoPermission
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	list, err := us.Storage.GetPvzList(ctx, req)
	var result []models.PVZDetail
	for _, p := range list {
		recs, err := us.Storage.GetPVZReceptions(ctx, p.ID, req.StartDate, req.EndDate)
		if err != nil {
			return nil, err
		}

		var recDetails []models.ReceptionDetail
		for _, rec := range recs {
			prods, err := us.Storage.GetReceptionProducts(ctx, rec.ID)
			if err != nil {
				return nil, err
			}
			recDetails = append(recDetails, models.ReceptionDetail{
				Reception: rec,
				Products:  prods,
			})
		}

		result = append(result, models.PVZDetail{
			PVZ:        p,
			Receptions: recDetails,
		})
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
