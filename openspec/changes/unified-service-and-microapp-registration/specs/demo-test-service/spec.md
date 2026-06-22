## ADDED Requirements

### Requirement: test-service contains only business logic
`services/test-service` SHALL implement a `TestService` gRPC service with `Ping` and `Echo` methods, and SHALL NOT contain any service registration, routing, authentication, authorization, or admin logic.

#### Scenario: Inspect test-service main.go
- **WHEN** reading `services/test-service/cmd/main.go`
- **THEN** it only assembles the kit server, registers the TestService handler, and runs; no gateway client or OPA setup is present

### Requirement: test-service auto-registers via kit
`services/test-service` SHALL enable kit auto-registration by configuring `GatewayRegistrar` in `server.Options`.

#### Scenario: test-service starts
- **WHEN** `test-service` starts
- **THEN** it appears in `gateway-service` `/api/gateway/v1/services` list

### Requirement: test-ui is a qiankun child app
`services/test-service/web/test-ui` SHALL export qiankun `bootstrap` / `mount` / `unmount` lifecycle functions and build as an IIFE bundle.

#### Scenario: test-ui builds
- **WHEN** running `pnpm run build` in `services/test-service/web/test-ui`
- **THEN** it produces `dist/test-ui.iife.js`

### Requirement: test-ui uses generated SDK
`services/test-service/web/test-ui/src/TestPage.tsx` SHALL import `TestServiceClient` from `@plantx/test-sdk-api` and call the generated `echo` method.

#### Scenario: test-ui calls backend
- **WHEN** the user types a message and clicks Echo
- **THEN** the page displays the echoed response returned by `TestServiceClient.echo`
