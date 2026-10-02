package model

import (
	"context"
	"errors"
)

var (
	ErrNotFound        = errors.New("car not found")
	ErrAlreadyReserved = errors.New("car already reserved")
	ErrNotReserved     = errors.New("car not reserved")
)

type Car struct {
	CarUID             string
	Brand              string
	Model              string
	RegistrationNumber string
	Power              int
	Price              int
	Type               string
	Available          bool
}

type CarRepository interface {
	GetAll(ctx context.Context, page, size int, showAll bool) ([]*Car, int, error)
	GetByUID(ctx context.Context, carUid string) (*Car, error)
	Reserve(ctx context.Context, carUid string) error
	Unreserve(ctx context.Context, carUid string) error
}
