## ADDED Requirements

### Requirement: All services are containerized
Every service SHALL include a `Dockerfile` that produces a runnable container image.

#### Scenario: Building a service image
- **WHEN** a CI pipeline runs `docker build` on a service directory
- **THEN** it produces a container image without service-specific build scripts

### Requirement: Service startup depends only on environment variables and config files
A service SHALL obtain all external configuration from environment variables and mounted configuration files, not from hardcoded values.

#### Scenario: Configuring database connection
- **WHEN** an operator sets `KIT_DB_DSN` environment variable
- **THEN** the service connects to the specified database without code changes

### Requirement: Docker Compose deployment is supported
The Kit platform SHALL provide a `deployments/docker-compose/docker-compose.yml` that starts all platform components and a sample service.

#### Scenario: One-command local deployment
- **WHEN** an operator runs `docker-compose up` in `deployments/docker-compose/`
- **THEN** MaxKey, OPA, PostgreSQL, NATS, gateway, and demo services start successfully

### Requirement: Kubernetes Helm deployment is supported
The Kit platform SHALL provide Helm charts under `deployments/k8s/` for production deployment.

#### Scenario: Installing on Kubernetes
- **WHEN** an operator runs `helm install plantx deployments/k8s/plantx`
- **THEN** all platform services are deployed with configurable replicas, resources, and ingress

### Requirement: Bare-metal deployment is supported
The Kit platform SHALL support compiling services to binaries and running them with systemd or supervisor on virtual machines or bare metal.

#### Scenario: Running without containers
- **WHEN** a service is built with `go build` and started on a Linux server
- **THEN** it runs successfully behind Nginx or APISIX with external PostgreSQL and NATS
