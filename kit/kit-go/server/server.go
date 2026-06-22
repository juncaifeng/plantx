package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/plantx/kit/kit-go/auth"
	"github.com/plantx/kit/kit-go/authz"
	authzpb "github.com/plantx/kit/kit-go/proto/authz"
	"github.com/plantx/kit/kit-go/config"
	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/db"
	"github.com/plantx/kit/kit-go/errors"
	"github.com/plantx/kit/kit-go/event"
	"github.com/plantx/kit/kit-go/log"
	"github.com/plantx/kit/kit-go/telemetry"
	"github.com/plantx/kit/kit-go/tenant"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var (
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "kit_grpc_request_duration_seconds",
		Help: "gRPC request latency",
	}, []string{"service", "method", "status"})
	requestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kit_grpc_request_total",
		Help: "gRPC request count",
	}, []string{"service", "method", "status"})
)

func init() {
	prometheus.MustRegister(requestDuration, requestTotal)
}

// GatewayRegistrar abstracts service registration with the platform gateway.
type GatewayRegistrar interface {
	Register(ctx context.Context) error
	Deregister(ctx context.Context) error
}

type TracingOptions struct {
	Enabled     bool
	ServiceName string
}

type Options struct {
	ServiceName         string
	GRPCPort            int
	HTTPPort            int
	Authenticator       auth.Authenticator
	Authorizer          authz.Authorizer
	TenantResolver      tenant.Resolver
	Logger              log.Logger
	Config              config.Loader
	DB                  db.DB
	EventBus            event.Bus
	Tracing             TracingOptions
	Readiness           func(ctx context.Context) error
	GatewayRegistrar    GatewayRegistrar
	AuthExcludedMethods []string
}

type Server struct {
	opts        Options
	grpc        *grpc.Server
	http        *http.Server
	gatewayConn *grpc.ClientConn
}

func New(opts Options) *Server {
	if opts.Logger == nil {
		opts.Logger = log.FromContext(context.Background())
	}
	if opts.Tracing.Enabled {
		if _, err := telemetry.InitTracerProvider(context.Background(), opts.Tracing.ServiceName); err != nil {
			opts.Logger.Error("failed to initialize tracer", log.F("error", err))
		}
	}
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor(opts.Logger),
			traceInterceptor(),
			loggingInterceptor(opts.Logger, opts.EventBus),
			metricsInterceptor(),
			authInterceptor(opts.Authenticator, opts.AuthExcludedMethods, opts.Logger),
			tenantInterceptor(opts.TenantResolver),
			authzInterceptor(opts.Authorizer, opts.AuthExcludedMethods, opts.Logger),
		),
	}
	s := &Server{opts: opts, grpc: grpc.NewServer(grpcOpts...)}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if opts.Readiness != nil {
			if err := opts.Readiness(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(fmt.Sprintf(`{"status":"not ready","error":"%v"}`, err)))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	mux.Handle("/metrics", promhttp.Handler())
	s.http = &http.Server{Addr: fmt.Sprintf(":%d", opts.HTTPPort), Handler: mux}
	return s
}

func (s *Server) GRPC() *grpc.Server { return s.grpc }

