package model

import (
	"context"
	"errors"
	"time"
)

const (
	StatusInProgress = "IN_PROGRESS"
	StatusFinished   = "FINISHED"
	StatusCanceled   = "CANCELED"
)

var (
	ErrNotFound        = errors.New("rental not found")
	ErrInvalidDates    = errors.New("dateFrom must be before dateTo")
	ErrAlreadyFinished = errors.New("rental already finished")
	ErrAlreadyCanceled = errors.New("rental already canceled")
	ErrForbidden       = errors.New("rental does not belong to user")
)

type Rental struct {
	RentalUID  string
	Username   string
	PaymentUID string
	CarUID     string
	DateFrom   time.Time
	DateTo     time.Time
	Status     string
}

type RentalRepository interface {
	GetByUsername(ctx context.Context, name string) ([]*Rental, error)
	GetByUID(ctx context.Context, rentalUid string) (*Rental, error)
	Create(ctx context.Context, rental *Rental) (*Rental, error)
	UpdateStatus(ctx context.Context, rentalUid, status string) error
}
