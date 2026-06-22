## Context

PlantX Kit 平台的设计目标是「业务开发零关注基础设施」：认证、鉴权、租户、日志、路由、部署等由平台团队负责，业务服务只写领域逻辑。当前代码已经具备 kit-go 运行时抽象、portal 控制台、order-service 及多个平台 admin 微前端，但实现偏离设计：

1. **服务注册未统一**：`gateway-service` 已提供服务注册 API，但 `order-service` 等未调用它，注册表只是内存数据，未产生实际价值。
2. **微前端加载未按设计使用 qiankun**：`apps/portal/src/MicroAppPage.tsx` 使用自定义 `<script>` 标签加载子应用，没有沙箱隔离、样式隔离、生命周期管理等 qiankun 能力。
3. **微前端路由/菜单硬编码**：新增业务微前端必须修改 `App.tsx` 和 `Layout.tsx` 并重新构建 portal。
4. **前端 SDK 未从 proto 生成**：`order-sdk-api` 是手写 HTTP 客户端，违反契约优先原则。
5. **kit-cli 模板有缺陷**：`kit new service` 生成的 `main.go` 存在编译错误；`kit generate` 不生成前端 SDK。
6. **平台 admin 能力有缺口**：portal admin 菜单未覆盖 `platform:admin` 权限；`audit-service` 未真正收集日志；`tenant/iam/gateway-service` 缺少 DDD 分层。

本次 change 将修正这些偏离，把「注册」这件事彻底收敛到 kit 与 gateway-service。

## Goals / Non-Goals

**Goals:**
- 业务服务启动时自动向 `gateway-service` 注册服务、REST 路由和微前端元数据，业务 `main.go` 不手写注册逻辑。
- Portal 使用 qiankun 加载子应用，并通过 manifest 动态渲染业务路由和菜单。
- `kit generate` 能够基于 buf 生成前端 TypeScript SDK；`kit new service` 生成可编译的服务骨架。
- Portal admin 菜单同时校验 `admin` 角色和 `platform:admin` 权限。
- `audit-service` 真正收集 kit server 拦截器产生的审计日志。
- `tenant-service`、`iam-service`、`gateway-service` 按 DDD 分层重构。
- 新增 `test-service` + `test-ui` 作为「业务零 admin 逻辑」的示范。

**Non-Goals:**
- 不替换 Nginx 作为边缘网关；`gateway-service` 先作为注册中心与动态发现源，Nginx 仍负责实际反向代理。
- 不一次性把所有 admin 微前端路由改为动态（tenant/iam/gateway/audit 的 `/admin/*` 路由先保持硬编码，后续可逐步迁移）。
- 不在本次 change 中把 `mock-auth` 替换为真实 MaxKey。
- 不改动 `notification-service`（当前为空，待后续单独提案）。
- 不修复 Helm chart（属于部署侧独立问题，建议单独提案）。

## Decisions

### 1. 服务注册由 kit-go 自动完成
- **方案 A（推荐）**：在 `kit/kit-go/server/server.go` 中新增 `GatewayRegistrar` 选项，启动成功后自动调用 `gateway-service` 的 `RegisterService`。
- **方案 B**：每个业务服务在 `main.go` 中手动创建 gateway client 并注册。
- **选择 A**：把注册逻辑下沉到 kit，业务服务只关注业务，符合设计目标。

### 2. 微前端元数据注册放在 gateway-service
- **方案 A（推荐）**：扩展 `gateway-service` 的 `Service` 消息，加入 `micro_app` 字段，提供 `RegisterMicroApp` / `ListMicroApps`。
- **方案 B**：新增独立的 `app-registry-service`。
- **选择 A**：复用现有 gateway-service，减少新服务数量；Nginx 仍负责静态资源代理，portal 从 gateway-service 拉取 manifest 做动态发现。

### 3. Portal 微前端加载使用 qiankun
- **方案 A（推荐）**：用 `qiankun` 的 `loadMicroApp` 按需加载，每个子应用一个 container。
- **方案 B**：用 `registerMicroApps` + `start` 一次性注册所有子应用。
- **选择 A**：与当前 React Router 集成更自然，每个路由进入时再加载子应用，避免首屏加载所有 bundle。

### 4. Manifest 先静态、后动态
- **阶段 1（本次）**：portal 中新增 `microApps.ts`，静态声明 order-ui 和 test-ui 的 manifest。
- **阶段 2（后续）**：登录后 portal 从 `/api/gateway/v1/micro-apps` 拉取 manifest，完全动态化。
- **理由**：先静态验证 qiankun + manifest 架构，再接入动态发现，降低一次改动风险。

### 5. 前端 SDK 生成使用 buf TypeScript 插件
- **方案 A（推荐）**：在 `buf.gen.yaml` 中配置 `@protobuf-ts/plugin`，`kit generate` 调用 `buf generate` 为每个服务生成 TS SDK。
- **方案 B**：手写 TypeScript 类型和客户端。
- **选择 A**：严格遵循契约优先，减少手写 SDK 与后端不一致的风险。

### 6. kit server 拦截器产生审计事件
- **方案 A（推荐）**：在 `loggingInterceptor` 中调用 `event.Bus.Publish` 发送审计事件；`audit-service` 订阅该事件并持久化。
- **方案 B**：新增独立审计拦截器。
- **选择 A**：复用已有事件总线抽象，对业务 handler 零侵入。

## Risks / Trade-offs

- **[Risk] qiankun 与当前 script 加载行为差异** → 先在一个测试微前端上验证，保留 order-ui 的加载方式作为 fallback，确认稳定后再迁移 order-ui。
- **[Risk] gateway-service 注册表重启后丢失** → 当前为内存存储，后续可接入 PostgreSQL；本次先接受内存存储，重启后由服务重新注册。
- **[Risk] buf TS 插件引入新依赖** → 先在 test-service 验证，再逐步替换 order-ui 的手写 SDK。
- **[Risk] 动态路由与 React Router 权限守卫集成** → manifest 中预留 `requirePermission` 字段，portal 渲染前结合 `useKitPermission` 过滤。
- **[Trade-off] Nginx 仍负责反向代理**：gateway-service 注册表只用于发现，不替代 Nginx。这样改动最小，但网关层仍有两套概念。

## Migration Plan

1. 修改 `kit/kit-go/server` 和新增 `kit/kit-go/gateway`。
2. 扩展 `gateway-service` proto 与实现。
3. 重构 `tenant/iam/gateway-service` 分层；`audit-service` 接入审计事件。
4. 修改 portal：qiankun + manifest 动态路由/菜单 + admin 权限修复。
5. 修复 `kit-cli` 模板与 `kit generate`。
6. 新增 `test-service` + `test-ui`。
7. 更新 Docker Compose / Nginx / portal Dockerfile。
8. 本地 `kit dev` 与 Docker Compose 双重验证。

## Open Questions

- `test-service` 是否需要一个独立的 PostgreSQL schema，还是先使用内存 repo？（建议先内存 repo，保持 demo 轻量。）
- 微前端 manifest 中权限字段是否先支持简单字符串，还是支持表达式？（建议先字符串。）
