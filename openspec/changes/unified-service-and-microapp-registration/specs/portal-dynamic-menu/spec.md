## ADDED Requirements

### Requirement: Portal uses a manifest to render business routes
`apps/portal/src/App.tsx` SHALL read a `microApps` manifest and render `<Route>` elements for business micro-apps dynamically, without hardcoding each app.

#### Scenario: New business micro-app is added
- **WHEN** a new entry is added to `microApps.ts` with route `/test`
- **THEN** portal renders `/test` without modifying `App.tsx`

### Requirement: Portal uses a manifest to render business menu items
`apps/portal/src/Layout.tsx` SHALL read the same `microApps` manifest and render top-level menu items dynamically.

#### Scenario: New business micro-app appears in menu
- **WHEN** a new entry is added to `microApps.ts` with `menuLabelKey: "nav.test"`
- **THEN** the top navigation shows the translated label without modifying `Layout.tsx`

### Requirement: Menu items respect permissions
If a micro-app manifest specifies `requirePermission`, the menu item SHALL only render when the user has that permission.

#### Scenario: User lacks required permission
- **WHEN** a micro-app requires `test:read` and the user does not have it
- **THEN** the menu item is not rendered
