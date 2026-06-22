## ADDED Requirements

### Requirement: Service interfaces are defined in Protobuf
All service APIs, including RPC methods, request/response messages, and events, SHALL be defined in `.proto` files under the service's `api/` directory.

#### Scenario: Creating a new service
- **WHEN** a developer runs `kit new service order`
- **THEN** a scaffolded `api/order.proto` file is generated with a sample service, message, and event definition

#### Scenario: Generating service code
- **WHEN** a developer runs `kit generate`
- **THEN** Go gRPC code, gRPC-Gateway HTTP handlers, and TypeScript SDK types are generated from the proto files

### Requirement: Data access contracts are defined in sqlc SQL files
All database queries and schema migrations SHALL be defined in `.sql` files; sqlc SHALL generate type-safe DAO code from these files.

#### Scenario: Defining a query
- **WHEN** a developer adds a query to `internal/infra/sqlc/queries.sql`
- **THEN** running `kit generate` produces Go functions with compile-time checked SQL and parameter types

#### Scenario: Schema migration
- **WHEN** a developer creates a migration file under `migrations/`
- **THEN** the migration is versioned and applied in order during service startup or `kit dev up`

### Requirement: Proto and sqlc are the single source of truth
No service SHALL expose an API, SDK, or UI type that is not derived from the service's proto or sqlc definitions.

#### Scenario: Frontend API consumption
- **WHEN** a frontend developer imports `{service}-sdk-api`
- **THEN** all request/response types and API methods are generated from the service proto, not hand-written
