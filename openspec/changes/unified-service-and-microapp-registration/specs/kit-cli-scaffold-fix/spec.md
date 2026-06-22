## ADDED Requirements

### Requirement: kit new service generates compilable main.go
`kit new service <name>` SHALL generate a `cmd/main.go` that compiles without modification and uses `server.New` with `context.Background()`.

#### Scenario: Developer scaffolds a new service
- **WHEN** the developer runs `kit new service demo`
- **THEN** `cd services/demo && go build ./cmd` succeeds

### Requirement: kit new service includes sqlc.yaml
`kit new service <name>` SHALL generate a `sqlc.yaml` configured for the service's migrations and queries.

#### Scenario: Developer runs sqlc generate
- **WHEN** the developer runs `sqlc generate` in the new service
- **THEN** it produces code under `internal/infra/sqlc/`

### Requirement: kit new service generates correct format strings
The generated `cmd/main.go` SHALL not contain unmatched `fmt.Printf` verbs or incorrect `srv.Run` arguments.

#### Scenario: Inspect generated main.go
- **WHEN** the developer reads `services/demo/cmd/main.go`
- **THEN** `fmt.Println` contains the correct service name and `srv.Run(context.Background())` is called
