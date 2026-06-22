## ADDED Requirements

### Requirement: Gateway service maintains service registry
`platform/gateway-service` SHALL expose gRPC/REST APIs to register services and declare their HTTP route prefixes.

#### Scenario: Register a service
- **WHEN** an authenticated platform admin calls `POST /api/gateway/v1/services` with `{ "name": "order-service", "grpc_host": "order-service:8080", "rest_prefix": "/api/order/" }`
- **THEN** the service records the registration and returns an id

#### Scenario: List registered services
- **WHEN** an authenticated platform admin calls `GET /api/gateway/v1/services`
- **THEN** the service returns all registered services and their route prefixes

#### Scenario: List routes for a service
- **WHEN** an authenticated platform admin calls `GET /api/gateway/v1/services/{id}/routes`
- **THEN** the service returns the declared routes

### Requirement: Gateway admin UI shows services and routes
`gateway-admin-ui` SHALL provide a page to view registered services and their routes.

#### Scenario: View service registry
- **WHEN** the admin navigates to `/admin/gateway/services`
- **THEN** the gateway-admin micro-app displays the registered services and route prefixes

#### Scenario: Register service from UI
- **WHEN** the admin fills the service form and submits
- **THEN** the UI calls the gateway service and updates the list
