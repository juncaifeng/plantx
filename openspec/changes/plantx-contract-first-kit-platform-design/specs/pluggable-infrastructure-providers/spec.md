## ADDED Requirements

### Requirement: Infrastructure capabilities are accessed through abstract interfaces
Business code SHALL interact with authentication, authorization, event bus, logging, configuration, and service discovery only through Kit-defined interfaces.

#### Scenario: Using the event bus
- **WHEN** business code publishes an event
- **THEN** it imports `github.com/plantx/kit/kit-go/event` and NOT `github.com/nats-io/nats.go`

### Requirement: Provider implementations are selected by configuration
The active provider for each infrastructure capability SHALL be determined by configuration, not by code changes in business services.

#### Scenario: Switching authentication provider
- **WHEN** an operator changes `kit.auth.provider` from `maxkey` to `keycloak`
- **THEN** the service starts using the Keycloak authenticator without recompiling business code

### Requirement: Default providers are provided out of the box
The Kit runtime SHALL ship with default providers for MaxKey authentication, OPA authorization, NATS event bus, PostgreSQL database, and Zap logging.

#### Scenario: Default configuration
- **WHEN** a service starts with default Kit configuration
- **THEN** it uses MaxKey, OPA, NATS, PostgreSQL, and Zap providers

### Requirement: Custom providers can be registered
Developers SHALL be able to register custom provider implementations by satisfying the Kit interface and registering them at startup.

#### Scenario: Registering a custom authenticator
- **WHEN** a platform developer implements `auth.Authenticator` and registers it via `kit.RegisterAuthProvider("custom", factory)`
- **THEN** services can select `"custom"` in configuration
