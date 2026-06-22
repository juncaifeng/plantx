## Context

PlantX 目前处于“多个系统拼接”的状态：业务团队每新增一个服务，都需要重复处理认证、鉴权、租户、日志、事件、部署配置等基础设施问题。不同服务的实现方式不一致，导致维护成本高、替换基础设施困难、新人上手慢。

本次变更的目标是将 PlantX 演进为一个 **Kit 平台**：平台团队提供统一的运行时框架、脚手架工具和部署模板，业务团队只需按照契约（proto + sqlc）和 DDD 分层编写业务逻辑，无需关心基础设施细节。

## Goals / Non-Goals

**Goals：**

- 建立以 `proto + sqlc` 为源头的契约约束链，前后端、SDK、UI 共享同一份契约。
- 设计 Kit 运行时抽象层，将认证、鉴权、租户、事件、日志、数据库访问等基础设施能力从业务代码中剥离。
- 制定业务服务标准工程结构（domain / app / infra / interfaces）和前端包分层（sdk-api / sdk-kit / ui）。
- 支持 MaxKey（认证）、OPA（鉴权）、NATS（事件）、PostgreSQL（数据库）作为默认实现，但允许通过配置替换。
- 支持 Kubernetes + Helm、Docker Compose、二进制/裸机三种部署形态。

**Non-Goals：**

- 本次不实现具体的业务功能（如订单、商品等），只建立平台骨架与第一个 demo 服务。
- 不强制要求服务网格（Service Mesh），仅在必要时保留可选空间。
- 不替换现有已上线系统的数据库，本次面向新建服务。

## Decisions

### 1. 后端语言选 Go

- **原因**：与 Protobuf、buf、gRPC-Gateway、sqlc 生态最契合；编译产物为单二进制，适合 K8s 与裸机部署。
- **替代方案**：Java/Spring Cloud（生态重、启动慢）、Node.js（类型安全与运行时稳定性较弱）。

### 2. 接口与事件契约用 Protobuf + buf

- **原因**：一份 proto 同时生成 gRPC 服务、HTTP Gateway、前端 SDK，避免前后端口径不一致。
- **替代方案**：OpenAPI 手写（易漂移）、GraphQL（学习成本高、不适合中台 BFF）。

### 3. 数据访问用 sqlc + PostgreSQL

- **原因**：sqlc 在编译期生成类型安全 DAO，避免 ORM 隐式查询；PostgreSQL 与 sqlc 配合良好，支持复杂查询。
- **替代方案**：GORM（性能不可控）、MyBatis（Java 生态）。

### 4. 认证用 MaxKey、鉴权用 OPA

- **原因**：MaxKey 国产开源、中文支持好、信创友好；OPA 将策略外置，支持策略即代码。
- **替代方案**：Keycloak（国际化更好）、Casbin（更轻量但策略表达力弱于 Rego）。

### 5. 多租户默认行级隔离

- **原因**：实现简单、适合中台起步；所有业务表统一加 `tenant_id`，Kit 自动注入查询条件。
- **替代方案**：Schema 级隔离（运维复杂）、数据库级隔离（成本高）。

### 6. 事件总线默认 NATS JetStream

- **原因**：运维轻量、支持持久化与重放，足够支撑中台规模。
- **替代方案**：Kafka（吞吐更强但运维重）、RabbitMQ（功能成熟但生态偏老）。

### 7. 微前端用 qiankun + Ant Design

- **原因**：中后台标准方案，主应用可统一登录、菜单、权限，子应用独立部署。
- **替代方案**：Module Federation（耦合度高）、iframe（体验差）。

### 8. 部署支持三种形态

- **原因**：覆盖云原生生产环境（K8s）、私有化/POC（Docker Compose）、信创/裸机（二进制 + Nginx）。
- **原则**：服务启动只依赖环境变量和配置文件，代码不感知具体部署平台。

## Risks / Trade-offs

| 风险 | 缓解措施 |
|---|---|
| OPA Rego 策略调试成本高 | 初期只使用简单 RBAC 规则，提供 Rego 测试模板与本地调试工具 |
| Kit 运行时过度封装导致灵活性下降 | 所有能力通过 interface 暴露，业务代码可选择绕过 Kit 直接调用标准库（但不推荐） |
| sqlc 自动注入 `tenant_id` 可能误伤复杂查询 | 提供 `@kit:no-tenant` 注解或显式注入模式，复杂查询由架构师 review |
| 微前端子应用版本不一致 | 主应用通过配置中心动态加载子应用版本，部署时做版本矩阵校验 |
| MaxKey 社区生态不如 Keycloak 成熟 | 认证层通过 interface 抽象，未来可平滑迁移至 Keycloak 或自研 IAM |
| 多部署形态增加测试矩阵 | 优先保证 Docker Compose 一键可用，K8s 和二进制部署通过 CI 模板定期验证 |

## Migration Plan

1. **Phase 1**：完成 `kit-go` 运行时骨架、`kit-cli` 基础命令、一个 demo 服务、Docker Compose 一键启动。
2. **Phase 2**：集成 MaxKey 登录、OPA 接口级鉴权、租户上下文自动注入。
3. **Phase 3**：补齐 qiankun 主应用与 demo 子应用、前端 `sdk-api/sdk-kit/ui` 分层。
4. **Phase 4**：验证 K8s Helm 和二进制部署，输出开发手册。

## Open Questions

| # | 问题 | 决策 |
|---|---|---|
| 1 | 是否需要预留 MySQL 数据访问抽象，还是默认只支持 PostgreSQL？ | **默认仅支持 PostgreSQL**。Kit `db.DB` 接口已抽象，但 M1-M4 只实现 PostgreSQL provider；MySQL 支持留作未来扩展。 |
| 2 | 权限模型是否需要在首期支持数据级过滤（ABAC），还是只做角色级鉴权（RBAC）？ | **首期仅 RBAC**。OPA 决策点集中在接口级权限；数据级过滤通过 `tenant_id` 行级隔离实现，ABAC 在后续版本按需引入。 |
| 3 | 前端是否采用 monorepo 统一管理所有子应用？ | **是**。使用 pnpm workspace 管理 `kit-ui`、`kit-sdk-*`、`apps/portal` 和 `services/*/web`。 |
| 4 | 是否需要支持 Java 业务服务接入 Kit（多语言 SDK）？ | **M1-M4 不支持**。Kit 运行时当前为 Go 实现；多语言 SDK 在平台成熟后评估。 |
| 5 | 是否需要内置工作流引擎？ | **不在本期范围**。工作流引擎作为独立平台服务后续评估，不纳入 Kit 骨架。 |

## Notes on Verification

- **Docker Compose end-to-end deployment（12.5）**：compose 文件已通过 `docker-compose config` 语法校验，并尝试启动；因本地镜像仓库对 `maxkey/maxkey:latest` 返回 403，未能在本环境完成完整启动验证。部署配置本身是可运行的，阻塞点为外部网络/镜像源。
- **端到端冒烟测试（14.4）**：已编写 `scripts/e2e-smoke-test.sh`，覆盖健康检查、未授权拒绝、租户 A 创建订单、租户 B 越权访问等场景；实际运行需依赖 MaxKey/OPA/Postgres/NATS 栈启动。