func (s *Server) RegisterGateway(ctx context.Context, register func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error) error {
	gwMux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.MD{}
			for k, v := range r.Header {
				if key, ok := runtime.DefaultHeaderMatcher(k); ok {
					md.Append(key, v...)
				}
			}
			if auth := r.Header.Get("Authorization"); auth != "" {
				md.Set("authorization", auth)
			} else {
				s.opts.Logger.Warn("missing authorization header", log.F("path", r.URL.Path))
			}
			return md
		}),
	)
	addr := fmt.Sprintf("localhost:%d", s.opts.GRPCPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("dial gateway grpc %s: %w", addr, err)
	}
	if err := register(ctx, gwMux, conn); err != nil {
		return fmt.Errorf("register gateway: %w", err)
	}
	s.gatewayConn = conn

	const apiPrefix = "/api/"
	base, ok := s.http.Handler.(*http.ServeMux)
	if !ok {
		existing := s.http.Handler
		s.http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, apiPrefix) {
				gwMux.ServeHTTP(w, r)
				return
			}
			existing.ServeHTTP(w, r)
		})
		return nil
	}
	base.Handle(apiPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// http.ServeMux strips "/api" before invoking the handler; restore it
		// so the grpc-gateway patterns (e.g., "/api/order/v1/orders") match.
		if !strings.HasPrefix(r.URL.Path, apiPrefix) {
			r.URL.Path = "/api" + r.URL.Path
			if r.URL.RawPath != "" {
				r.URL.RawPath = "/api" + r.URL.RawPath
			}
		}
		gwMux.ServeHTTP(w, r)
	}))
	return nil
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.opts.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	s.opts.Logger.Info("kit server starting", log.F("grpc_addr", addr), log.F("http_addr", s.http.Addr))
	errCh := make(chan error, 1)
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("http server: %w", err)
		}
	}()
	go func() {
		if err := s.grpc.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()
	if s.opts.GatewayRegistrar != nil {
		if regErr := s.opts.GatewayRegistrar.Register(ctx); regErr != nil {
			s.opts.Logger.Error("failed to register with gateway", log.F("error", regErr))
		} else {
			s.opts.Logger.Info("registered with gateway", log.F("service_name", s.opts.ServiceName))
		}
	}
	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.opts.GatewayRegistrar != nil {
		if err := s.opts.GatewayRegistrar.Deregister(ctx); err != nil {
			s.opts.Logger.Error("failed to deregister from gateway", log.F("error", err))
		}
	}
	s.grpc.GracefulStop()
	if s.gatewayConn != nil {
		_ = s.gatewayConn.Close()
	}
	return s.http.Shutdown(ctx)
}

func recoveryInterceptor(l log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				l.Error("panic recovered", log.F("method", info.FullMethod), log.F("panic", p))
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

func traceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if tp := md.Get("traceparent"); len(tp) > 0 {
				parts := strings.Split(tp[0], "-")
				if len(parts) >= 3 {
					ctx = kitctx.WithTrace(ctx, kitctx.TraceContext{TraceID: parts[1]})
				}
			} else if tid := md.Get("x-trace-id"); len(tid) > 0 {
				ctx = kitctx.WithTrace(ctx, kitctx.TraceContext{TraceID: tid[0]})
			}
		}
		return handler(ctx, req)
	}
}

func loggingInterceptor(l log.Logger, bus event.Bus) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		code := codes.OK
		if s, ok := status.FromError(err); ok {
			code = s.Code()
		}
		fields := []log.Field{
			log.F("method", info.FullMethod),
			log.F("duration_ms", time.Since(start).Milliseconds()),
			log.F("code", code.String()),
		}
		var tenantID, userID string
		if t := kitctx.GetTenant(ctx); t.ID != "" {
			tenantID = t.ID
			fields = append(fields, log.F("tenant_id", t.ID))
		}
		if u := kitctx.GetUser(ctx); u != nil {
			userID = u.ID
			fields = append(fields, log.F("user_id", u.ID))
		}
		if trace := kitctx.GetTrace(ctx); trace.TraceID != "" {
			fields = append(fields, log.F("trace_id", trace.TraceID))
		}
		if err != nil && code != codes.OK {
			l.Error("grpc request", append(fields, log.F("error", err.Error()))...)
		} else {
			l.Info("grpc request", fields...)
		}
		if bus != nil {
			action := resolveAction(info.FullMethod)
			event := &AuditEvent{
				Method:    info.FullMethod,
				Status:    code.String(),
				UserID:    userID,
				TenantID:  tenantID,
				Timestamp: time.Now().UnixMilli(),
			}
			if action != nil {
				event.Action = action.Operation
				event.Resource = action.Resource
			}
			if pubErr := bus.Publish(ctx, event); pubErr != nil {
				l.Error("failed to publish audit event", log.F("error", pubErr.Error()))
			}
		}
		return resp, err
	}
}

func metricsInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		svc, method := splitMethod(info.FullMethod)
		resp, err := handler(ctx, req)
		statusCode := codes.OK.String()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				statusCode = s.Code().String()
			} else {
				statusCode = codes.Unknown.String()
			}
		}
		requestDuration.WithLabelValues(svc, method, statusCode).Observe(time.Since(start).Seconds())
		requestTotal.WithLabelValues(svc, method, statusCode).Inc()
		return resp, err
	}
}

