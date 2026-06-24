---
name: sdk-usage-guide
description: >
  Guides business developers on using PlantX SDKs and integrating services into the platform.
  Covers frontend usage of @plantx/kit-sdk-api, @plantx/kit-sdk-kit, and @plantx/kit-ui,
  backend usage of kit-go, service registration via registry-service, micro-frontend registration,
  permission declarations, and frontend-backend integration. Use when developers need to build
  business services or micro-frontends on PlantX.
metadata:
  author: PlantX Platform Team
  version: "1.2"
  updated: "2026-06-24"
---

# PlantX SDK Usage Guide

This skill guides business developers through building on the PlantX platform.

## 1. Platform Architecture

PlantX is a multi-tenant micro-frontend + microservices platform:

- **Platform services**
  - `iam-service`: users, roles, permissions
  - `tenant-service`: tenant management
  - `registry-service`: service, application, micro-app, and menu registration
  - `gateway-service`: legacy/management registry API (not the request proxy)
  - `audit-service`: audit logs
  - `notification-service`: notifications
- **Business services**: e.g. `services/order`, exposed via gRPC + grpc-gateway
- **Frontend SDKs**: `@plantx/kit-sdk-api`, `@plantx/kit-sdk-kit`, `@plantx/kit-ui`
- **Backend SDK**: `github.com/plantx/kit/kit-go`
- **API entry point**: nginx / apisix (deployed via `deployments/nginx/`)

## 2. SDK Packages

| Package | Language/Framework | Purpose |
|---------|-------------------|---------|
| `@plantx/kit-sdk-api` | TypeScript | HTTP clients for platform services (registry, iam, tenant, gateway, audit) |
| `@plantx/kit-sdk-kit` | React | React context and hooks built on `kit-sdk-api`, including `useKitUser`, `useKitTenant`, `useKitPermission`, `useApplications`, `useMenus`, `useMicroApps`. `useMenus`/`useMicroApps` filter online resources by default |
| `@plantx/kit-ui` | React | Minimal placeholder UI components (`KitLayout`, `UserMenu`). Depends on antd but does not yet use antd components internally |
| `github.com/plantx/kit/kit-go` | Go | Server framework with auth, authz, db, events, and gateway registration. `gateway.AutoRegister` supports `WithApplication`, `WithMicroApp`, and `WithMenu` |

## 3. Frontend Development

### 3.1 Install

```bash
npm install @plantx/kit-sdk-api @plantx/kit-sdk-kit @plantx/kit-ui
# or
pnpm add @plantx/kit-sdk-api @plantx/kit-sdk-kit @plantx/kit-ui
```

### 3.2 Initialize API Client

```typescript
import { createClient } from '@plantx/kit-sdk-api';
import { RegistryServiceClient, IAMServiceClient } from '@plantx/kit-sdk-api/registry';

const apiClient = createClient({
  baseURL: 'https://api.plantx.example.com',
  getToken: () => localStorage.getItem('access_token'),
  onUnauthorized: () => {
    localStorage.removeItem('access_token');
    window.location.href = '/login';
  },
});

const registry = new RegistryServiceClient(apiClient);
const iam = new IAMServiceClient(apiClient);
```

`baseURL` points to the platform API gateway (nginx/apisix). All platform and business endpoints are exposed under `/api/<service>/v1/...`.

### 3.3 Call Platform APIs

```typescript
import { TenantServiceClient } from '@plantx/kit-sdk-api/tenant';
import { AuditServiceClient } from '@plantx/kit-sdk-api/audit';

const tenantClient = new TenantServiceClient(apiClient);
const { tenants } = await tenantClient.listTenants();

const auditClient = new AuditServiceClient(apiClient);
const { logs } = await auditClient.listAuditLogs({
  tenant_id: currentTenantId,
  limit: 20,
});
```

### 3.4 React Kit Context

