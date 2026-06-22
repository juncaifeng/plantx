## ADDED Requirements

### Requirement: Kit runtime provides authentication abstraction
The Kit runtime SHALL provide an `auth.Authenticator` interface that resolves a JWT or bearer token into a `UserInfo` struct.

#### Scenario: Authenticating a request
- **WHEN** a request reaches a service with an `Authorization` header
- **THEN** the Kit server interceptor calls the configured authenticator and injects `UserInfo` into the request context

### Requirement: Kit runtime provides authorization abstraction
The Kit runtime SHALL provide an `authz.Authorizer` interface and enforce authorization through proto annotations.

#### Scenario: Enforcing a policy
- **WHEN** an RPC is annotated with `(plantx.authz.action)`
- **THEN** the Kit interceptor SHALL call the authorizer before invoking the handler and reject the request if unauthorized

### Requirement: Kit runtime propagates tenant context
The Kit runtime SHALL extract `tenant_id` from authenticated user claims and propagate it through the request context.

#### Scenario: Accessing tenant context
- **WHEN** business code calls `kitctx.GetTenant(ctx)`
- **THEN** it receives the tenant identifier associated with the current request

### Requirement: Kit runtime provides event bus abstraction
The Kit runtime SHALL provide an `event.Bus` interface for publishing and subscribing to domain events without binding to a specific message broker.

#### Scenario: Publishing a domain event
- **WHEN** business code calls `event.Publish(ctx, &OrderCreatedEvent{...})`
- **THEN** the event is serialized and sent through the configured broker with trace and tenant context attached

### Requirement: Kit runtime provides database access helpers
The Kit runtime SHALL provide a database connection pool, transaction helper, and sqlc integration utilities.

#### Scenario: Running a transaction
- **WHEN** application code calls `kitdb.WithTx(ctx, func(tx *sql.Tx) error { ... })`
- **THEN** the function executes within a database transaction and commits or rolls back automatically
