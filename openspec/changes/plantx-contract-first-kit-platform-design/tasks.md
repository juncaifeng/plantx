## 0. Milestones

| Milestone | 覆盖任务组 | 核心交付物 | 验收标准 |
|---|---|---|---|
| **M1: Kit 骨架可运行** | 1, 2, 4, 5, 6, 7 | kit-go 抽象接口、kit-cli 基础命令、order demo 服务、Docker Compose 本地环境 | `kit new service order` + `kit dev up` 能启动，调用创建订单接口返回成功 |
| **M2: 认证鉴权与租户闭环** | 3（认证/鉴权/租户相关）、8、9 | MaxKey/OPA 集成、接口级鉴权、行级租户隔离 | 未登录请求 401，无权限请求 403，租户 A 无法读取租户 B 数据 |
| **M3: 前端契约链闭环** | 10、11 | kit-ui、kit-sdk-kit、order-ui 子应用、portal 主应用 | 登录 portal 后动态加载 order 子应用，完成订单 CRUD |
| **M4: 部署与可观测性** | 12、13、14 | K8s Helm、二进制/systemd 部署、日志/追踪/指标、完整文档 | 至少验证 Docker Compose、K8s、二进制三种形态中的一种；e2e 冒烟测试通过 |

**Milestone 切换原则**：前一里程碑所有验收标准达成后，方可进入下一里程碑。每个里程碑结束时进行架构 review 和 demo。

## 1. Project Foundation

- [x] 1.1 Initialize monorepo structure: `kit/`, `platform/`, `services/`, `apps/`, `deployments/`, `proto/`
- [x] 1.2 Create root `go.work` and `pnpm-workspace.yaml` configuration files
- [x] 1.3 Define repository coding standards and lint rules (golangci-lint, prettier, eslint)
- [x] 1.4 Set up CI pipeline skeleton for build, test, and lint

## 2. Kit-Go Runtime Interfaces

- [x] 2.1 Define `auth.Authenticator` interface and `auth.UserInfo` struct
- [x] 2.2 Define `authz.Authorizer` interface and proto annotation for policy binding
- [x] 2.3 Define `tenant.Resolver` and `kitctx` helpers for user/tenant context propagation
- [x] 2.4 Define `event.Bus` interface for publishing and subscribing to domain events
- [x] 2.5 Define `db.DB` interface, transaction helper, and sqlc integration utilities
- [x] 2.6 Define `log.Logger` interface and structured logging helpers
- [x] 2.7 Define `config.Loader` and `discovery.Registry` abstractions
- [x] 2.8 Implement Kit server startup wrapper with interceptor chain (recovery, trace, log, auth, authz, metrics)

## 3. Infrastructure Provider Implementations

- [x] 3.1 Implement MaxKey `auth.Authenticator` using OIDC/JWT validation
- [x] 3.2 Implement OPA `authz.Authorizer` with Sidecar and standalone modes
- [x] 3.3 Implement NATS JetStream `event.Bus` with retry and context propagation
- [x] 3.4 Implement PostgreSQL `db.DB` with connection pool and transaction support
- [x] 3.5 Implement Zap-based `log.Logger` with OpenTelemetry trace context injection
- [x] 3.6 Implement environment-variable `config.Loader`
- [x] 3.7 Implement static and K8s DNS `discovery.Registry` providers

## 4. Kit Common Proto

- [x] 4.1 Create `proto/plantx/kit/authz.proto` with policy annotation definitions
- [x] 4.2 Create `proto/plantx/kit/context.proto` with request context messages
- [x] 4.3 Create `proto/plantx/kit/event.proto` with event envelope definitions
- [x] 4.4 Configure buf workspace and generate Go/TypeScript code for Kit protos

## 5. Kit-CLI Tooling

- [x] 5.1 Implement `kit new service <name>` command with `--ui` and `--gateway` flags
- [x] 5.2 Implement `kit generate` command invoking buf, sqlc, and SDK generators
- [x] 5.3 Implement `kit migrate new <name>` command for timestamped migration files
- [x] 5.4 Implement `kit dev up/down/logs` commands using Docker Compose
- [x] 5.5 Implement `kit test` command running Go and frontend tests
- [x] 5.6 Implement `kit build` command producing container images

## 6. Contract-First Code Generation

- [x] 6.1 Configure buf to generate Go gRPC, gRPC-Gateway, and OpenAPI from service protos
- [x] 6.2 Configure sqlc generation for service queries with tenant-aware plugin hook
- [x] 6.3 Configure TypeScript SDK generator for `{service}-sdk-api`
- [x] 6.4 Validate generated code in CI and fail on drift

## 7. Demo Service: order-service

