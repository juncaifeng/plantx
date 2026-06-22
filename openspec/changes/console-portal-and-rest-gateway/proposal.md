## Why

The PlantX Kit platform currently has a backend `order-service` and a Qiankun micro-app skeleton (`order-ui`), but there is no unified console portal to host micro-frontends, and `order-service` only exposes gRPC. Business developers consuming the platform need a browser-accessible console and REST entry point for the demo service.

## What Changes

- Build a Qiankun-based **console portal** (`apps/console`) that bootstraps micro-frontends, handles login against the gateway `/oauth/token` endpoint, and passes user/tenant/permissions context to child apps.
- Add a **REST gateway** for `order-service` so the frontend SDK can call `/api/order/v1/orders` through the nginx gateway.
- Update `services/order/web/order-ui` to mount inside the portal and use the shared API client.
- Update `deployments/docker-compose` with a `portal` service and the necessary nginx routes.
- Add an end-to-end browser smoke test (or curl-based equivalent) that logs in through the portal and creates/lists orders.

## Capabilities

### New Capabilities

- `console-portal`: A Qiankun shell application that loads micro-frontends, provides login/logout, top navigation, and propagates auth context.
- `order-rest-gateway`: REST endpoints (`/api/order/v1/*`) mapped to `order-service` gRPC methods via grpc-gateway, wired through the nginx gateway.

### Modified Capabilities

- None.

## Impact

- New `apps/console` package (React + Vite + Qiankun + React Router).
- `services/order/api/order.proto` gains `google.api.http` annotations.
- `services/order/cmd/main.go` registers the grpc-gateway handler on the kit HTTP server.
- `kit/kit-go/server` exposes a hook for registering grpc-gateway muxes.
- `deployments/docker-compose/nginx.conf` routes `/api/order/` and serves the portal at `/`.
- `deployments/docker-compose/docker-compose.yml` adds a `portal` service.
