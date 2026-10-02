package service

import (
	"context"
	"github.com/google/uuid"

	"github.com/LeUrok/DS-lab2/payment-service/internal/model"
)

type PaymentService struct {
	rep model.PaymentRepository
}

func NewPaymentService(rep model.PaymentRepository) *PaymentService {
	return &PaymentService{rep: rep}
}

func (s *PaymentService) Create(ctx context.Context, price int) (*model.Payment, error) {
	if price <= 0 {
		return nil, model.ErrInvalidPrice
	}

	return s.rep.Create(ctx, &model.Payment{
		PaymentUID: uuid.New().String(),
		Status:     model.StatusPaid,
		Price:      price,
	})
}

func (s *PaymentService) GetByUID(ctx context.Context, pUID string) (*model.Payment, error) {
	return s.rep.GetByUID(ctx, pUID)
}

func (s *PaymentService) Cancel(ctx context.Context, pUID string) error {
	return s.rep.Cancel(ctx, pUID)
}
