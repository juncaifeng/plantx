## ADDED Requirements

### Requirement: Portal exposes admin navigation
`apps/portal` SHALL render an admin section in the navigation when the authenticated user has the `admin` role or `platform:admin` permission.

#### Scenario: Admin user sees admin menu
- **WHEN** a user with admin permissions logs in
- **THEN** the portal shows an "Admin" menu with items: Tenants, IAM, Gateway, Audit

### Requirement: Portal loads admin micro-apps
`apps/portal` SHALL register and lazily load the four admin micro-apps under `/admin/tenants`, `/admin/iam`, `/admin/gateway`, and `/admin/audit`.

#### Scenario: Open tenant admin
- **WHEN** the admin clicks "Admin > Tenants"
- **THEN** the portal loads `/apps/tenant-admin-ui/` and mounts it in the main content area

#### Scenario: Non-admin user cannot see admin menu
- **WHEN** a regular business user logs in
- **THEN** the admin menu is hidden
