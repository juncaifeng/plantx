## ADDED Requirements

### Requirement: All business tables contain a tenant_id column
Every table created by a business service SHALL include a non-nullable `tenant_id` column.

#### Scenario: Creating a new table
- **WHEN** a developer writes a migration for a new business entity
- **THEN** the migration MUST define `tenant_id` as part of the table schema

### Requirement: Tenant context is automatically injected into queries
The Kit sqlc plugin or repository wrapper SHALL automatically append `tenant_id` filtering to data access operations.

#### Scenario: Listing orders
- **WHEN** business code calls `ListOrdersByStatus(ctx, "pending")`
- **THEN** the executed SQL includes `WHERE tenant_id = <current_tenant> AND status = 'pending'`

### Requirement: Cross-tenant data access is denied by default
No query executed through Kit data access helpers SHALL return data belonging to a different tenant than the one in the current context.

#### Scenario: Tenant isolation enforcement
- **WHEN** a request for tenant A attempts to read tenant B's data
- **THEN** the query returns no rows and the system logs an access attempt

### Requirement: Tenant identifier is propagated through service calls
When one service calls another via gRPC or event bus, the tenant identifier SHALL be carried in metadata or event envelope.

#### Scenario: Service-to-service call
- **WHEN** service A calls service B with `tenant_id = t_001`
- **THEN** service B receives the same `tenant_id` in its request context
