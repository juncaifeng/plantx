## ADDED Requirements

### Requirement: Micro-app metadata is registered alongside backend service
When a service exposes a micro-frontend, the `GatewayRegistrar` SHALL also register a `MicroApp` manifest containing `name`, `route`, `bundle_url`, `menu_label_key`, and optional `require_permission`.

#### Scenario: test-service registers its micro-app
- **WHEN** `test-service` starts with `GatewayRegistrar: gateway.AutoRegister("test-service", gateway.WithMicroApp(gateway.MicroApp{Name: "test-ui", Route: "/test", BundleURL: "/apps/test-ui/test-ui.js", MenuLabelKey: "nav.test"}))`
- **THEN** `gateway-service` stores both the backend service entry and the micro-app manifest

### Requirement: Portal can list registered micro-apps
`gateway-service` SHALL expose `ListMicroApps` returning all registered micro-app manifests.

#### Scenario: Portal requests micro-app list
- **WHEN** portal calls `GET /api/gateway/v1/micro-apps`
- **THEN** it receives a JSON array containing every registered micro-app manifest

### Requirement: Micro-app manifest is typed in kit-sdk-kit
`kit-sdk-kit` SHALL export a `MicroAppManifest` TypeScript interface matching the proto contract.

#### Scenario: Portal imports manifest type
- **WHEN** portal imports `MicroAppManifest` from `@plantx/kit-sdk-kit`
- **THEN** the type contains `name`, `route`, `bundleUrl`, `menuLabelKey`, and `requirePermission` fields