```tsx
import { KitProvider } from '@plantx/kit-sdk-kit';
import type { KitContextValue } from '@plantx/kit-sdk-kit';

const value: KitContextValue = {
  user: {
    id: 'u-123',
    username: 'alice',
    displayName: 'Alice',
    roles: ['admin'],
    permissions: ['order:read', 'order:write'],
  },
  tenant: { id: 't-1', name: 'Default' },
  apiClient,
};

function App() {
  return (
    <KitProvider value={value}>
      <Router />
    </KitProvider>
  );
}
```

### 3.5 Permission Hook

```tsx
import { useKitPermission } from '@plantx/kit-sdk-kit';

function OrderPage() {
  const canWrite = useKitPermission('order:write');
  return (
    <div>
      <h1>Order Management</h1>
      {canWrite && <button>Create Order</button>}
    </div>
  );
}
```

### 3.6 UI Components

```tsx
import { KitLayout, UserMenu } from '@plantx/kit-ui';

function Layout({ children }: { children: React.ReactNode }) {
  return (
    <KitLayout title="Order Center" user={{ displayName: 'Alice' }}>
      <UserMenu onLogout={() => { /* ... */ }} />
      {children}
    </KitLayout>
  );
}
```

`@plantx/kit-ui` currently provides minimal placeholder components. antd is available as a dependency for future enhancement.

### 3.7 Data-Fetching Hooks

`@plantx/kit-sdk-kit` also provides hooks that load platform registry data. By default `useMenus` and `useMicroApps` only return `RESOURCE_STATUS_ONLINE` resources so the UI hides menus and micro-apps whose backend service is offline. Pass `includeOffline: true` when you need to show or manage offline resources (e.g. in an admin console).

```tsx
import {
  useApplications,
  useMenus,
  useMicroApps,
  useRegistryClient,
} from '@plantx/kit-sdk-kit';

function Navigation() {
  const { activeApplications, loading } = useApplications();
  const { menus } = useMenus();
  const { microApps } = useMicroApps({ applicationId: 'demo' });

  // Admin view that also lists offline resources
  const { menus: allMenus } = useMenus({ includeOffline: true });

  if (loading) return <div>Loading...</div>;

  return (
    <nav>
      <h3>Applications</h3>
      <ul>
        {activeApplications.map((app) => (
          <li key={app.key}>{app.name}</li>
        ))}
      </ul>

      <h3>Menus</h3>
      <ul>
        {menus.map((menu) => (
          <li key={menu.id}>{menu.labelKey}</li>
        ))}
      </ul>

      <h3>Micro Apps</h3>
      <ul>
        {microApps.map((app) => (
          <li key={app.name}>{app.name}</li>
        ))}
      </ul>
    </nav>
  );
}
```

These hooks use the `apiClient` from `KitContext` automatically. You can also pass a custom `RegistryServiceClient` as the second argument if you do not want to use context.

### 3.8 Micro-Frontend Registration

A business frontend is usually a qiankun micro-app. The metadata is registered by the backend on startup:

- `name`: unique micro-app name
- `route`: mount route, e.g. `/order`
- `bundleUrl`: JS bundle URL
- `menuLabelKey`: i18n key for the menu label
- `requirePermission`: permission required to access the app

The frontend only needs to build and deploy static assets to the configured `bundleUrl`.

### 3.9 Calling Business Service Endpoints

`@plantx/kit-sdk-api` currently exports clients only for `iam`, `tenant`, `gateway`, `audit`, and `registry`. There is no generated `order` client yet. To call a business service endpoint, use the generic client:

```typescript
const orders = await apiClient.get<{ orders: any[] }>('/api/order/v1/orders');
```

## 4. Backend Development

### 4.1 Service Structure

Reference `services/order/`:

```text
services/<service>/
├── api/                  # Generated gRPC/gateway code
├── cmd/main.go           # Entry point
├── internal/
│   ├── app/              # Use cases
│   ├── domain/           # Domain models/interfaces
│   ├── infra/            # Infrastructure (db, events, repos, sqlc)
│   └── interfaces/       # gRPC handlers
├── migrations/           # DB migrations (optional)
├── Dockerfile
├── go.mod
└── go.sum
```

