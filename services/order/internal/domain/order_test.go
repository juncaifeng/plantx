package domain

import (
	"testing"
)

func TestNewOrder(t *testing.T) {
	order, err := NewOrder("t_001", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.TenantID != "t_001" {
		t.Errorf("expected tenant t_001, got %s", order.TenantID)
	}
	if order.CustomerName != "Alice" {
		t.Errorf("expected customer Alice, got %s", order.CustomerName)
	}
	if order.Status != StatusPending {
		t.Errorf("expected status pending, got %s", order.Status)
	}
}

func TestNewOrderMissingTenant(t *testing.T) {
	_, err := NewOrder("", "Alice")
	if err == nil {
		t.Error("expected error for missing tenant")
	}
}

func TestOrderConfirm(t *testing.T) {
	order, _ := NewOrder("t_001", "Alice")
	if err := order.Confirm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != StatusConfirmed {
		t.Errorf("expected confirmed, got %s", order.Status)
	}
}
