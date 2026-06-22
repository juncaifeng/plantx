package grpc

import (
	"context"

	pb "github.com/plantx/services/demotmp/api"
	"github.com/plantx/services/demotmp/internal/app"
)

type Handler struct {
	app *app.Service
}

func NewHandler(app *app.Service) *Handler {
	return &Handler{app: app}
}

func (h *Handler) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PongResponse, error) {
	return &pb.PongResponse{Message: h.app.Ping(ctx, req.Message)}, nil
}
