## ADDED Requirements

### Requirement: kit generate uses buf for all services
`kit generate` SHALL invoke `buf generate` for each service's proto directory, producing Go gRPC / grpc-gateway code and TypeScript SDK code.

#### Scenario: Developer runs kit generate for test-service
- **WHEN** the developer runs `kit generate` in `services/test-service`
- **THEN** `services/test-service/api/test.pb.go`, `test_grpc.pb.go`, `test.pb.gw.go`, and `services/test-service/web/test-sdk-api/src/generated/test.ts` are produced

### Requirement: Generated TS SDK reuses kit-sdk-api
The generated TypeScript SDK SHALL import `KitApiClient` from `@plantx/kit-sdk-api` and wrap it with service-specific typed methods.

#### Scenario: test-ui uses generated TestServiceClient
- **WHEN** `test-ui` imports `TestServiceClient` from `@plantx/test-sdk-api`
- **THEN** it can call `client.echo({ message: "hello" })` with full TypeScript types

### Requirement: buf.gen.yaml includes TypeScript plugin
`buf.gen.yaml` SHALL include a TypeScript plugin configuration for generating frontend SDKs.

#### Scenario: CI runs buf generate
- **WHEN** CI executes `buf generate`
- **THEN** TypeScript SDK files are generated alongside Go files
