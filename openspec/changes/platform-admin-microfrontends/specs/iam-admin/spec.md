## ADDED Requirements

### Requirement: IAM service exposes user and role APIs
`platform/iam-service` SHALL expose gRPC/REST APIs to list users and roles and assign roles to users.

#### Scenario: List users
- **WHEN** an authenticated platform admin calls `GET /api/iam/v1/users`
- **THEN** the service returns a JSON list of users

#### Scenario: Create user
- **WHEN** an authenticated platform admin calls `POST /api/iam/v1/users` with `{ "username": "bob", "tenant_id": "t_001" }`
- **THEN** the service creates and returns the user

#### Scenario: List roles
- **WHEN** an authenticated platform admin calls `GET /api/iam/v1/roles`
- **THEN** the service returns a JSON list of roles and their permissions

### Requirement: IAM admin UI manages users and roles
`iam-admin-ui` SHALL provide pages to list users, list roles, and assign roles.

#### Scenario: View users
- **WHEN** the admin navigates to `/admin/iam/users`
- **THEN** the iam-admin micro-app displays a user table

#### Scenario: View roles
- **WHEN** the admin navigates to `/admin/iam/roles`
- **THEN** the UI displays roles and their permissions
