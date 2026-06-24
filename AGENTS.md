# PlantX 平台架构规范（AGENTS.md）

> 本文档约束 PlantX 平台在微服务、微前端、契约管理、代码生成、DDD 分层等方面的设计原则与实施标准。
> 适用于所有 AI 编码助手、平台开发者、业务服务开发者。

---

## 1. 核心设计原则

### 1.1 契约优先（Contract First）

所有后端服务必须先定义 `proto`，再生成代码、再实现业务逻辑。

- **禁止**先写 Go struct / handler，再补 proto。
- proto 是服务对外暴露能力的唯一权威契约。
- proto 同时驱动：
  - gRPC server/client
  - HTTP gateway（grpc-gateway）
  - OpenAPI 文档
  - TypeScript / Go SDK
  - 授权注解（`plantx.kit.authz.action`）

### 1.2 基础设施与业务解耦

业务开发者只需关心：

1. 写 proto 契约
2. 写领域模型与应用逻辑
3. 调用 `kit` 提供的 `server.New` + `gateway.AutoRegister`

不关心：

- nginx / apisix / envoy 如何配置路由
- OPA 策略怎么写
- JWT 怎么解析
- SDK 怎么生成
- 菜单怎么渲染

### 1.3 UI 只负责展示逻辑

- UI（portal、admin-ui、业务微前端）只负责：**渲染、交互、路由、表单验证**。
- 不直接调用后端接口，统一通过 `@plantx/kit-sdk-api` 生成的 SDK 调用。
- 不硬编码菜单、权限、微前端列表，全部从 `registry-service` 动态获取。

---

## 2. Proto 规范

### 2.1 目录结构

```
platform/<svc>/api/<svc>.proto
services/<svc>/api/<svc>.proto
proto/plantx/kit/        # 平台公共扩展
proto/google/api/        # google api 标准
```

### 2.2 包名规范

```protobuf
package plantx.<service>.v1;
option go_package = "github.com/plantx/<group>/<service>/api";
```

示例：

- `platform/tenant-service/api/tenant.proto` → `package plantx.tenant.v1;`
- `services/order/api/order.proto` → `package plantx.order.v1;`

### 2.3 HTTP 路径规范

统一 REST 前缀：

```protobuf
rpc CreateOrder(CreateOrderRequest) returns (Order) {
  option (google.api.http) = {
    post: "/api/order/v1/orders"
    body: "*"
  };
}

rpc ListOrders(ListOrdersRequest) returns (OrderList) {
  option (google.api.http) = {
    get: "/api/order/v1/orders"
  };
}
```

规则：

- 路径格式：`/api/<service-short>/v1/<resource>`
- `<service-short>` 为服务名去掉 `-service` 后缀，例如 `order-service` → `order`
- 单资源用路径参数：`/api/order/v1/orders/{id}`
- 动作型 RPC 用动词：`/api/order/v1/orders/{id}/confirm`

### 2.4 授权注解

每个需要保护的 RPC 必须标注权限：

```protobuf
option (plantx.kit.authz.action) = {
  service: "order"
  resource: "order"
  operation: "create"
};
```

权限表达式：`<resource>:<operation>`，例如 `order:create`、`order:read`、`order:list`。

### 2.5 生成配置

`buf.gen.yaml` 统一配置：

```yaml
plugins:
  - plugin: go
  - plugin: go-grpc
  - plugin: grpc-gateway
  - plugin: ts            # TypeScript SDK
  - plugin: buf.build/community/google-gnostic-openapi  # OpenAPI v3 文档
```

- Go / gRPC / grpc-gateway / TypeScript 生成目录跟随 proto 源目录，使用 `paths=source_relative`。
- OpenAPI 插件使用 `output_mode=source_relative`，先生成到 `openapi/` 下的源相对路径。
- `scripts/generate.sh` 在 `npx buf generate` 后执行归一化，将服务契约移动到 `openapi/<service-short>.yaml`（如 `openapi/iam.yaml`、`openapi/order.yaml`），并清理共享 proto 与合并产物。

---

## 3. SQLc 规范

### 3.1 使用场景

- 需要使用 PostgreSQL 持久化的服务必须使用 sqlc。
- 不允许手写大量 `database/sql` 模板代码。

### 3.2 目录结构

```
services/<svc>/
  migrations/
    001_init.up.sql
    001_init.down.sql
  internal/infra/sqlc/
    queries.sql
    queries.sql.go   # 生成
    models.go        # 生成
    db.go            # 生成
    querier.go       # 生成
  internal/infra/repo/
    <entity>_repo.go
```

### 3.3 配置

