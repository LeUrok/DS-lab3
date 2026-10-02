package repository

import (
	"context"
	"errors"

	"github.com/LeUrok/DS-lab2/rental-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RentalRepository struct {
	pool *pgxpool.Pool
}

func NewRentalRepository(pool *pgxpool.Pool) *RentalRepository {
	return &RentalRepository{pool: pool}
}

func (r *RentalRepository) GetByUsername(ctx context.Context, name string) ([]*model.Rental, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT rental_uid, username, payment_uid, car_uid, date_from, date_to, status
		 				FROM rental WHERE username = $1 ORDER BY date_from DESC`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rentals := make([]*model.Rental, 0)
	for rows.Next() {
		var rental model.Rental
		if err := rows.Scan(&rental.RentalUID, &rental.Username, &rental.PaymentUID,
			&rental.CarUID, &rental.DateFrom, &rental.DateTo, &rental.Status); err != nil {
			return nil, err
		}
		rentals = append(rentals, &rental)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rentals, nil
}

func (r *RentalRepository) GetByUID(ctx context.Context, rentalUID string) (*model.Rental, error) {
	var rental model.Rental
	err := r.pool.QueryRow(ctx,
		`SELECT rental_uid, username, payment_uid, car_uid, date_from, date_to, status
		 					FROM rental WHERE rental_uid = $1`, rentalUID).Scan(
		&rental.RentalUID, &rental.Username, &rental.PaymentUID,
		&rental.CarUID, &rental.DateFrom, &rental.DateTo, &rental.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *RentalRepository) Create(ctx context.Context, rental *model.Rental) (*model.Rental, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rental (rental_uid, username, payment_uid, car_uid, date_from, date_to, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		rental.RentalUID, rental.Username, rental.PaymentUID,
		rental.CarUID, rental.DateFrom, rental.DateTo, rental.Status)

	if err != nil {
		return nil, err
	}
	return rental, nil
}

func (r *RentalRepository) UpdateStatus(ctx context.Context, rentalUid, status string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE rental SET status = $1 WHERE rental_uid = $2`, status, rentalUid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
