## ADDED Requirements

### Requirement: Frontend packages follow the sdk-api > sdk-kit > ui chain
Each service SHALL expose three frontend packages: `{service}-sdk-api`, `{service}-sdk-kit`, and `{service}-ui`, with strict dependency rules.

#### Scenario: Package dependency validation
- **WHEN** the CI pipeline builds `{service}-ui`
- **THEN** the build SHALL fail if `{service}-ui` imports `{service}-sdk-api` directly or imports any non-kit business package

### Requirement: sdk-api is generated from proto
The `{service}-sdk-api` package SHALL be auto-generated from the service proto and SHALL handle authentication headers and base URL configuration.

#### Scenario: Calling a generated API
- **WHEN** a frontend developer calls `orderApi.createOrder({...})`
- **THEN** the request includes the correct path, typed payload, and `Authorization` header

### Requirement: sdk-kit encapsulates business hooks
The `{service}-sdk-kit` package SHALL provide React hooks and state management built on top of `{service}-sdk-api` and `kit-sdk-kit`.

#### Scenario: Using a business hook
- **WHEN** a frontend developer calls `useOrders()`
- **THEN** it returns typed order data, loading state, and error handling without direct API calls

### Requirement: qiankun portal provides shared runtime context
The qiankun portal application SHALL authenticate users, fetch permissions, and inject user/tenant/API client context into child micro-apps.

#### Scenario: Loading a child micro-app
- **WHEN** the portal loads the order sub-app
- **THEN** the sub-app receives `user`, `tenant`, `permissions`, and `apiClient` via props

### Requirement: Child micro-apps do not implement authentication
Child micro-apps SHALL NOT implement their own OIDC login flow; they SHALL rely on the portal for authentication and permission checks.

#### Scenario: Checking permission in a sub-app
- **WHEN** a sub-app uses `useKitPermission('order:create')`
- **THEN** it evaluates the permission injected by the portal without additional authentication requests
