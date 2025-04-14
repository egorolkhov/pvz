package storage

import (
	"avito/internal/models"
	"avito/internal/storage/transactionManager"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"strconv"
	"time"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage_mock.go -package=mocks

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

type DataBase struct {
	Tm *transactionManager.Transactor
}

func NewStorage(tm *transactionManager.Transactor) *DataBase {
	return &DataBase{Tm: tm}
}

func (d *DataBase) CreateUser(ctx context.Context, user models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := d.Tm.DB.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
	)

	return err
}

func (d *DataBase) CheckUser(ctx context.Context, email string) (models.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var user models.User
	err := d.Tm.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (d *DataBase) CreatePvz(ctx context.Context, pvz models.PVZ) error {
	query := `
		INSERT INTO pvz (id, registration_date, city)
		VALUES ($1, $2, $3)
	`
	_, err := d.Tm.DB.ExecContext(
		ctx,
		query,
		pvz.ID,
		pvz.RegistrationDate,
		pvz.City,
	)
	return err
}

func (d *DataBase) GetPvzList(ctx context.Context, req models.PvzListRequest) ([]models.PVZ, error) {
	offset := (req.Page - 1) * req.Limit
	query := `
		SELECT id, registration_date, city
		FROM pvz
		ORDER BY registration_date DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := d.Tm.DB.QueryContext(ctx, query, req.Limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.PVZ
	for rows.Next() {
		var p models.PVZ
		if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (d *DataBase) CreateReception(ctx context.Context, reception models.Reception) error {
	query := `
		INSERT INTO receptions
		    (id, pvz_id, datetime, status)
		VALUES ($1, $2, $3, $4)
	`

	_, err := d.Tm.DB.ExecContext(
		ctx,
		query,
		reception.ID,
		reception.PVZID,
		reception.DateTime,
		reception.Status,
	)

	return err
}

func (d *DataBase) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := d.Tm.DB.ExecContext(ctx, query, productID)
	return err
}

func (d *DataBase) GetLastReceptionInProgress(ctx context.Context, pvzId uuid.UUID) (models.Reception, error) {
	query := `
		SELECT id, pvz_id, datetime, status
		FROM receptions
		WHERE pvz_id = $1 AND status = 'in_progress'
		ORDER BY datetime DESC
		LIMIT 1
	`
	row := d.Tm.DB.QueryRowContext(ctx, query, pvzId)

	var rec models.Reception
	if err := row.Scan(&rec.ID, &rec.PVZID, &rec.DateTime, &rec.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Reception{}, ErrNotFound
		}
		return models.Reception{}, err
	}
	return rec, nil
}

func (d *DataBase) GetLastProductByReception(ctx context.Context, receptionID uuid.UUID) (models.Product, error) {
	query := `
		SELECT id, reception_id, datetime, type
		FROM products
		WHERE reception_id = $1
		ORDER BY datetime DESC
		LIMIT 1
	`
	row := d.Tm.DB.QueryRowContext(ctx, query, receptionID)

	var p models.Product
	if err := row.Scan(&p.ID, &p.ReceptionID, &p.DateTime, &p.Type); err != nil {
		return models.Product{}, err
	}
	return p, nil
}

func (d *DataBase) AddProduct(ctx context.Context, product models.Product) error {
	query := `
		INSERT INTO products (id, reception_id, datetime, type)
		VALUES ($1, $2, $3, $4)
	`
	_, err := d.Tm.DB.ExecContext(ctx, query, product.ID, product.ReceptionID, product.DateTime, product.Type)
	return err
}

func (d *DataBase) UpdateReceptionStatus(ctx context.Context, receptionID uuid.UUID, status string) error {
	query := `
		UPDATE receptions
		SET status = $1
		WHERE id = $2
	`
	_, err := d.Tm.DB.ExecContext(ctx, query, status, receptionID)
	return err
}

func (d *DataBase) GetPVZByID(ctx context.Context, id uuid.UUID) (models.PVZ, error) {
	query := `
		SELECT id, registration_date, city
		FROM pvz
		WHERE id = $1
	`
	row := d.Tm.DB.QueryRowContext(ctx, query, id)
	var p models.PVZ
	if err := row.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
		return models.PVZ{}, err
	}
	return p, nil
}

func (d *DataBase) GetPVZReceptions(ctx context.Context, pvzId uuid.UUID, startDate, endDate *time.Time) ([]models.Reception, error) {
	query := `
		SELECT id, pvz_id, datetime, status
		FROM receptions
		WHERE pvz_id = $1
	`
	params := []interface{}{pvzId}
	paramIndex := 2

	if startDate != nil {
		query += " AND datetime >= $" + strconv.Itoa(paramIndex)
		params = append(params, *startDate)
		paramIndex++
	}
	if endDate != nil {
		query += " AND datetime <= $" + strconv.Itoa(paramIndex)
		params = append(params, *endDate)
		paramIndex++
	}

	query += " ORDER BY datetime DESC"

	rows, err := d.Tm.DB.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receptions []models.Reception
	for rows.Next() {
		var rec models.Reception
		if err := rows.Scan(&rec.ID, &rec.PVZID, &rec.DateTime, &rec.Status); err != nil {
			return nil, err
		}
		receptions = append(receptions, rec)
	}
	return receptions, nil
}

func (d *DataBase) GetReceptionProducts(ctx context.Context, receptionId uuid.UUID) ([]models.Product, error) {
	query := `
		SELECT id, reception_id, datetime, type
		FROM products
		WHERE reception_id = $1
		ORDER BY datetime DESC
	`
	rows, err := d.Tm.DB.QueryContext(ctx, query, receptionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.ReceptionID, &p.DateTime, &p.Type); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