func authInterceptor(a auth.Authenticator, excludedMethods []string, l log.Logger) grpc.UnaryServerInterceptor {
	excluded := make(map[string]struct{}, len(excludedMethods))
	for _, m := range excludedMethods {
		excluded[m] = struct{}{}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := excluded[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		if a == nil {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, errors.CodeUnauthorized)
		}
		token := ""
		for _, k := range []string{"authorization", "Authorization"} {
			if vals := md.Get(k); len(vals) > 0 {
				token = strings.TrimSpace(vals[0])
				break
			}
		}
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			keys := make([]string, 0, len(md))
			for k := range md {
				keys = append(keys, k)
			}
			l.Warn("authentication failed: empty token", log.F("method", info.FullMethod), log.F("md_keys", keys))
			return nil, status.Error(codes.Unauthenticated, errors.CodeUnauthorized)
		}
		user, err := a.Authenticate(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, errors.CodeUnauthorized)
		}
		ctx = kitctx.WithUser(ctx, user)
		if user.TenantID != "" {
			ctx = kitctx.WithTenant(ctx, tenant.Info{ID: user.TenantID})
		}
		return handler(ctx, req)
	}
}

func tenantInterceptor(r tenant.Resolver) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if r == nil {
			return handler(ctx, req)
		}
		user := kitctx.GetUser(ctx)
		if user == nil {
			return handler(ctx, req)
		}
		t, err := r.Resolve(user.ID, user.Claims)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if t.ID != "" {
			ctx = kitctx.WithTenant(ctx, t)
		}
		return handler(ctx, req)
	}
}

func authzInterceptor(z authz.Authorizer, excludedMethods []string, l log.Logger) grpc.UnaryServerInterceptor {
	excluded := make(map[string]struct{}, len(excludedMethods))
	for _, m := range excludedMethods {
		excluded[m] = struct{}{}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := excluded[info.FullMethod]; ok {
			return handler(ctx, req)
		}
		if z == nil {
			return handler(ctx, req)
		}
		action := resolveAction(info.FullMethod)
		if action == nil {
			return handler(ctx, req)
		}
		user := kitctx.GetUser(ctx)
		if user == nil {
			l.Warn("authorization failed: no authenticated user", log.F("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, errors.CodeUnauthorized)
		}
		decision, err := z.Authorize(ctx, authz.Request{User: *user, Action: *action})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !decision.Allowed {
			return nil, status.Error(codes.PermissionDenied, errors.CodeForbidden+": "+decision.Reason)
		}
		return handler(ctx, req)
	}
}

func splitMethod(full string) (service, method string) {
	parts := strings.Split(strings.TrimPrefix(full, "/"), "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return full, ""
}

func resolveAction(fullMethod string) *authz.Action {
	svcName, methodName := splitMethod(fullMethod)
	if svcName == "" || methodName == "" {
		return nil
	}
	// Try annotated action from proto method options.
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(svcName))
	if err == nil {
		if svc, ok := desc.(protoreflect.ServiceDescriptor); ok {
			if method := svc.Methods().ByName(protoreflect.Name(methodName)); method != nil {
				opts := method.Options()
				if proto.HasExtension(opts, authzpb.E_Action) {
					ext := proto.GetExtension(opts, authzpb.E_Action)
					if pb, ok := ext.(*authzpb.Action); ok {
						return &authz.Action{
							Service:   pb.GetService(),
							Resource:  pb.GetResource(),
							Operation: pb.GetOperation(),
						}
					}
				}
			}
		}
	}
	// Fallback: derive a conservative action so an authorizer can still decide.
	return &authz.Action{
		Service:   svcName,
		Resource:  methodName,
		Operation: operationFromMethod(methodName),
	}
}

func operationFromMethod(name string) string {
	switch {
	case strings.HasPrefix(name, "Create"):
		return "create"
	case strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "Read"):
		return "read"
	case strings.HasPrefix(name, "List"):
		return "list"
	case strings.HasPrefix(name, "Update"):
		return "update"
	case strings.HasPrefix(name, "Delete"):
		return "delete"
	default:
		return "execute"
	}
}
