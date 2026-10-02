package repository

import (
	"context"
	"errors"

	"github.com/LeUrok/DS-lab2/payment-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *model.Payment) (*model.Payment, error) {
	_, err := r.pool.Exec(ctx,
		`insert into payment (payment_uid, status, price) values ($1, $2, $3)`,
		p.PaymentUID, p.Status, p.Price)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *PaymentRepository) GetByUID(ctx context.Context, pUID string) (*model.Payment, error) {
	var p model.Payment
	err := r.pool.QueryRow(ctx,
		`select payment_uid, status, price from payment where payment_uid=$1`, pUID).
		Scan(&p.PaymentUID, &p.Status, &p.Price)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PaymentRepository) Cancel(ctx context.Context, pUID string) error {
	tag, err := r.pool.Exec(ctx,
		`update payment set status = $1 where payment_uid = $2`, model.StatusCanceled, pUID)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}

	return nil
}
