package model

import (
	"context"
	"errors"
)

var (
	ErrNotFound     = errors.New("payment not found")
	ErrInvalidPrice = errors.New("price must be positive")
)

const (
	StatusPaid     = "PAID"
	StatusCanceled = "CANCELED"
)

type Payment struct {
	PaymentUID string
	Status     string
	Price      int
}

type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) (*Payment, error)
	GetByUID(ctx context.Context, pUID string) (*Payment, error)
	Cancel(ctx context.Context, pUID string) error
}
