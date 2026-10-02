package repository

import (
	"context"
	"errors"

	"github.com/LeUrok/DS-lab2/cars-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CarRepository struct {
	pool *pgxpool.Pool
}

func NewCarRepository(p *pgxpool.Pool) *CarRepository {
	return &CarRepository{pool: p}
}

func (r *CarRepository) GetAll(ctx context.Context, page, size int, showAll bool) ([]*model.Car, int, error) {
	var total int
	countQuery := `select count(*) from cars`
	if !showAll {
		countQuery += ` where availability = true`
	}
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := page * size
	query := `SELECT car_uid, brand, model, registration_number, power, price, type, availability FROM cars`
	if !showAll {
		query += ` where availability = true`
	}
	query += ` ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cars []*model.Car
	for rows.Next() {
		var c model.Car
		if err := rows.Scan(&c.CarUID, &c.Brand, &c.Model, &c.RegistrationNumber,
			&c.Power, &c.Price, &c.Type, &c.Available); err != nil {
			return nil, 0, err
		}
		cars = append(cars, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return cars, total, nil
}

func (r *CarRepository) GetByUID(ctx context.Context, carUid string) (*model.Car, error) {
	var c model.Car
	err := r.pool.QueryRow(ctx,
		`SELECT car_uid, brand, model, registration_number, power, price, type, availability
		 			FROM cars WHERE car_uid = $1`, carUid).Scan(&c.CarUID, &c.Brand, &c.Model,
		&c.RegistrationNumber, &c.Power, &c.Price, &c.Type, &c.Available)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CarRepository) Reserve(ctx context.Context, carUID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE cars SET availability = false WHERE car_uid = $1 AND availability = true`,
		carUID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM cars WHERE car_uid = $1)`, carUID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return model.ErrNotFound
		}
		return model.ErrAlreadyReserved
	}
	return nil
}

func (r *CarRepository) Unreserve(ctx context.Context, carUID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE cars SET availability = true WHERE car_uid = $1 AND availability = false`,
		carUID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM cars WHERE car_uid = $1)`, carUID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return model.ErrNotFound
		}
		return model.ErrNotReserved
	}
	return nil
}
