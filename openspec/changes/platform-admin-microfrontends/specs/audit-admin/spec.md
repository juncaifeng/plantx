## ADDED Requirements

### Requirement: Audit service collects operation logs
`platform/audit-service` SHALL expose a gRPC/REST API to query operation logs emitted by kit server interceptors.

#### Scenario: Query audit logs
- **WHEN** an authenticated platform admin calls `GET /api/audit/v1/logs?tenant_id=t_001`
- **THEN** the service returns a paginated list of operation logs

### Requirement: Audit admin UI displays logs
`audit-admin-ui` SHALL provide a page to filter and view audit logs.

#### Scenario: View logs
- **WHEN** the admin navigates to `/admin/audit/logs`
- **THEN** the audit-admin micro-app displays a paginated log table
