# Kit Platform Team Guide

This guide is for platform maintainers who operate and extend the PlantX Kit runtime.

## Kit Runtime Abstractions

| Capability | Interface | Default Provider | Replaceable |
|---|---|---|---|
| Authentication | `auth.Authenticator` | `auth/maxkey` (OIDC/JWT) | Yes |
| Authorization | `authz.Authorizer` | `authz/opa` (Rego/HTTP) | Yes |
| Event Bus | `event.Bus` | `event/nats` (JetStream) | Yes |
| Database | `db.DB` | `db/postgres` | Yes |
| Logger | `log.Logger` | `log/zap` | Yes |
| Config | `config.Loader` | `config/env` | Yes |
| Discovery | `discovery.Registry` | `discovery/static`, `discovery/k8s` | Yes |

## Adding a Provider

1. Create a package under `kit/kit-go/<capability>/<name>`.
2. Implement the interface.
3. Add an env-aware constructor.
4. Update `services/<demo>/cmd/main.go` wiring example.

## Server Interceptors

The Kit server chains:

1. Recovery
2. Trace context extraction
3. Structured logging
4. Metrics
5. Authentication
6. Tenant resolution
7. Authorization (via proto annotations)

## Security Defaults

- All gRPC methods annotated with `plantx.kit.authz.action` require a matching permission or role.
- Tenant ID is extracted from the OIDC token and injected into the request context.
- Repository wrappers enforce row-level tenant isolation.

## Observability

- Health: `/health`
- Readiness: `/ready`
- Metrics: `/metrics` (Prometheus)
- Traces: OpenTelemetry stdout exporter by default; set `OTEL_EXPORTER_OTLP_ENDPOINT` for OTLP.
