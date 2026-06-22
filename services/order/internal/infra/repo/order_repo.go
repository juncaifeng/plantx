package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/db"
	"github.com/plantx/services/order/internal/domain"
	"github.com/plantx/services/order/internal/infra/sqlc"
)

// OrderRepo implements domain.Repository using sqlc.
type OrderRepo struct {
	queries *sqlc.Queries
}

// NewOrderRepo creates a new OrderRepo.
func NewOrderRepo(d db.DB) *OrderRepo {
	return &OrderRepo{queries: sqlc.New(d)}
}

func tenantID(ctx context.Context) string {
	return kitctx.GetTenant(ctx).ID
}

// Save persists a new order.
func (r *OrderRepo) Save(ctx context.Context, order *domain.Order) error {
	_, err := r.queries.CreateOrder(ctx, sqlc.CreateOrderParams{
		TenantID:     order.TenantID,
		CustomerName: order.CustomerName,
		Status:       order.Status,
	})
	return err
}

// Get retrieves an order by id scoped to the current tenant.
func (r *OrderRepo) Get(ctx context.Context, id string) (*domain.Order, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	o, err := r.queries.GetOrder(ctx, sqlc.GetOrderParams{
		ID:       uid,
		TenantID: tenantID(ctx),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomain(o), nil
}

// ListByStatus lists orders for a tenant filtered by status.
func (r *OrderRepo) ListByStatus(ctx context.Context, tenantID, status string) ([]*domain.Order, error) {
	rows, err := r.queries.ListOrdersByStatus(ctx, sqlc.ListOrdersByStatusParams{
		TenantID: tenantID,
		Status:   status,
	})
	if err != nil {
		return nil, err
	}
	orders := make([]*domain.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, toDomain(row))
	}
	return orders, nil
}

func toDomain(o sqlc.Order) *domain.Order {
	return &domain.Order{
		ID:           o.ID.String(),
		TenantID:     o.TenantID,
		CustomerName: o.CustomerName,
		Status:       o.Status,
		CreatedAt:    o.CreatedAt,
	}
}
