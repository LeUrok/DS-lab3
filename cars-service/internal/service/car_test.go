package service

import (
	"context"
	"testing"

	"github.com/LeUrok/DS-lab2/cars-service/internal/model"
)

type mockCarRepo struct {
	cars []*model.Car
}

func (m *mockCarRepo) GetAll(ctx context.Context, page, size int, showAll bool) ([]*model.Car, int, error) {
	return m.cars, len(m.cars), nil
}

func (m *mockCarRepo) GetByUID(ctx context.Context, uid string) (*model.Car, error) {
	return nil, model.ErrNotFound
}

func (m *mockCarRepo) Reserve(ctx context.Context, uid string) error {
	return nil
}

func (m *mockCarRepo) Unreserve(ctx context.Context, uid string) error {
	return nil
}

func TestCarService_GetAll_DefaultPageSize(t *testing.T) {
	repo := &mockCarRepo{cars: []*model.Car{{CarUID: "abc"}}}
	svc := NewCarService(repo)

	_, _, err := svc.GetAll(context.Background(), -1, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCarService_GetAll_SizeTooLarge(t *testing.T) {
	repo := &mockCarRepo{}
	svc := NewCarService(repo)

	cars, _, err := svc.GetAll(context.Background(), 0, 9999, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = cars
}