## Why

当前 PlantX Kit 平台已经落地了 kit-go 运行时抽象、portal 控制台、order-service 及多个平台 admin 微前端，但实际实现与《架构设计》及多个 OpenSpec change 存在明显偏离。最核心的问题是：**业务服务仍在重复编写基础设施样板代码，平台注册能力没有真正下沉到 kit，portal 微前端加载未按设计使用 qiankun**。这些偏离会导致后续新增业务服务的成本越来越高，平台能力无法在 kit 中统一演进。本次 change 旨在把「服务注册、路由注册、微前端注册、统一认证/鉴权」彻底收敛到 kit 与 gateway-service，让业务服务只写业务逻辑。

## What Changes

- **kit-go 新增统一注册能力**：业务服务启动时自动向 `gateway-service` 注册 gRPC 服务、REST 路由与微前端元数据，业务 `main.go` 只需传入服务名，不再手写注册客户端。
- **portal 接入 qiankun**：将当前自定义 `<script>` 加载改造为 qiankun 的 `loadMicroApp` / `registerMicroApps`，并按设计文档通过 props 注入 `user/tenant/permissions/apiClient/locale`。
- **portal 动态微前端注册**：基于 manifest（先静态、后从 gateway-service 拉取）动态渲染业务路由和菜单，新增业务微前端不再修改 portal 源码。
- **修正 kit-cli 缺陷**：修复 `kit new service` 生成的 `main.go` 编译错误；扩展 `kit generate` 支持 buf 工作流并生成前端 TypeScript SDK。
- **修正平台 admin 缺口**：portal 菜单权限同时检查 `admin` 角色和 `platform:admin` 权限；`audit-service` 真正收集 kit server 拦截器日志；`tenant/iam/gateway-service` 补齐 DDD 分层。
- **gateway-service 真正生效**：在保留 Nginx 作为边缘网关的同时，让 gateway-service 的注册表被 portal 消费，用于动态发现服务与微前端。
- **新增 demo 测试服务 `test-service` + `test-ui`**：验证业务服务不写任何 admin/注册逻辑，全部由 kit 完成。

## Capabilities

### New Capabilities
- `kit-gateway-registration`: kit-go 自动向 gateway-service 注册后端服务与 REST 路由。
- `kit-microapp-registration`: kit-go / kit-sdk-kit 统一注册微前端元数据，portal 动态渲染。
- `portal-qiankun-loader`: portal 使用 qiankun 加载子应用，替代自定义 script 注入。
- `portal-dynamic-menu`: portal 基于微前端 manifest 动态生成业务菜单和路由。
- `kit-cli-generate-ts-sdk`: kit generate 调用 buf 生成前端 TypeScript SDK。
- `kit-cli-scaffold-fix`: 修复 kit new service 模板编译错误，生成可立即构建的服务。
- `platform-admin-menu-permission`: portal admin 菜单同时校验 `admin` 角色和 `platform:admin` 权限。
- `platform-audit-log-collection`: kit server 拦截器产生审计事件并写入 audit-service。
- `platform-service-layering`: tenant/iam/gateway-service 按 DDD 分层重构（domain/app/infra/grpc）。
- `demo-test-service`: 新增 test-service 与 test-ui，作为业务服务零 admin 逻辑的示范。

### Modified Capabilities
<!-- 无现有 spec，留空 -->

## Impact

- `kit/kit-go/server/server.go`：新增 `GatewayRegistrar`、`ServiceName` 选项及自动注册逻辑。
- `kit/kit-go/gateway/`（新增）：gateway-service gRPC 客户端与 `AutoRegister` 辅助函数。
- `platform/gateway-service/api/gateway.proto`：扩展 Service 消息，新增微前端元数据字段与 `RegisterMicroApp` / `ListMicroApps`。
- `platform/gateway-service/cmd/main.go`：实现微前端注册表。
- `apps/portal/src/MicroAppPage.tsx`、`App.tsx`、`Layout.tsx`：改为 qiankun 加载 + manifest 动态路由/菜单。
- `kit/kit-cli/internal/scaffold/service.go`、`cmd/generate.go`：修复模板，接入 buf 生成 TS SDK。
- `platform/tenant-service`、`platform/iam-service`、`platform/gateway-service`：补齐 DDD 分层。
- `platform/audit-service`：接入 kit 拦截器日志；`kit/kit-go/server` 拦截器产生审计事件。
- `services/test-service/`、`services/test-service/web/test-ui/`（新增）：零 admin 逻辑的 demo 服务。
- `deployments/docker-compose/docker-compose.yml`、`nginx.conf`、`apps/portal/Dockerfile`：集成 test-service/test-ui。
