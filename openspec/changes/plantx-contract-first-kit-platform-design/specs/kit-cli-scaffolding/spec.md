## ADDED Requirements

### Requirement: CLI can scaffold a new service
The `kit-cli` tool SHALL provide a `kit new service <name>` command that generates a complete service directory following the DDD structure.

#### Scenario: Creating an order service
- **WHEN** a developer runs `kit new service order --ui`
- **THEN** the directories `api/`, `internal/domain/`, `internal/app/`, `internal/infra/`, `internal/interfaces/`, `migrations/`, `web/`, and `Dockerfile` are created

### Requirement: CLI can generate code from proto and sqlc
The `kit-cli` tool SHALL provide a `kit generate` command that invokes buf, sqlc, and SDK generators for the current service.

#### Scenario: Regenerating after proto change
- **WHEN** a developer modifies `api/order.proto` and runs `kit generate`
- **THEN** Go gRPC code, gateway code, sqlc DAO, and TypeScript SDK are regenerated

### Requirement: CLI can manage local development environment
The `kit-cli` tool SHALL provide `kit dev up`, `kit dev down`, and `kit dev logs` commands to start and stop local dependencies.

#### Scenario: Starting local environment
- **WHEN** a developer runs `kit dev up`
- **THEN** PostgreSQL, NATS, MaxKey, OPA, and the API gateway are started via Docker Compose

### Requirement: CLI can create database migrations
The `kit-cli` tool SHALL provide a `kit migrate new <name>` command that creates a timestamped migration file pair.

#### Scenario: Adding a new table
- **WHEN** a developer runs `kit migrate new add_order_status`
- **THEN** `migrations/<timestamp>_add_order_status.up.sql` and `.down.sql` files are created
