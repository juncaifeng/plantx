## ADDED Requirements

### Requirement: kit server interceptor emits audit events
`kit/kit-go/server/server.go` logging interceptor SHALL publish an audit event to the configured `event.Bus` for each gRPC/HTTP request, including `user_id`, `tenant_id`, `method`, `action`, and `timestamp`.

#### Scenario: Authenticated request is handled
- **WHEN** an authenticated request reaches any kit-powered service
- **THEN** an audit event is published to the event bus

### Requirement: audit-service subscribes to audit events
`platform/audit-service` SHALL subscribe to the `audit.events` subject on the event bus and persist each event to its repository.

#### Scenario: Audit event is published
- **WHEN** an audit event is published by a kit service
- **THEN** `audit-service` receives it and stores it in memory (or PostgreSQL) repository

### Requirement: Audit logs are queryable
`audit-service` SHALL expose `ListAuditLogs` filtered by `tenant_id`, `user_id`, and time range.

#### Scenario: Admin queries audit logs
- **WHEN** an admin calls `GET /api/audit/v1/logs?tenant_id=t_001`
- **THEN** the response contains only logs for that tenant
