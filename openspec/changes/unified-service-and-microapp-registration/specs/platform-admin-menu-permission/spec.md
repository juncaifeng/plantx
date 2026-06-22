## ADDED Requirements

### Requirement: Admin menu renders for admin role or platform:admin permission
`apps/portal/src/Layout.tsx` SHALL render the Admin menu section when the user has either the `admin` role or the `platform:admin` permission.

#### Scenario: User has platform:admin but not admin role
- **WHEN** the authenticated user has `permissions: ["platform:admin"]` and no `admin` role
- **THEN** the Admin menu is visible

### Requirement: Admin menu hidden for regular users
If the user has neither `admin` role nor `platform:admin` permission, the Admin menu SHALL not render.

#### Scenario: Regular user logs in
- **WHEN** the user has no admin role and no platform:admin permission
- **THEN** the Admin menu is not visible
