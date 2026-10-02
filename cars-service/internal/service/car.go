package service

import (
	"context"

	"github.com/LeUrok/DS-lab2/cars-service/internal/model"
)

type CarService struct {
	rep model.CarRepository
}

func NewCarService(rep model.CarRepository) *CarService {
	return &CarService{rep: rep}
}

func (s *CarService) GetAll(ctx context.Context, page, size int, showAll bool) ([]*model.Car, int, error) {
	if page < 0 {
		page = 0
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return s.rep.GetAll(ctx, page, size, showAll)
}

func (s *CarService) GetByUID(ctx context.Context, carUid string) (*model.Car, error) {
	return s.rep.GetByUID(ctx, carUid)
}

func (s *CarService) Reserve(ctx context.Context, carUid string) error {
	return s.rep.Reserve(ctx, carUid)
}

func (s *CarService) Unreserve(ctx context.Context, carUid string) error {
	return s.rep.Unreserve(ctx, carUid)
}
