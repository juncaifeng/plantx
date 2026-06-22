## ADDED Requirements

### Requirement: Portal authenticates users
The portal SHALL obtain an access token from the gateway `/oauth/token` endpoint using the OAuth2 password grant and SHALL use that token for all downstream API calls.

#### Scenario: Successful login
- **WHEN** the user submits valid credentials (`demo-a` / `demo-a`) on the portal login page
- **THEN** the portal receives an `access_token` and navigates to the console home

#### Scenario: Failed login
- **WHEN** the user submits invalid credentials
- **THEN** the portal displays an authentication error and does not store a token

### Requirement: Portal loads micro-frontends
The portal SHALL register and load the `order-ui` Qiankun micro-application at runtime when the user navigates to the Orders menu item.

#### Scenario: Orders menu opened
- **WHEN** the authenticated user clicks the "Orders" menu item
- **THEN** the portal mounts the `order-ui` micro-app inside the main content area

### Requirement: Portal propagates context to micro-apps
The portal SHALL pass the current user, tenant, permissions, and an authenticated API client to the mounted micro-app via Qiankun `mount` props.

#### Scenario: Micro-app receives context
- **WHEN** the `order-ui` micro-app is mounted
- **THEN** it receives a `user` object with `id`, `tenant_id`, `roles`, and `permissions`, plus an `apiClient` configured with the access token

### Requirement: Portal provides navigation and layout
The portal SHALL render a top navigation bar with the current user display name, a logout button, and links to each registered micro-app.

#### Scenario: User views console home
- **WHEN** the authenticated user opens the portal root path
- **THEN** the layout shows the navigation bar and a welcome panel

### Requirement: Portal logout clears session
The portal SHALL remove the stored access token and unmount the active micro-application when the user clicks logout.

#### Scenario: User logs out
- **WHEN** the user clicks the logout button
- **THEN** the token is cleared, the micro-app is unmounted, and the user is returned to the login page
