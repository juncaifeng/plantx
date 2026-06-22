## 1. Tenant Admin

- [x] 1.1 Implement `platform/tenant-service` proto with `ListTenants` and `CreateTenant`
- [x] 1.2 Generate grpc-gateway code for tenant-service
- [x] 1.3 Implement tenant-service in-memory repository and gRPC handler
- [x] 1.4 Register tenant-service REST gateway in `cmd/main.go`
- [x] 1.5 Create `platform/tenant-service/web/tenant-admin-ui` with Antd table and create form
- [x] 1.6 Add tenant-service and tenant-admin-ui to Docker Compose

## 2. IAM Admin

- [x] 2.1 Implement `platform/iam-service` proto with `ListUsers`, `CreateUser`, `ListRoles`
- [x] 2.2 Generate grpc-gateway code for iam-service
- [x] 2.3 Implement iam-service in-memory repository and gRPC handler
- [x] 2.4 Register iam-service REST gateway in `cmd/main.go`
- [x] 2.5 Create `platform/iam-service/web/iam-admin-ui` with users and roles pages
- [x] 2.6 Add iam-service and iam-admin-ui to Docker Compose

## 3. Gateway Admin

- [x] 3.1 Create `platform/gateway-service` module
- [x] 3.2 Implement gateway-service proto with `RegisterService`, `ListServices`, `ListRoutes`
- [x] 3.3 Generate grpc-gateway code for gateway-service
- [x] 3.4 Implement in-memory service registry and gRPC handler
- [x] 3.5 Register gateway-service REST gateway in `cmd/main.go`
- [x] 3.6 Create `platform/gateway-service/web/gateway-admin-ui` with service/registry page
- [x] 3.7 Add gateway-service and gateway-admin-ui to Docker Compose

## 4. Audit Admin

- [x] 4.1 Implement `platform/audit-service` proto with `QueryLogs`
- [x] 4.2 Generate grpc-gateway code for audit-service
- [x] 4.3 Implement audit-service in-memory log store and gRPC handler
- [x] 4.4 Register audit-service REST gateway in `cmd/main.go`
- [x] 4.5 Create `platform/audit-service/web/audit-admin-ui` with log table
- [x] 4.6 Add audit-service and audit-admin-ui to Docker Compose

## 5. Portal Admin Menu

- [x] 5.1 Update `apps/portal/src/Layout.tsx` to conditionally render admin menu based on permissions
- [x] 5.2 Register admin micro-apps in `apps/portal/src/App.tsx` or `OrdersPage.tsx` equivalent
- [x] 5.3 Add `/admin/tenants`, `/admin/iam`, `/admin/gateway`, `/admin/audit` routes
- [x] 5.4 Build admin UIs and copy them into portal dist under `/apps/*-admin-ui/`

## 6. Deployment and Verification

- [x] 6.1 Update `deployments/docker-compose/nginx.conf` to route `/api/tenant/`, `/api/iam/`, `/api/gateway/`, `/api/audit/`
- [x] 6.2 Update `deployments/docker-compose/docker-compose.yml` with platform services and admin UIs
- [x] 6.3 Update portal Dockerfile to build all admin UIs
- [x] 6.4 Extend `scripts/e2e-smoke-test.sh` to verify service registration and tenant listing
- [x] 6.5 Run `docker-compose up -d --build` and verify admin menu loads