每个需要 sqlc 的服务独立配置，或在根 `sqlc.yaml` 中按服务拆分：

```yaml
sql:
  - schema: "services/order/migrations"
    queries: "services/order/internal/infra/sqlc"
    engine: "postgresql"
    gen:
      go:
        package: "sqlc"
        out: "services/order/internal/infra/sqlc"
        emit_interface: true
        emit_json_tags: true
```

### 3.4 分层约束

- `sqlc.Queries` 只允许在 `internal/infra/repo` 中使用。
- `internal/app` 和 `internal/domain` 不直接依赖 sqlc。
- `internal/domain` 定义 `Repository` 接口，`internal/infra/repo` 实现。

---

## 4. DDD 分层规范

每个 Go 服务必须遵循以下分层：

```
cmd/main.go                  # 依赖组装、启动入口
api/                         # proto 生成代码（接口层输入）
internal/
  domain/                    # 领域层：聚合根、值对象、领域事件、Repository 接口
  app/                       # 应用层：用例、事务编排
  infra/
    repo/                    # 仓库实现
    sqlc/                    # sqlc 生成代码
    event/                   # 事件发布实现（可选）
  interfaces/
    grpc/handler.go          # gRPC handler
    http/                    # 自定义 HTTP handler（如有）
```

### 4.1 依赖方向

```
domain ← app ← infra/repo
        ↓
   interfaces/grpc
```

- `domain` 不依赖任何其他层。
- `app` 只依赖 `domain`。
- `infra` 依赖 `domain` + `app`。
- `interfaces` 依赖 `app` + `api`。

### 4.2 Handler 必须薄

`interfaces/grpc/handler.go` 只做：

1. 参数校验 / 转换
2. 调用 `app` 层用例
3. 返回结果转换

不允许在 handler 里写业务逻辑。

---

## 5. OpenAPI 与 SDK 生成

### 5.1 自动生成链

每个服务的 proto 变更触发以下生成：

```
proto
  ├── buf generate
  │     ├── Go gRPC server/client
  │     ├── grpc-gateway HTTP handler
  │     ├── TypeScript SDK (@plantx/kit-sdk-api/<svc>)
  │     └── OpenAPI spec (openapi/<svc>.yaml)
```

### 5.2 SDK 包结构

```
kit/
  kit-sdk-api/               # 前端 TypeScript SDK
    src/
      order/
        index.ts
        client.ts
        types.ts
      tenant/
      iam/
      gateway/
      ...
  kit-sdk-kit/               # 前端 kit 运行时
    src/
      context.tsx            # KitContext
      client.ts              # createClient
      micro-app.ts           # qiankun 封装
  kit-go/
    server/                  # 后端 server 框架
    gateway/                 # 服务注册客户端
    auth/                    # 认证
    authz/                   # 授权
    sdk/                     # 生成的 Go SDK（可选）
      order/
      tenant/
```

### 5.3 SDK 发布

- TypeScript SDK 作为 workspace package 发布到内部 npm registry。
- Go SDK 可以通过 `go.mod` replace 或作为独立 module 发布。
- UI 只通过 `@plantx/kit-sdk-api/<svc>` 调用后端，不允许直接 `fetch('/api/xxx')`。

---

## 6. Kit 业务包规范

`kit` 是平台层，包含两类能力：

### 6.1 框架能力（Framework）

- `kit-go/server`：gRPC/HTTP server 启动、拦截器、健康检查
- `kit-go/gateway`：服务注册客户端
- `kit-go/auth`：MaxKey/OIDC 认证
- `kit-go/authz`：OPA/fallback 授权
- `kit-go/tenant`：租户解析
- `kit-go/event`：事件总线抽象

### 6.2 生成能力（Generated）

- `kit-sdk-api/<svc>`：按服务生成的 TS SDK
- `openapi/<svc>.yaml`：按服务生成的 OpenAPI

### 6.3 使用约束

业务服务 `main.go` 示例：

```go
srv := server.New(server.Options{
    ServiceName: "order-service",
    Authenticator:  authmaxkey.New(...),
    Authorizer:     authzopa.New(...),
    GatewayRegistrar: gateway.AutoRegister("order-service",
        gateway.WithApplication(gateway.Application{
            Key:         "order",
            Name:        "Order Management",
            LabelKey:    "nav.orders",
            Icon:        "AppstoreOutlined",
            Description: "Order management application",
            Status:      api.ApplicationStatus_APPLICATION_STATUS_ACTIVE,
            SortOrder:   10,
        }),
        gateway.WithMicroApp(gateway.MicroApp{
            Name:              "order-ui",
            Route:             "/order",
            BundleURL:         "/apps/order-ui/order-ui.js",
            MenuLabelKey:      "nav.orders",
            RequirePermission: "order:read",
        }),
    ),
})
```