sqlc queries live in `internal/infra/sqlc/`.

### 4.2 Start a Service with kit-go

```go
package main

import (
    "context"
    "log"

    "github.com/plantx/kit/kit-go/config/env"
    "github.com/plantx/kit/kit-go/gateway"
    "github.com/plantx/kit/kit-go/server"
    "github.com/plantx/services/order/api"
    "github.com/plantx/services/order/internal/app"
    "github.com/plantx/services/order/internal/infra/repo"
    grpcsrv "github.com/plantx/services/order/internal/interfaces/grpc"
    "google.golang.org/grpc/reflection"
)

func main() {
    cfg := env.New("ORDER")
    logger, _ := zaplog.New()
    repository := repo.NewInMemoryRepo()
    orderApp := app.NewOrderService(repository)

    srv := server.New(server.Options{
        ServiceName: "order-service",
        GRPCPort:    8080,
        HTTPPort:    8081,
        Logger:      logger,
        GatewayRegistrar: gateway.AutoRegister("order-service"),
    })

    handler := grpcsrv.NewHandler(orderApp)
    api.RegisterOrderServiceServer(srv.GRPC(), handler)
    reflection.Register(srv.GRPC())

    if err := srv.RegisterGateway(context.Background(), api.RegisterOrderServiceHandler); err != nil {
        log.Fatal(err)
    }
    if err := srv.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### 4.3 Register Service with Platform

Use `gateway.AutoRegister` to register with `registry-service` on startup. A service can declare its application, micro-frontend manifest, and portal menus in one place:

```go
gateway.AutoRegister("order-service",
    gateway.WithApplication(gateway.Application{
        Key:       "order",
        Name:      "Order",
        LabelKey:  "nav.orders",
        Icon:      "ShoppingCartOutlined",
        SortOrder: 10,
        // Status defaults to ACTIVE if omitted
    }),
    gateway.WithMicroApp(gateway.MicroApp{
        Name:              "order-ui",
        Route:             "/order",
        BundleURL:         "/apps/order-ui/order-ui.js",
        MenuLabelKey:      "nav.orders",
        RequirePermission: "order:read",
    }),
    gateway.WithMenu(gateway.Menu{
        LabelKey:          "nav.orders.list",
        Route:             "/order",
        Icon:              "ShoppingCartOutlined",
        SortOrder:         10,
        MicroAppName:      "order-ui",
        RequirePermission: "order:read",
    }),
    gateway.WithMenu(gateway.Menu{
        LabelKey:          "nav.orders.settings",
        Route:             "/order/settings",
        Icon:              "SettingOutlined",
        SortOrder:         20,
        MicroAppName:      "order-ui",
        RequirePermission: "order:admin",
    }),
    gateway.WithGRPCHost("order-service:8080"),
    gateway.WithRESTPrefix("/api/order/v1"),
    gateway.WithRegistryAddr("registry-service:8080"),
)
```

**Defaults**:

- Registry address: env `REGISTRY_SERVICE_GRPC_ADDR`, default `registry-service:8080`
- gRPC host: env `<SERVICE_UPPER>_GRPC_HOST`, default `<service-name>:8080`
- REST prefix: env `<SERVICE_UPPER>_REST_PREFIX`, default `/api/<base>/v1` (strips `-service`)

For `order-service`, the default REST prefix is `/api/order/v1`.

**Idempotency**: menu registration is upserted by `(application_id, label_key, route)`. Restarting a service updates existing menus instead of creating duplicates. The service lifecycle state machine sets menus and micro-apps `OFFLINE` when the service stops and back to `ONLINE` when it starts.

### 4.4 Route Registration

`registry-service` records the service's `rest_prefix` as a wildcard route (`{Path: svc.RestPrefix, Method: "*"}`). Individual REST paths are handled by each service's own grpc-gateway. The platform API gateway (nginx/apisix) uses registry data to route `/api/order/v1/...` to the backend.

### 4.5 Declare Permissions

In the proto file:

```protobuf
import "plantx/kit/authz.proto";

