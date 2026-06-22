package grpc

import (
	"context"

	"github.com/plantx/kit/kit-go/errors"
	"github.com/plantx/services/order/api"
	"github.com/plantx/services/order/internal/app"
	"github.com/plantx/services/order/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the OrderService gRPC server.
type Handler struct {
	api.UnimplementedOrderServiceServer
	app *app.OrderService
}

// NewHandler creates a new Handler.
func NewHandler(app *app.OrderService) *Handler {
	return &Handler{app: app}
}

// CreateOrder handles order creation.
func (h *Handler) CreateOrder(ctx context.Context, req *api.CreateOrderRequest) (*api.Order, error) {
	order, err := h.app.CreateOrder(ctx, req.CustomerName)
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(order), nil
}

// GetOrder handles single order retrieval.
func (h *Handler) GetOrder(ctx context.Context, req *api.GetOrderRequest) (*api.Order, error) {
	order, err := h.app.GetOrder(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(order), nil
}

// ListOrders handles listing orders.
func (h *Handler) ListOrders(ctx context.Context, req *api.ListOrdersRequest) (*api.OrderList, error) {
	orders, err := h.app.ListOrders(ctx, req.Status)
	if err != nil {
		return nil, mapError(err)
	}
	list := make([]*api.Order, 0, len(orders))
	for _, o := range orders {
		list = append(list, toProto(o))
	}
	return &api.OrderList{Orders: list}, nil
}

func mapError(err error) error {
	if k, ok := err.(*errors.KitError); ok {
		switch k.Code {
		case errors.CodeNotFound:
			return status.Error(codes.NotFound, k.Message)
		case errors.CodeInvalidInput:
			return status.Error(codes.InvalidArgument, k.Message)
		case errors.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, k.Message)
		case errors.CodeForbidden:
			return status.Error(codes.PermissionDenied, k.Message)
		}
		return status.Error(codes.Internal, err.Error())
	}
	return err
}

func toProto(o *domain.Order) *api.Order {
	return &api.Order{
		Id:           o.ID,
		TenantId:     o.TenantID,
		CustomerName: o.CustomerName,
		Status:       o.Status,
		CreatedAt:    o.CreatedAt.Unix(),
	}
}
