# Kit Getting Started

This guide helps business developers build a new PlantX service in minutes using the Kit platform.

## Prerequisites

- Go 1.22+
- Node.js 20 + pnpm 9
- Docker and Docker Compose
- protoc / buf

## Create a Service

```bash
kit new service order --ui
```

This scaffolds:

```
services/order
├── api/order.proto
├── internal/domain
├── internal/app
├── internal/infra/repo
├── internal/interfaces/grpc
├── web/order-ui
└── Dockerfile
```

## Define the Contract

Edit `services/order/api/order.proto`:

```proto
service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (Order) {
    option (plantx.kit.authz.action) = { service: "order" resource: "order" operation: "create" };
  }
}
```

## Generate Code

```bash
kit generate
```

This runs buf, sqlc, and frontend SDK generators.

## Implement Business Logic

Write domain models in `internal/domain`, use cases in `internal/app`, and the gRPC handler in `internal/interfaces/grpc`.

## Run Locally

```bash
kit dev up
```

The Kit runtime wires authentication, authorization, tenant context, tracing, and metrics automatically.

## Next Steps

- Read `docs/kit-platform-team-guide.md` to customize infrastructure providers.
- Read `docs/deployment-guide.md` to deploy to K8s, Docker Compose, or bare metal.