---

## 7. UI 只负责展示逻辑

### 7.1 Portal 职责

- 登录态管理
- 从 `registry-service` 拉取菜单和微前端列表
- 根据用户权限过滤菜单
- 通过 qiankun 加载子应用

### 7.2 业务微前端职责

- 只负责自身业务 UI
- 通过 `@plantx/kit-sdk-api/<svc>` 调用后端
- 通过 `@plantx/kit-sdk-kit` 获取 `user`、`tenant`、`permissions`、`apiClient`

### 7.3 Admin 控制台职责

- 应用管理（CRUD Application）
- 服务管理（查看/下线 Service）
- 菜单管理（动态配置菜单结构、绑定权限）
- 权限管理（角色 ↔ 权限点）
- API Explorer（在线查看 proto/openapi、调试接口）

### 7.4 禁止事项

- UI 不允许硬编码菜单。
- UI 不允许硬编码微前端列表。
- UI 不允许直接 `fetch('/api/xxx')`。
- UI 不允许在组件里写业务规则（如权限判断逻辑除外）。

---

## 8. 服务注册与发现

### 8.1 Registry Service（目标架构）

当前 `platform/gateway-service` 临时兼任注册中心，需拆分为：

- `platform/registry-service`：领域服务，管理 Application / Service / MicroApp / API Contract / Menu / Permission，数据持久化到 PostgreSQL。
- `gateway`（nginx / apisix / envoy）：纯基础设施适配器，从 `registry-service` 同步路由配置。

### 8.2 注册流程

```
业务服务启动
  → kit-go/gateway.AutoRegister
    → gRPC 调用 registry-service.RegisterService
    → gRPC 调用 registry-service.RegisterMicroApp
    → gRPC 调用 registry-service.RegisterApplication
```

### 8.3 发现流程

```
Portal 登录
  → registry-service.ListApplications
    → 渲染产品切换器（Alibaba-Cloud-like product switcher）
  → registry-service.ListMicroApps / ListMenus
    → 根据当前选中的 Application 与用户权限过滤
      → 渲染菜单 + 加载 qiankun 子应用
```

---

## 9. Application（应用）概念

一个 **Application** 是业务交付的最小单元（对应阿里云式产品切换器中的一个产品），包含：

| 字段 | 说明 |
|---|---|
| `id` | 应用唯一标识 |
| `key` | 应用关键字，如 `order`、`tenant`，用于 URL/配置引用 |
| `name` | 应用显示名，如 `Order Management` |
| `labelKey` | 显示名称 i18n key，如 `nav.orders` |
| `icon` | 图标标识，如 `AppstoreOutlined` |
| `description` | 应用描述 |
| `status` | 应用状态；proto enum 值为 `APPLICATION_STATUS_ACTIVE` / `APPLICATION_STATUS_OFFLINE` / `APPLICATION_STATUS_UNSPECIFIED`（默认 `ACTIVE`），UI SDK 已映射为对应字符串 |
| `sortOrder` | 排序权重，数值越小越靠前 |

应用通过 `application_id` / `application_key` 与 `Service`、`MicroApp`、`Menu` 关联。应用注册后：

- 后端服务自动接入 API 网关。
- 前端微前端自动出现在 Portal 产品切换器与菜单。
- 权限点自动进入权限体系。
- API 契约自动进入 API Explorer。

---

## 10. 菜单与权限

### 10.1 权限模型

- 权限点：`resource:operation`
- 角色：一组权限点的集合
- 用户：拥有一个或多个角色
- 菜单项：绑定一个权限点（可选）

### 10.2 菜单来源

菜单由 `registry-service` 提供，结构示例：

```json
{
  "menus": [
    {
      "key": "/order",
      "label": "nav.orders",
      "route": "/order",
      "micro_app": "order-ui",
      "require_permission": "order:read"
    },
    {
      "key": "/admin",
      "label": "nav.admin",
      "children": [
        { "key": "/admin/gateway", "label": "nav.gateway", "require_permission": "gateway:read" }
      ]
    }
  ]
}
```

### 10.3 鉴权流程

1. 用户登录获取 JWT，claims 中包含 `roles` 和 `permissions`。
2. Portal 根据 `permissions` 过滤菜单。
3. 用户访问接口时，`kit-go` 拦截器解析 token。
4. `authz` 层根据 proto 注解的 action 调用 OPA 或本地 fallback 决策。

---

## 11. CI/CD 流程

