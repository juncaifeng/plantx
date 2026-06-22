package grpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/plantx/kit/kit-go/auth"
	"github.com/plantx/kit/kit-go/authz"
	"github.com/plantx/kit/kit-go/provider/stub"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/plantx/services/order/api"
	"github.com/plantx/services/order/internal/app"
	"github.com/plantx/services/order/internal/infra/repo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func startTestServer(t *testing.T, authenticator auth.Authenticator, authorizer authz.Authorizer) (api.OrderServiceClient, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	_ = lis.Close()

	repository := repo.NewTenantRepo(repo.NewInMemoryRepo())
	orderApp := app.NewOrderService(repository)
	handler := NewHandler(orderApp)

	srv := server.New(server.Options{
		GRPCPort:       port,
		HTTPPort:       0,
		Authenticator:  authenticator,
		Authorizer:     authorizer,
		TenantResolver: tenant.NewResolver(),
	})
	api.RegisterOrderServiceServer(srv.GRPC(), handler)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx)
	}()

	// wait briefly for server to start
	time.Sleep(50 * time.Millisecond)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		cancel()
		t.Fatalf("dial: %v", err)
	}
	client := api.NewOrderServiceClient(conn)

	cleanup := func() {
		_ = conn.Close()
		cancel()
		select {
		case <-errCh:
		case <-time.After(time.Second):
		}
	}
	return client, cleanup
}

func TestCreateOrderMissingToken(t *testing.T) {
	authn := &stub.Authenticator{}
	client, cleanup := startTestServer(t, authn, nil)
	defer cleanup()

	_, err := client.CreateOrder(context.Background(), &api.CreateOrderRequest{CustomerName: "Alice"})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestCreateOrderForbidden(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{
		"user-token": {ID: "u_001", TenantID: "t_001", Username: "alice"},
	}}
	authz := &stub.Authorizer{AllowFunc: func(context.Context, authz.Request) bool { return false }}
	client, cleanup := startTestServer(t, authn, authz)
	defer cleanup()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "user-token"))
	_, err := client.CreateOrder(ctx, &api.CreateOrderRequest{CustomerName: "Alice"})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.PermissionDenied {
		t.Fatalf("expected permission denied, got %v", err)
	}
}

func TestTenantIsolation(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{
		"tenant-a": {ID: "u_a", TenantID: "t_a", Username: "alice", Permissions: []string{"order:order:create", "order:order:read"}},
		"tenant-b": {ID: "u_b", TenantID: "t_b", Username: "bob", Permissions: []string{"order:order:read"}},
	}}
	authz := &stub.Authorizer{AllowFunc: func(_ context.Context, req authz.Request) bool {
		want := req.Action.Service + ":" + req.Action.Resource + ":" + req.Action.Operation
		for _, p := range req.User.Permissions {
			if p == want {
				return true
			}
		}
		return false
	}}
	client, cleanup := startTestServer(t, authn, authz)
	defer cleanup()

	ctxA := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "tenant-a"))
	created, err := client.CreateOrder(ctxA, &api.CreateOrderRequest{CustomerName: "Alice"})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	ctxB := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "tenant-b"))
	_, err = client.GetOrder(ctxB, &api.GetOrderRequest{Id: created.Id})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "not found") {
		t.Fatalf("expected not found for cross-tenant read, got %v", err)
	}

	got, err := client.GetOrder(ctxA, &api.GetOrderRequest{Id: created.Id})
	if err != nil {
		t.Fatalf("get own order: %v", err)
	}
	if got.TenantId != "t_a" {
		t.Fatalf("expected tenant t_a, got %s", got.TenantId)
	}
}

func TestTenantContextPropagation(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{
		"tenant-a": {ID: "u_a", TenantID: "t_a", Username: "alice", Permissions: []string{"*:*:*"}},
	}}
	authz := stub.AllowAll()
	client, cleanup := startTestServer(t, authn, authz)
	defer cleanup()

	ctxA := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "tenant-a"))
	created, err := client.CreateOrder(ctxA, &api.CreateOrderRequest{CustomerName: "Alice"})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if created.TenantId != "t_a" {
		t.Fatalf("expected tenant t_a in response, got %s", created.TenantId)
	}
}
