package service

import (
	"context"

	"github.com/LeUrok/DS-lab2/rental-service/internal/model"
	"github.com/google/uuid"
)

type RentalService struct {
	rep model.RentalRepository
}

func NewRentalService(rep model.RentalRepository) *RentalService {
	return &RentalService{rep: rep}
}

func (s *RentalService) GetByUsername(ctx context.Context, name string) ([]*model.Rental, error) {
	if name == "" {
		return nil, model.ErrForbidden
	}

	return s.rep.GetByUsername(ctx, name)
}

func (s *RentalService) GetByUID(ctx context.Context, rentalUID string, username string) (*model.Rental, error) {
	rental, err := s.rep.GetByUID(ctx, rentalUID)
	if err != nil {
		return nil, err
	}
	if rental.Username != username {
		return nil, model.ErrForbidden
	}
	return rental, nil
}

func (s *RentalService) Create(ctx context.Context, rental *model.Rental) (*model.Rental, error) {
	if !rental.DateFrom.Before(rental.DateTo) {
		return nil, model.ErrInvalidDates
	}
	rental.RentalUID = uuid.NewString()
	rental.Status = model.StatusInProgress
	return s.rep.Create(ctx, rental)
}

func (s *RentalService) Finish(ctx context.Context, rentalUid, username string) error {
	rental, err := s.rep.GetByUID(ctx, rentalUid)
	if err != nil {
		return err
	}
	if rental.Username != username {
		return model.ErrForbidden
	}
	if rental.Status == model.StatusFinished {
		return model.ErrAlreadyFinished
	}
	if rental.Status == model.StatusCanceled {
		return model.ErrAlreadyCanceled
	}
	return s.rep.UpdateStatus(ctx, rentalUid, model.StatusFinished)
}

func (s *RentalService) Cancel(ctx context.Context, rentalUid, username string) error {
	rental, err := s.rep.GetByUID(ctx, rentalUid)
	if err != nil {
		return err
	}
	if rental.Username != username {
		return model.ErrForbidden
	}
	if rental.Status == model.StatusFinished {
		return model.ErrAlreadyFinished
	}
	if rental.Status == model.StatusCanceled {
		return model.ErrAlreadyCanceled
	}
	return s.rep.UpdateStatus(ctx, rentalUid, model.StatusCanceled)
}