- [x] 7.1 Create `services/order/api/order.proto` with sample RPCs and events
- [x] 7.2 Create `services/order/migrations/001_init.up.sql` with `tenant_id` column
- [x] 7.3 Create `services/order/internal/infra/sqlc/queries.sql`
- [x] 7.4 Implement domain model, repository interface, and application service
- [x] 7.5 Implement gRPC handler and wire into Kit server
- [x] 7.6 Add unit tests for domain logic and integration tests for repository
- [x] 7.7 Add `Dockerfile` and verify container build

## 8. Authentication and Authorization Integration

- [x] 8.1 Deploy MaxKey in Docker Compose and configure OIDC client for PlantX
- [x] 8.2 Deploy OPA in Docker Compose with basic RBAC Rego policies
- [x] 8.3 Configure API gateway to proxy MaxKey login and inject user headers
- [x] 8.4 Verify `order-service` rejects unauthorized requests and accepts authorized ones
- [x] 8.5 Add Rego policy tests and integrate into CI

## 9. Multi-Tenant Data Isolation

- [x] 9.1 Implement tenant context extraction from MaxKey JWT claims
- [x] 9.2 Implement automatic `tenant_id` injection in sqlc-generated queries
- [x] 9.3 Add repository wrapper that enforces tenant isolation
- [x] 9.4 Add test verifying cross-tenant data access is denied
- [x] 9.5 Propagate tenant context through gRPC metadata and event envelopes

## 10. Frontend Contract Chain

- [x] 10.1 Create `kit-ui` package with common AntD components and layout
- [x] 10.2 Create `kit-sdk-api` package with base HTTP client and auth header injection
- [x] 10.3 Create `kit-sdk-kit` package with `useKitUser`, `useKitPermission`, and tenant context
- [x] 10.4 Generate `order-sdk-api` from `order.proto`
- [x] 10.5 Create `order-sdk-kit` with `useOrders` hook and order state management
- [x] 10.6 Create `order-ui` qiankun sub-application with sample pages

## 11. Micro-Frontend Portal

- [x] 11.1 Create `apps/portal` qiankun main application with Ant Design Pro
- [x] 11.2 Implement OIDC login flow redirecting to MaxKey
- [x] 11.3 Fetch user/tenant/permissions after login and store in global state
- [x] 11.4 Dynamically load registered sub-applications based on permissions
- [x] 11.5 Pass shared context (`user`, `tenant`, `permissions`, `apiClient`) to child apps via props

## 12. Deployment Packaging

- [x] 12.1 Create `deployments/docker-compose/docker-compose.yml` for MaxKey, OPA, PostgreSQL, NATS, gateway, and demo services
- [x] 12.2 Create `deployments/k8s/plantx` Helm chart with values for all services
- [x] 12.3 Create systemd service unit templates for binary deployment
- [x] 12.4 Create Nginx/APISIX gateway configuration for non-K8s deployment
- [x] 12.5 Verify Docker Compose deployment end-to-end

## 13. Observability

- [x] 13.1 Integrate OpenTelemetry trace exporter into Kit server
- [x] 13.2 Configure structured request logging with trace and tenant fields
- [x] 13.3 Add health check and readiness endpoints
- [x] 13.4 Add Prometheus metrics endpoint for request latency and error rates

## 14. Documentation and Validation

- [x] 14.1 Write `docs/kit-getting-started.md` for business developers
- [x] 14.2 Write `docs/kit-platform-team-guide.md` for platform maintainers
- [x] 14.3 Write `docs/deployment-guide.md` covering K8s, Docker Compose, and binary modes
- [x] 14.4 Run end-to-end smoke test: login → create order → list orders across tenant boundary
- [x] 14.5 Review and resolve design open questions, update design.md if decisions change

## 15. Branch and Release Management

- [x] 15.1 Select and document Git branching model in `docs/git-branch-strategy.md` (Git Flow / GitHub Flow / Trunk-Based)
- [x] 15.2 Define branch naming conventions: `feature/<issue-id>-description`, `fix/<issue-id>-description`, `release/vx.y.z`, `hotfix/<description>`
- [x] 15.3 Configure branch protection for `main` and `release/*`: require at least one PR review, CI pass, and up-to-date branch
- [x] 15.4 Enforce Conventional Commits and add commitlint check in CI
- [x] 15.5 Set up monorepo changeset workflow for independent versioning of `kit-go`, `kit-cli`, `kit-ui`, platform services, and business services
- [x] 15.6 Define release process: cut `release/vx.y.z` branch, run full integration tests, tag `vx.y.z`, merge back to `main`
- [x] 15.7 Define hotfix process: branch from latest tag, apply fix, tag `vx.y.z+1`, merge to both `main` and active release branches
- [x] 15.8 Configure CI to build and publish all container images on tag push
- [x] 15.9 Document rollback strategy for K8s, Docker Compose, and binary deployments
- [x] 15.10 Define PR merge rules: squash vs merge commit per change type, required checks, and CODEOWNERS
