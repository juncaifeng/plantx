package app

import (
	"context"
	"fmt"

	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/errors"
	"github.com/plantx/services/order/internal/domain"
)

// OrderService implements order use cases.
type OrderService struct {
	repo domain.Repository
}

// NewOrderService creates a new OrderService.
func NewOrderService(repo domain.Repository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder creates a new order in the current tenant.
func (s *OrderService) CreateOrder(ctx context.Context, customerName string) (*domain.Order, error) {
	tenant := kitctx.GetTenant(ctx)
	if tenant.ID == "" {
		return nil, errors.New(errors.CodeInvalidInput, "tenant context missing")
	}
	order, err := domain.NewOrder(tenant.ID, customerName)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInvalidInput, "invalid order", err)
	}
	if err := s.repo.Save(ctx, order); err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to save order", err)
	}
	return order, nil
}

// GetOrder retrieves an order by id, implicitly scoped by tenant context.
func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to get order", err)
	}
	if order == nil {
		return nil, errors.New(errors.CodeNotFound, fmt.Sprintf("order %s not found", id))
	}
	return order, nil
}

// ListOrders lists orders filtered by status in the current tenant.
func (s *OrderService) ListOrders(ctx context.Context, status string) ([]*domain.Order, error) {
	tenant := kitctx.GetTenant(ctx)
	if tenant.ID == "" {
		return nil, errors.New(errors.CodeInvalidInput, "tenant context missing")
	}
	if status == "" {
		status = domain.StatusPending
	}
	return s.repo.ListByStatus(ctx, tenant.ID, status)
}
