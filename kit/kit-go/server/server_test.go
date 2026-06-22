package server

import (
	"context"
	"net"
	"testing"

	"github.com/plantx/kit/kit-go/auth"
	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/log"
	"github.com/plantx/kit/kit-go/provider/stub"
	"github.com/plantx/kit/kit-go/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type testServer struct{}

func (*testServer) Call(_ context.Context, _ *Empty) (*Empty, error) {
	return &Empty{}, nil
}

// Empty is a minimal proto message.
type Empty struct{}

func (*Empty) Reset()         {}
func (*Empty) String() string { return "Empty" }
func (*Empty) ProtoMessage()  {}

// TestService registration helpers.
type TestServiceServer interface {
	Call(context.Context, *Empty) (*Empty, error)
}

func RegisterTestServiceServer(s *grpc.Server, srv TestServiceServer) {
	s.RegisterService(&TestServiceServiceDesc, srv)
}

var TestServiceServiceDesc = grpc.ServiceDesc{
	ServiceName: "TestService",
	HandlerType: (*TestServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(Empty)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(TestServiceServer).Call(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/TestService/Call",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(TestServiceServer).Call(ctx, req.(*Empty))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "test.proto",
}

func startTestServer(t *testing.T, opts Options) (*grpc.ClientConn, func()) {
	t.Helper()
	if opts.Logger == nil {
		opts.Logger = log.FromContext(context.Background())
	}
	srv := New(opts)
	RegisterTestServiceServer(srv.GRPC(), &testServer{})

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		if err := srv.GRPC().Serve(lis); err != nil {
			t.Logf("server serve: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	return conn, func() {
		_ = conn.Close()
		srv.GRPC().Stop()
	}
}

func TestMissingTokenReturns401(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{"ok": {ID: "u1"}}}
	conn, cleanup := startTestServer(t, Options{Authenticator: authn})
	defer cleanup()

	client := &testClient{cc: conn}
	_, err := client.Call(context.Background(), &Empty{})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestInvalidTokenReturns401(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{"ok": {ID: "u1"}}}
	conn, cleanup := startTestServer(t, Options{Authenticator: authn})
	defer cleanup()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "bad-token"))
	client := &testClient{cc: conn}
	_, err := client.Call(ctx, &Empty{})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestUnauthorizedPermissionReturns403(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{"ok": {ID: "u1", TenantID: "t1"}}}
	conn, cleanup := startTestServer(t, Options{
		Authenticator:  authn,
		Authorizer:     stub.DenyAll(),
		TenantResolver: tenant.NewResolver(),
	})
	defer cleanup()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "ok"))
	client := &testClient{cc: conn}
	_, err := client.Call(ctx, &Empty{})
	if s, ok := status.FromError(err); !ok || s.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestTenantContextPropagated(t *testing.T) {
	authn := &stub.Authenticator{Tokens: map[string]*auth.UserInfo{
		"ok": {ID: "u1", TenantID: "t1", Claims: map[string]string{"tenant_id": "t1"}},
	}}
	srv := New(Options{
		Authenticator:  authn,
		Authorizer:     stub.AllowAll(),
		TenantResolver: tenant.NewResolver(),
		Logger:         log.FromContext(context.Background()),
	})
	var captured string
	RegisterTestServiceServer(srv.GRPC(), &tenantCapturingServer{captured: &captured})

	lis := bufconn.Listen(1024 * 1024)
	go func() { _ = srv.GRPC().Serve(lis) }()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	defer srv.GRPC().Stop()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "ok"))
	client := &testClient{cc: conn}
	_, err = client.Call(ctx, &Empty{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured != "t1" {
		t.Fatalf("expected tenant t1, got %s", captured)
	}
}

type tenantCapturingServer struct {
	captured *string
}

func (s *tenantCapturingServer) Call(ctx context.Context, _ *Empty) (*Empty, error) {
	*s.captured = kitctx.GetTenant(ctx).ID
	return &Empty{}, nil
}

type testClient struct {
	cc *grpc.ClientConn
}

func (c *testClient) Call(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/TestService/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
