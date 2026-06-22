## ADDED Requirements

### Requirement: kit-go server auto-registers backend service on startup
`kit/kit-go/server.Server` SHALL register the running service with `gateway-service` after both gRPC and HTTP listeners are healthy, when a `GatewayRegistrar` is configured.

#### Scenario: Service starts with GatewayRegistrar configured
- **WHEN** a service calls `server.New(server.Options{GatewayRegistrar: gateway.AutoRegister("test-service"), ServiceName: "test-service"})` and starts successfully
- **THEN** `gateway-service` receives a `RegisterService` request with name `test-service`, grpc_host derived from the service container, and rest_prefix `/api/test/v1`

### Requirement: kit-go server deregisters backend service on shutdown
`kit/kit-go/server.Server` SHALL deregister the service from `gateway-service` when `Shutdown` is called.

#### Scenario: Service shuts down gracefully
- **WHEN** `srv.Shutdown(ctx)` is invoked
- **THEN** the service entry is removed from `gateway-service` registry

### Requirement: GatewayRegistrar is optional
A service SHALL be able to start without a `GatewayRegistrar`; in that case no registration request is sent.

#### Scenario: Service starts without GatewayRegistrar
- **WHEN** a service calls `server.New(server.Options{})` without `GatewayRegistrar`
- **THEN** the server starts normally and does not contact `gateway-service`
