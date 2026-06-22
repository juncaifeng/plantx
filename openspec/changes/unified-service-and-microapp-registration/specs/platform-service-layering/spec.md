## ADDED Requirements

### Requirement: tenant-service follows DDD layering
`platform/tenant-service` SHALL be refactored into `internal/domain/`, `internal/app/`, `internal/infra/repo/`, `internal/infra/sqlc/`, and `internal/interfaces/grpc/` packages.

#### Scenario: Inspect tenant-service structure
- **WHEN** listing `platform/tenant-service/internal/`
- **THEN** it contains `domain/`, `app/`, `infra/`, and `interfaces/grpc/` subdirectories

### Requirement: iam-service follows DDD layering
`platform/iam-service` SHALL be refactored into the same DDD package structure.

#### Scenario: Inspect iam-service structure
- **WHEN** listing `platform/iam-service/internal/`
- **THEN** it contains `domain/`, `app/`, `infra/`, and `interfaces/grpc/` subdirectories

### Requirement: gateway-service follows DDD layering
`platform/gateway-service` SHALL separate its registry storage and handler logic into DDD layers, while keeping the gRPC service definition in `api/`.

#### Scenario: Inspect gateway-service structure
- **WHEN** listing `platform/gateway-service/internal/`
- **THEN** it contains `domain/`, `app/`, `infra/`, and `interfaces/grpc/` subdirectories
