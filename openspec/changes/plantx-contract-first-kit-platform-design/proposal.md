## Why

PlantX 需要从一个“系统集合”演进为一个面向业务开发者的 **Kit 平台**。当前业务团队在不同服务中重复处理认证、鉴权、租户、日志、事件、部署等基础设施问题，导致开发效率低、实现不一致、替换成本高。我们需要一份以 **proto + sqlc 为契约源头**、以 **DDD 为业务组织方式**、以 **Kit 运行时集中收敛基础设施** 的架构提案，为后续平台骨架开发建立统一蓝图。

## What Changes

- 建立 PlantX Kit 平台总体架构，明确 `proto + sqlc > sdk > kit > ui` 的契约约束链。
- 定义 Kit 运行时抽象接口，将认证（MaxKey）、鉴权（OPA）、租户、事件总线、日志、服务发现等基础设施与业务代码解耦。
- 设计业务服务标准工程结构（DDD 分层：domain / app / infra / interfaces）。
- 设计前端三层包结构（`{service}-sdk-api` / `{service}-sdk-kit` / `{service}-ui`）以及 qiankun 主应用集成方式。
- 明确多部署形态支持：Kubernetes + Helm、Docker Compose、二进制/裸机。
- 明确业务开发工作流：从 `kit new service` 到本地开发、测试、部署的标准路径。

## Capabilities

### New Capabilities

- `contract-first-service-definition`：以 proto 定义服务接口与事件，以 sqlc 定义数据访问契约，作为前后端、SDK、UI 的统一源头。
- `kit-runtime-abstractions`：Kit 运行时框架提供的认证、鉴权、租户上下文、事件总线、日志追踪、数据库访问等抽象接口与默认实现。
- `kit-cli-scaffolding`：业务服务脚手架工具，支持 `kit new service`、`kit generate`、`kit dev` 等命令。
- `micro-frontend-contract-chain`：qiankun 主应用与业务子应用的集成，以及 `sdk-api > sdk-kit > ui` 的前端包约束链。
- `multi-tenant-data-isolation`：基于 `tenant_id` 的行级租户隔离，业务 SQL 无需手动注入租户条件。
- `pluggable-infrastructure-providers`：基础设施实现可替换机制，支持 MaxKey/Keycloak、OPA/Casbin、NATS/Kafka、K8s/Docker 等不同实现。
- `multi-deployment-packaging`：支持 K8s Helm、Docker Compose、二进制/裸机三种部署形态，服务启动仅依赖环境变量和配置文件。

### Modified Capabilities

- 无现有 spec，本次不修改既有能力。

## Impact

- **后端技术栈**：Go、Protobuf、buf、gRPC-Gateway、sqlc、PostgreSQL。
- **前端技术栈**：qiankun、Ant Design、React、pnpm workspace / Turborepo。
- **基础设施**：MaxKey（认证）、OPA（鉴权）、NATS JetStream（事件）、PostgreSQL（数据库）、Loki/Jaeger（可观测性）。
- **部署方式**：Kubernetes + Helm（生产）、Docker Compose（中小规模/私有化）、二进制 + Nginx/APISIX（裸机/信创）。
- **组织影响**：平台团队负责 Kit 运行时与基础设施；业务团队仅编写 proto、sqlc、领域逻辑和前端页面。
