# PlantX Demo App

This directory contains a minimal end-to-end demo showing how to build a business micro-service and micro-frontend on PlantX **without reimplementing kit-layer responsibilities**.

## What It Demonstrates

- **Backend (`backend/`)**: A Go gRPC + grpc-gateway service built on `kit/kit-go`.
  - Uses `server.New` from kit-go (logging, interceptors, readiness).
  - Uses `gateway.AutoRegister` to register itself, an application, and a micro-app with `registry-service` automatically.
  - Declares authz actions in proto so kit-go enforces permissions.
  - Uses `kitctx.GetTenant(ctx)` for tenant isolation instead of parsing headers manually.
  - Does **not** implement authentication, authorization, service discovery, or route registration itself.

- **Frontend (`frontend/`)**: A React qiankun micro-app built on `@plantx/kit-sdk-*` and `@plantx/kit-ui`.
  - Receives user, tenant, permissions, and `apiClient` from the PlantX portal via qiankun props.
  - Uses `KitProvider` from `@plantx/kit-sdk-kit` to inject platform context.
  - Uses `useKitUser`, `useKitTenant`, and `useKitPermission` for UI decisions.
  - Uses `KitLayout` from `@plantx/kit-ui` for consistent layout.
  - Calls backend endpoints through the shared `KitApiClient` (token, 401 handling, base URL are all managed by kit).
  - Does **not** implement login, token storage, or permission checking logic itself.

## Directory Structure

```text
demo_app/
├── backend/
│   ├── api/
│   │   └── demo.proto          # gRPC + HTTP + authz annotations
│   ├── cmd/
│   │   └── main.go             # service entry point
│   ├── internal/
│   │   ├── app/                # use cases
│   │   ├── domain/             # domain model
│   │   ├── infra/repo/         # in-memory repository
│   │   └── interfaces/grpc/    # gRPC handler
│   ├── Dockerfile
│   └── go.mod
└── frontend/
    ├── src/
    │   ├── index.tsx           # qiankun lifecycle exports + KitProvider
    │   ├── main.tsx            # standalone dev entry point
    │   ├── DemoPage.tsx        # layout + navigation
    │   ├── DemoHome.tsx        # items page
    │   ├── ConfigManagement.tsx # settings CRUD page
    │   └── SystemSettings.tsx  # admin-only page
    ├── package.json
    ├── tsconfig.json
    └── vite.config.ts
```

## Kit vs. Business Boundaries

This demo intentionally uses kit for every cross-cutting concern and only implements business-specific behavior.

**Kit-layer concerns (not reimplemented)**:

- Authentication (JWT validation, token parsing)
- Authorization (permission enforcement from proto annotations)
- Tenant/user context propagation
- Service/application/micro-app registration
- gRPC/HTTP server plumbing, interceptors, metrics
- API client token injection and 401 handling
- Micro-app lifecycle (qiankun bootstrap/mount/unmount)

**Business-layer concerns (implemented in demo)**:

- Item and Setting domain models
- In-memory persistence
- Setting scope rules (global vs. tenant) — this is the ABAC demo
- CRUD UI for items and settings

**One platform setup step**: `cmd/main.go` seeds demo menus via the registry-service gRPC API after `gateway.AutoRegister` has created the application. This is not reimplementing kit; it is calling the platform registry API to bootstrap menu entries that the portal will render.

## How It Uses Kit

### Backend

| Concern | Kit Responsibility | Demo Code |
|---------|-------------------|-----------|
| gRPC/HTTP server | `kit-go/server` | `cmd/main.go` |
| Service/application/micro-app registration | `kit-go/gateway.AutoRegister` | `cmd/main.go` |
| Authentication interceptor | `kit-go/server` auth interceptor | configured by `server.Options` |
| Authorization interceptor | `kit-go/server` authz interceptor + proto annotations | `api/demo.proto` |
| Tenant resolution | `kit-go/tenant` + `kit-go/context` | `internal/app/demo_service.go` |

### Frontend

| Concern | Kit Responsibility | Demo Code |
|---------|-------------------|-----------|
| API client (token, base URL, 401) | `@plantx/kit-sdk-api` `createClient` | `main.tsx`, portal-provided in `index.tsx` |
| User/tenant/permissions context | `@plantx/kit-sdk-kit` `KitProvider` | `index.tsx` |
| Permission check | `@plantx/kit-sdk-kit` `useKitPermission` | `DemoPage.tsx` |
| Layout component | `@plantx/kit-ui` `KitLayout` | `DemoPage.tsx` |

## Local Development

### 1. Start the Platform

Make sure the PlantX platform is running (e.g. via Docker Compose). At minimum you need:

- `registry-service`
- `iam-service`
- `gateway-service` (for registry management APIs)
- nginx / apisix as the API entry point

The backend must be able to reach `registry-service:8080` (or set `REGISTRY_SERVICE_GRPC_ADDR`).

### 2. Generate Backend Code from Proto

The generated `*.pb.go` files are gitignored. Generate them first:

```bash
cd E:/git/plantx
protoc \
  --proto_path=proto \
  --proto_path=demo_app/backend/api \
  --go_out=demo_app/backend/api \
  --go_opt=paths=source_relative \
  --go-grpc_out=demo_app/backend/api \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=demo_app/backend/api \
  --grpc-gateway_opt=paths=source_relative \
  demo_app/backend/api/demo.proto
```

### 3. Run the Backend

```bash
cd demo_app/backend
go run ./cmd
```

The service will:

- Listen on gRPC `:8080` and HTTP `:8081`.
- Register `demo-service` with `registry-service`.
- Register the `demo` application and `demo-ui` micro-app.

### 4. Run the Frontend

```bash
cd demo_app/frontend
pnpm dev
```

For standalone development, `main.tsx` creates a mock `KitProvider` with a demo user/tenant.

### 5. Build the Frontend for the Portal

```bash
cd demo_app/frontend
pnpm build
```

Output is `dist/demo-ui.iife.js`. Serve it at `/apps/demo-ui/demo-ui.js` so the portal can load it as a qiankun micro-app.

## Verifying Service Registration

Once the backend is running, query the registry:

```bash
# List registered services
curl /api/registry/v1/services

# List registered micro-apps
curl /api/registry/v1/micro-apps

# List applications
curl /api/registry/v1/applications
```

You should see `demo-service`, `demo-ui`, and the `demo` application.

## Verifying Permissions

The demo backend declares two authz actions in `api/demo.proto`:

- `item:create` → required to show the Create button
- `item:list` → required to show items (and also set as `RequirePermission` for the micro-app)

In `iam-service`, create a role with permissions `item:list` and `item:create`, assign it to a user, and log in through the PlantX portal. The demo UI will reflect the user's permissions automatically via `useKitPermission`.

## Important Notes

- The backend repository is in-memory. Data is lost on restart.
- The frontend uses the generic `KitApiClient` because `@plantx/kit-sdk-api` does not yet export a generated demo service client.
- This demo is intentionally minimal. Production services should use Postgres, real logging, tracing, and structured error handling.