rpc ListOrders(ListOrdersRequest) returns (OrderList) {
  option (google.api.http) = {
    get: "/api/order/v1/orders"
  };
  option (plantx.kit.authz.action) = {
    service: "order"
    resource: "order"
    operation: "list"
  };
}
```

The `kit-go` authz interceptor validates that the caller has the required permission.

### 4.6 Call Platform Services from Go

```go
import (
    "context"
    "github.com/plantx/kit/kit-go/gateway"
)

client, err := gateway.NewClient("registry-service:8080")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

apps, err := client.ListApplications(context.Background())
```

## 5. Permission Naming

Permissions are managed by `iam-service` in the form `<resource>:<operation>`, e.g.:

- `order:read`
- `order:write`
- `order:list`

Proto authz actions (`service=order, resource=order, operation=list`) map to the permission `order:list`. The `service` field is metadata, not part of the permission expression.

Frontend uses `useKitPermission('order:read')` for button-level control.

## 6. Application-Scoped Menus and Micro-Apps

`registry-service` provides APIs to fetch menus and micro-apps for a given application:

- `GET /api/registry/v1/applications/{application_id}/menus`
- `GET /api/registry/v1/applications/{application_id}/micro-apps`

The platform portal uses these endpoints to render navigation and load micro-frontends.

## 7. Frontend-Backend Integration Flow

1. Backend implements gRPC service and declares authz actions.
2. Backend starts with `gateway.AutoRegister`, registering service, application, micro-app, and menus with `registry-service`.
3. The platform API gateway (nginx/apisix) routes `/api/order/v1/...` to the order service based on registry data.
4. Frontend uses `@plantx/kit-sdk-api` to call `/api/order/v1/...` through the API gateway.
5. The portal uses `useMenus` / `useMicroApps` to load online resources and render navigation / load micro-frontends.
6. When the service stops, the Temporal lifecycle workflow marks related menus and micro-apps `OFFLINE`; they become visible again after the service restarts.

## 8. Local Environment Variables

```bash
# Service
ORDER_GRPC_PORT=8080
ORDER_HTTP_PORT=8081
ORDER_DATABASE_DSN=postgres://user:pass@localhost:5432/order?sslmode=disable
ORDER_NATS_URL=nats://localhost:4222
ORDER_TRACING_ENABLED=true

# Auth (MaxKey)
MAXKEY_ISSUER=https://maxkey.example.com
MAXKEY_JWKS_URL=https://maxkey.example.com/.well-known/jwks.json

# Authorization (OPA)
OPA_URL=http://localhost:8181
OPA_DECISION_PATH=v1/data/plantx/authz/allow

# Registry
REGISTRY_SERVICE_GRPC_ADDR=localhost:8080
```

Authentication and authorization are optional at runtime unless the corresponding environment variables are configured.

## 9. Troubleshooting

| Symptom | Cause / Fix |
|---------|-------------|
| 401 from frontend | Check `getToken` returns a valid access token; handle expiration |
| Backend fails to register | Verify `REGISTRY_SERVICE_GRPC_ADDR` and service name uniqueness |
| Micro-app missing from menu | Check `WithApplication` key, `RequirePermission`, and gateway route sync |
| Menu missing after service restart | Menus are upserted by `(application_id, label_key, route)`; verify the menu still has the same key/route and the service lifecycle workflow completed |
| Duplicate menus after restart | Should not happen with menu upsert; check migration `008_menus_upsert` has been applied |
| Permission denied | Verify proto authz action and IAM role assignment |
| Route not proxied | Verify `registry-service` has the service's `rest_prefix` and nginx/apisix has synced routes |

## 10. Related Files

- `kit/kit-sdk-api/src/`
- `kit/kit-sdk-kit/src/`
- `kit/kit-ui/src/`
- `kit/kit-go/gateway/`
- `kit/kit-go/server/`
- `platform/registry-service/api/registry.proto`
- `services/order/cmd/main.go`
- `deployments/nginx/nginx.conf`
- `deployments/nginx/apisix.yaml`
