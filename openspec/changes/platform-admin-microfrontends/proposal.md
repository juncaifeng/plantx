## Why

The current PlantX portal only hosts the `order-ui` business micro-app. There is no admin center for platform operators to manage registered services, gateway routes, tenants, users, roles, or audit logs. Without these, the platform cannot be operated as a true Kit platform.

## What Changes

- Implement minimal backend CRUD services for `platform/tenant-service`, `platform/iam-service`, and a new `platform/gateway-service` (service registry + route management).
- Implement `platform/audit-service` to collect and query operation logs emitted by kit interceptors.
- Create one admin micro-frontend per platform service:
  - `platform/tenant-service/web/tenant-admin-ui`
  - `platform/iam-service/web/iam-admin-ui`
  - `platform/gateway-service/web/gateway-admin-ui`
  - `platform/audit-service/web/audit-admin-ui`
- Add an admin menu section in `apps/portal` that loads these micro-apps under `/admin/...` routes.
- Add REST gateway annotations to each new service so the admin UIs can call them through the nginx gateway.
- Extend the Docker Compose stack and smoke test to cover service registration and route listing.

## Capabilities

### New Capabilities

- `tenant-admin`: Manage tenants (list, create, view) via `tenant-service` and `tenant-admin-ui`.
- `iam-admin`: Manage users, roles, and permissions via `iam-service` and `iam-admin-ui`.
- `gateway-admin`: Register services and declare HTTP routes in `gateway-service`; routes drive the nginx gateway configuration.
- `audit-admin`: View operation logs collected by `audit-service`.
- `portal-admin-menu`: Admin navigation and micro-app loader in the portal shell.

### Modified Capabilities

- `console-portal`: Add an admin section to the navigation and support loading multiple micro-app categories.

## Impact

- New Go modules: `platform/gateway-service` and updates to existing platform stubs.
- New frontend packages: four admin micro-apps plus shared admin components.
- `apps/portal` gains an admin menu and dynamic micro-app registration.
- `deployments/docker-compose` adds platform services and their admin UIs.
- Smoke test extended to verify service registration and route listing.