仓库 CI/CD 定义在 `.github/workflows/` 下，核心工作流如下。

### 11.1 `CI`

文件：`.github/workflows/ci.yml`

触发条件：`push` 到 `main` / `release/*`、`pull_request`、标签 `v*`。

任务：

1. **generate**
   - 安装 `protoc`、`protoc-gen-go`、`protoc-gen-go-grpc`、`protoc-gen-grpc-gateway`、`sqlc`。
   - 执行 `pnpm install` 与 `bash scripts/generate.sh`，统一生成 proto / sqlc / SDK / OpenAPI 代码。
   - 将生成产物打包为 `generated-artifacts` artifact，供下游 job 下载，避免每个 job 重复生成。

2. **lint-go**（依赖 generate）
   - 下载生成产物。
   - 安装 `golangci-lint`。
   - 遍历 `go.work` 中的每个主 module 执行 `golangci-lint run ./...`。
   - 当前 `.golangci.yml` 启用：`gosimple`、`govet`、`ineffassign`、`staticcheck`（排除 SA1019）、`unused`、`gofmt`、`goimports`、`misspell`；禁用 `errcheck`、`revive` 以及生成目录 `api`、`sqlc`、`gen`、`vendor`。

3. **test-go**（依赖 generate）
   - 下载生成产物。
   - 执行 `go test ./kit/kit-go/... ./kit/kit-go/gateway/... ./services/order/... ./kit/kit-cli/...`。

4. **lint-web**（依赖 generate）
   - 下载生成产物。
   - 构建 `kit/*` SDK 包，执行 `pnpm lint` 与 `pnpm typecheck`。

5. **build-images**（依赖 lint-go / test-go / lint-web）
   - 下载生成产物。
   - 使用仓库根目录作为 Docker build context，构建 `services/order/Dockerfile`：
     ```bash
     docker build -f services/order/Dockerfile -t plantx/order-service:${{ github.sha }} .
     ```
   - Dockerfile 中设置 `GOWORK=off`，仅复制必要 module 源码，避免 workspace 全量依赖问题。

6. **publish-images**（依赖 build-images，仅在 `v*` 标签触发）
   - 登录 `ghcr.io`。
   - 构建并推送 `ghcr.io/plantx/order-service:<VERSION>` 与 `latest`。

### 11.2 `Generate Check`

文件：`.github/workflows/generate-check.yml`

- 在 `push` / `pull_request` 时运行 `scripts/generate.sh`。
- 检查生成后的代码是否与仓库一致（`git diff --exit-code`），防止提交者漏提生成产物。

### 11.3 `Generate SDK`

文件：`.github/workflows/generate-sdk.yml`

- 当 `proto/**`、`platform/**/migrations/**`、`scripts/generate.sh` 或本 workflow 变更时触发。
- 运行生成器后通过 `peter-evans/create-pull-request` 自动创建 `chore/regenerate-sdk` PR。

### 11.4 `Release SDK`

文件：`.github/workflows/release-sdk.yml`

- 通过 git tag 触发（`v*`）。
- 标签版本即为所有 `kit/*` 包的发布版本，例如 `v0.2.0` 会：
  - 把 npm 包 `@plantx/kit-sdk-api`、`@plantx/kit-sdk-kit`、`@plantx/kit-ui` 发布为 `0.2.0`。
  - 为 Go 子模块创建 tag：`kit/kit-go/v0.2.0`、`kit/kit-go/gateway/v0.2.0`、`kit/kit-cli/v0.2.0`。
- Workflow 会自动 bump `kit/**/package.json` 版本、执行 `pnpm ci:publish`、推送 Go module tags，并将版本变更提交回 `main`。
- npm 发布脚本：`pnpm ci:publish` → `pnpm -r --filter './kit/**' exec npm publish --access public`。
- npm registry：`https://registry.npmjs.org`。
- Go modules 通过 `git tag` 发布到默认 Go module proxy（`proxy.golang.org`）。
- 必需 secret：`NPM_TOKEN`（需为 Granular Access Token，带 `@plantx` scope 的 publish 权限并启用 **Bypass 2FA**）。

### 11.5 `Commitlint`

文件：`.github/workflows/commitlint.yml`

- 检查最近一次的 commit message 是否符合 conventional commit 规范。
- 使用仓库本地安装的 `@commitlint/cli` 与 `@commitlint/config-conventional`。

### 11.6 本地常用命令

