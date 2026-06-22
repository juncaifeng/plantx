## ADDED Requirements

### Requirement: Tenant service exposes tenant CRUD APIs
`platform/tenant-service` SHALL expose gRPC/REST APIs to list and create tenants, scoped to platform admin permissions.

#### Scenario: List tenants
- **WHEN** an authenticated platform admin calls `GET /api/tenant/v1/tenants`
- **THEN** the service returns a JSON list of tenants

#### Scenario: Create tenant
- **WHEN** an authenticated platform admin calls `POST /api/tenant/v1/tenants` with `{ "name": "Tenant A" }`
- **THEN** the service creates and returns a tenant with a generated id

### Requirement: Tenant admin UI manages tenants
`tenant-admin-ui` SHALL provide a page to list and create tenants using Ant Design components.

#### Scenario: View tenant list
- **WHEN** the admin navigates to `/admin/tenants`
- **THEN** the tenant-admin micro-app mounts and displays a table of tenants

#### Scenario: Create tenant from UI
- **WHEN** the admin fills the tenant name and clicks "Create"
- **THEN** the UI calls the tenant service API and refreshes the list
