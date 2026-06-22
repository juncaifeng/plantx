package app

import (
	"context"
	"testing"

	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/plantx/services/order/internal/domain"
	"github.com/plantx/services/order/internal/infra/repo"
)

func TestCreateOrder(t *testing.T) {
	ctx := kitctx.WithTenant(context.Background(), tenant.Info{ID: "t_001"})
	svc := NewOrderService(repo.NewInMemoryRepo())

	order, err := svc.CreateOrder(ctx, "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.TenantID != "t_001" {
		t.Errorf("expected tenant t_001, got %s", order.TenantID)
	}
}

func TestCreateOrderMissingTenant(t *testing.T) {
	ctx := context.Background()
	svc := NewOrderService(repo.NewInMemoryRepo())

	_, err := svc.CreateOrder(ctx, "Alice")
	if err == nil {
		t.Error("expected error for missing tenant")
	}
}

func TestListOrdersTenantIsolation(t *testing.T) {
	repo := repo.NewInMemoryRepo()
	svc := NewOrderService(repo)

	ctxA := kitctx.WithTenant(context.Background(), tenant.Info{ID: "t_001"})
	ctxB := kitctx.WithTenant(context.Background(), tenant.Info{ID: "t_002"})

	if _, err := svc.CreateOrder(ctxA, "Alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.CreateOrder(ctxB, "Bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	orders, err := svc.ListOrders(ctxA, domain.StatusPending)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order for tenant A, got %d", len(orders))
	}
	if orders[0].CustomerName != "Alice" {
		t.Errorf("expected Alice, got %s", orders[0].CustomerName)
	}
}
