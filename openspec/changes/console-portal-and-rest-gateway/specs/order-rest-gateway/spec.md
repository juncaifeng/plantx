## ADDED Requirements

### Requirement: Proto defines HTTP mappings
The `services/order/api/order.proto` file SHALL include `google.api.http` annotations mapping each RPC to a REST path and method.

#### Scenario: Proto annotations present
- **WHEN** inspecting `order.proto`
- **THEN** `CreateOrder` maps to `POST /v1/orders`, `GetOrder` maps to `GET /v1/orders/{id}`, and `ListOrders` maps to `GET /v1/orders`

### Requirement: Generated grpc-gateway handler serves REST requests
`order-service` SHALL generate and register a grpc-gateway handler that translates the annotated HTTP requests into gRPC calls on the local `OrderService`.

#### Scenario: List orders via REST
- **WHEN** an authenticated client sends `GET /api/order/v1/orders` with a valid bearer token
- **THEN** the service returns a JSON list of orders scoped to the caller tenant

#### Scenario: Create order via REST
- **WHEN** an authenticated client sends `POST /api/order/v1/orders` with `{ "customer_name": "Alice" }`
- **THEN** the service creates and returns a JSON order with status `pending` and tenant matching the token

#### Scenario: Get order via REST
- **WHEN** an authenticated client sends `GET /api/order/v1/orders/{id}` for an existing order
- **THEN** the service returns the order JSON

### Requirement: REST endpoints enforce auth and authorization
The grpc-gateway handler SHALL run the same authentication and OPA authorization interceptors as native gRPC calls.

#### Scenario: Request without token
- **WHEN** a client sends `GET /api/order/v1/orders` without an `Authorization` header
- **THEN** the service returns HTTP 401 Unauthorized

#### Scenario: Cross-tenant access denied
- **WHEN** a user from tenant B requests an order belonging to tenant A
- **THEN** the service returns HTTP 404 Not Found

### Requirement: Nginx gateway exposes REST routes
The nginx gateway SHALL route `/api/order/` to the REST HTTP port of `order-service`.

#### Scenario: Gateway routing
- **WHEN** a browser sends `GET /api/order/v1/orders` through `http://localhost`
- **THEN** the request reaches `order-service` and returns the tenant-scoped order list