```bash
# 安装依赖
pnpm install

# 生成所有代码（proto / sqlc / SDK / OpenAPI）
bash scripts/generate.sh

# 构建 kit SDK
pnpm -r --filter './kit/**' run build

# 前端 lint / typecheck
pnpm lint
pnpm typecheck

# Go 测试
go test ./kit/kit-go/... ./kit/kit-go/gateway/... ./services/order/... ./kit/kit-cli/...

# 构建 order-service 镜像（仓库根目录执行）
docker build -f services/order/Dockerfile -t plantx/order-service:latest .
```

---

## 12. 实施路线图

### Phase 1：注册中心持久化 ✅

- [x] 设计 `registry-service` proto：`Application`、`Service`、`MicroApp`、`Permission`、`Menu`
- [x] 实现 `registry-service` Postgres 持久化
- [x] 迁移 `gateway.AutoRegister` 目标从 `gateway-service` 到 `registry-service`
- [x] `gateway-service` 从 `registry-service` 读取服务列表做路由/转发

### Phase 2：Portal 动态菜单 ✅

- [x] 删除 `apps/portal/src/microApps.ts` 硬编码
- [x] Portal 登录后调用 `registry-service.ListMicroApps` / `ListMenus`
- [x] Admin 菜单也动态生成

### Phase 3：Kit SDK 与 OpenAPI（进行中）

- [x] 配置 buf openapi 生成插件，输出归一化到 `openapi/<service-short>.yaml`
- [x] 建立 `kit/kit-sdk-api/<svc>` 模块（iam / tenant / gateway / audit / registry）并配置 subpath exports
- [x] CI/CD 自动生成并发布 SDK（`release-sdk.yml` 通过 git tag 自动版本/发布到 npmjs.org 的 `@plantx` scope）
- [x] 平台 admin UI 与 Portal 已迁移到 `@plantx/kit-sdk-api/<svc>` 调用

### Phase 4：Admin 控制台

- [x] registry-service 菜单配置后端：Menu CRUD + ReorderMenus（proto / sqlc / repo / app / handler）
- [x] registry-service 微应用管理后端：UpdateMicroApp / DeleteMicroApp，支持一个服务注册多个微应用
- [x] 应用管理 CRUD（`registry-admin-ui` Applications 标签页）
- [x] 微应用管理 CRUD（`registry-admin-ui` Micro Apps 标签页）
- [x] 服务管理（查看/下线）（`registry-admin-ui` Services 标签页）
- [x] 菜单拖拽配置（`registry-admin-ui` Menus 标签页）
- [x] 权限点与角色管理（`iam-admin-ui` Permissions / Roles 标签页）
- [x] API Explorer（`api-explorer-ui` 基于 `openapi/*.yaml` 渲染 Swagger UI）
- [x] Portal 导航支持从 registry 菜单配置渲染，空配置时回退到微应用自动分组

### Phase 5：网关适配器解耦 ✅

- [x] 将 `gateway-service` 的网关代理职责迁移到 nginx（动态 upstream/location 从 `registry-service` 同步）
- [x] `registry-service` 提供路由同步 API（`SyncRoutes`）
- [x] 支持灰度、限流、鉴权等网关策略（`RoutePolicy` + `GetRoutePolicy` / `SetRoutePolicy`）

### Phase 6：运行时策略闭环与契约流水线（建议）

- [ ] nginx 根据 `authRequired` 策略对接 iam-service 做动态鉴权（或保留 grpc-gateway 层鉴权并统一认证入口）
- [ ] nginx 根据 `canaryWeight` / `canaryHost` 实现流量切分（当前已生成配置注释，待接入加权 upstream）
- [ ] 下线 `platform/gateway-service` 的代理转发代码，保留其管理面 API
- [ ] 实现 CI 契约流水线：`buf lint` / `buf breaking` / 自动生成 OpenAPI / SDK / PR 检查
- [ ] 接入可观测性：分布式 trace、服务指标、网关访问日志

---

## 13. 对 AI 编码助手的约束

在修改本仓库代码时，必须遵守：

1. **先写 proto，再写实现**。任何新增接口必须先更新 proto 并 `buf generate`。
2. **DDD 分层不可破坏**。domain 不能依赖 infra，handler 不能写业务逻辑。
3. **UI 不写业务逻辑和硬编码数据**。菜单、权限、微前端列表必须来自后端。
4. **需要持久化的服务优先使用 sqlc**。禁止手写大量 SQL 模板。
5. **新增服务使用 kit-cli 模板**。保持目录结构和依赖一致。
6. **不要直接修改生成代码**。生成代码由 `buf generate` / `sqlc generate` 产生。
7. **变更注册中心/网关/菜单等基础设施时，同步更新本文档**。

---

*版本：1.2*
*最后更新：2026-06-23*
