# PlantX Temporal 生命周期状态机设计

## 目标

在 PlantX 平台引入 Temporal 作为生命周期编排引擎，同时在各服务的 DDD 分层中引入状态机，实现菜单、微应用、服务、权限、租户、应用、路由策略等资源的上线/下线/审批流程可观测、可回滚、可审计。

## 核心原则

1. **状态机管"是什么"**：单个实体的状态（online/offline/draft/updating）由各服务自己维护，状态机控制合法转移。
2. **工作流管"怎么做"**：跨服务、长周期、需重试/补偿/审批的流程由 Temporal Workflow 编排。
3. **DB 是状态唯一来源**：Temporal 不替代数据库，Workflow 通过 Activity 调用服务 API 来驱动状态变更。

## 架构

```text
┌─────────────────────────────────────────────────────────────┐
│                        Portal 前端                           │
│            只展示 status = ONLINE / ACTIVE 的资源             │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    平台服务层                                 │
│  registry-service │ iam-service │ tenant-service │ gateway   │
│  各服务内部维护实体的状态机（status + 转移规则）               │
└───────────────────────┬─────────────────────────────────────┘
                        │ 事件 / API 调用
┌───────────────────────▼─────────────────────────────────────┐
│                  Temporal Worker                             │
│                 platform/temporal-worker                     │
│  Workflows: ServiceLifecycle, ApplicationLifecycle,          │
│             MenuLifecycle, PermissionLifecycle,              │
│             TenantProvisioning, TenantOffboarding,           │
│             RolePermissionApproval, RoutePolicyChange        │
└───────────────────────┬─────────────────────────────────────┘
                        │ Activities
┌───────────────────────▼─────────────────────────────────────┐
│                   Temporal Server                            │
│              (PostgreSQL persistence)                        │
└─────────────────────────────────────────────────────────────┘
```

## 基础设施

### Docker Compose

在 `deployments/docker-compose/docker-compose.yml` 增加：

```yaml
  temporal-server:
    image: temporalio/auto-setup:1.22
    container_name: docker-compose-temporal-server-1
    environment:
      - DB=postgres12
      - DB_PORT=5432
      - POSTGRES_USER=plantx
      - POSTGRES_PWD=plantx
      - POSTGRES_SEEDS=postgres
      - DYNAMIC_CONFIG_FILE_PATH=config/dynamicconfig/development-sql.yaml
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - plantx

  temporal-ui:
    image: temporalio/ui:2.21
    container_name: docker-compose-temporal-ui-1
    environment:
      - TEMPORAL_ADDRESS=temporal-server:7233
      - TEMPORAL_CORS_ORIGINS=http://localhost:3000
    ports:
      - "8083:8080"
    depends_on:
      - temporal-server
    networks:
      - plantx

  temporal-worker:
    build:
      context: ../..
      dockerfile: platform/temporal-worker/Dockerfile
    container_name: docker-compose-temporal-worker-1
    environment:
      - TEMPORAL_HOST=temporal-server:7233
      - REGISTRY_SERVICE_GRPC_ADDR=registry-service:8080
      - IAM_SERVICE_GRPC_ADDR=iam-service:8080
      - GATEWAY_SERVICE_GRPC_ADDR=gateway-service:8080
      - AUDIT_SERVICE_GRPC_ADDR=audit-service:8080
      - NOTIFICATION_SERVICE_GRPC_ADDR=notification-service:8080
    depends_on:
      - temporal-server
      - registry-service
    networks:
      - plantx
```

### Worker 模块

```text
platform/temporal-worker/
├── cmd/main.go
├── Dockerfile
├── go.mod
├── internal/
│   ├── workflows/
│   │   ├── service_lifecycle.go
│   │   ├── application_lifecycle.go
│   │   ├── menu_lifecycle.go
│   │   ├── permission_lifecycle.go
│   │   ├── tenant_provisioning.go
│   │   ├── tenant_offboarding.go
│   │   ├── role_permission_approval.go
│   │   └── route_policy_change.go
│   ├── activities/
│   │   ├── registry.go
│   │   ├── iam.go
│   │   ├── gateway.go
│   │   ├── audit.go
│   │   └── notification.go
│   └── worker/
│       └── worker.go
```

## 状态机设计

### Menu 状态机

```go
type MenuStatus string

const (
    MenuStatusDraft    MenuStatus = "DRAFT"
    MenuStatusPending  MenuStatus = "PENDING"
    MenuStatusOnline   MenuStatus = "ONLINE"
    MenuStatusUpdating MenuStatus = "UPDATING"
    MenuStatusOffline  MenuStatus = "OFFLINE"
)

var menuTransitions = map[MenuStatus][]MenuStatus{
    MenuStatusDraft:    {MenuStatusPending},
    MenuStatusPending:  {MenuStatusOnline, MenuStatusOffline, MenuStatusDraft},
    MenuStatusOnline:   {MenuStatusUpdating, MenuStatusOffline},
    MenuStatusUpdating: {MenuStatusOnline, MenuStatusOffline},
    MenuStatusOffline:  {MenuStatusPending, MenuStatusDraft},
}
```

### MicroApp 状态机

与 Menu 相同：`DRAFT → PENDING → ONLINE → OFFLINE`。

### Service 状态机

```go
type ServiceHealthStatus string

const (
    ServiceHealthStatusUnknown    ServiceHealthStatus = "UNKNOWN"
    ServiceHealthStatusHealthy    ServiceHealthStatus = "HEALTHY"
    ServiceHealthStatusUnhealthy  ServiceHealthStatus = "UNHEALTHY"
    ServiceHealthStatusDeregistered ServiceHealthStatus = "DEREGISTERED"
)
```

### Application 状态机

已有 `ACTIVE / OFFLINE`，增加 `PENDING` 用于初始化流程。

### Permission 状态机

```go
type PermissionStatus string

const (
    PermissionStatusDeclared      PermissionStatus = "DECLARED"
    PermissionStatusAvailable     PermissionStatus = "AVAILABLE"
    PermissionStatusUnavailable   PermissionStatus = "UNAVAILABLE"
    PermissionStatusDeprecated    PermissionStatus = "DEPRECATED"
)
```

## Workflow 设计

### 1. ServiceLifecycleWorkflow

**触发事件**：服务注册 / 注销 / 健康状态变化

**状态**：REGISTERING → ROUTE_SYNCED → MENUS_PUBLISHED → PERMISSIONS_LOADED → ONLINE → OFFLINE

**流程**：

```go
func ServiceLifecycleWorkflow(ctx workflow.Context, serviceName string, event LifecycleEvent) error {
    switch event {
    case ServiceRegistered:
        // 1. 注册服务到 registry
        // 2. 同步网关路由
        // 3. 发布菜单（MenuLifecycleWorkflow 子流程）
        // 4. 加载权限到 IAM/OPA
        // 5. 写审计日志
        // 6. 通知管理员
    case ServiceDeregistered:
        // 1. 标记服务为 DEREGISTERED
        // 2. 下线路由
        // 3. 下线菜单和微应用
        // 4. 标记权限为 UNAVAILABLE
        // 5. 写审计日志
        // 6. 通知管理员
    case ServiceUnhealthy:
        // 1. 标记服务为 UNHEALTHY
        // 2. 下线菜单和微应用（不删除）
        // 3. 标记权限为 UNAVAILABLE
    case ServiceHealthy:
        // 恢复上述状态
    }
}
```

### 2. MenuLifecycleWorkflow

**触发事件**：菜单创建 / 更新 / 上下线请求

**流程**：

```go
func MenuLifecycleWorkflow(ctx workflow.Context, menuID string, target MenuStatus) error {
    // 1. 校验权限存在（iam-service）
    // 2. 同步网关路由/配置（gateway-service）
    // 3. 更新菜单状态（registry-service）
    // 4. 写审计日志
    // 5. 通知前端刷新配置
}
```

### 3. PermissionLifecycleWorkflow

**触发事件**：服务声明/注销权限

**流程**：

```go
func PermissionLifecycleWorkflow(ctx workflow.Context, serviceName string, permissions []Permission) error {
    // 1. 校验权限定义格式
    // 2. 同步到 IAM 服务
    // 3. 同步到 OPA
    // 4. 更新权限状态
    // 5. 审计日志
}
```

### 4. TenantProvisioningWorkflow

**触发事件**：租户开通

**流程**：

```go
func TenantProvisioningWorkflow(ctx workflow.Context, tenantID string) error {
    // 1. 创建租户记录
    // 2. 初始化租户数据/ schema
    // 3. 分配默认角色
    // 4. 注册默认服务
    // 5. 发送欢迎通知
    // 6. 审计日志
}
```

### 5. TenantOffboardingWorkflow

**触发事件**：租户退订/删除

**流程**：归档数据 → 清理权限 → 停止服务 → 删除租户 → 通知。

### 6. RolePermissionApprovalWorkflow

**触发事件**：角色/权限申请

**流程**：提交 → 等待审批 Signal → 批准/拒绝 → 生效/回滚 → 通知。

### 7. ApplicationLifecycleWorkflow

**触发事件**：应用激活/下线

**流程**：应用状态变更 → 级联服务状态 → 级联菜单/微应用 → 通知。

### 8. RoutePolicyChangeWorkflow

**触发事件**：路由策略变更（灰度、限流、鉴权开关）

**流程**：更新策略 → 同步网关 → 验证 → 失败回滚。

## 事件驱动入口

各服务在状态变化时发送事件到 NATS 或调用 Temporal Workflow：

```go
// registry-service 中
func (r *Registry) RegisterService(...) {
    // ... 本地注册
    temporalClient.ExecuteWorkflow(ctx, options, ServiceLifecycleWorkflow, name, ServiceRegistered)
}
```

## 错误处理

- 每个 Activity 配置 RetryPolicy
- Workflow 失败时通过 Temporal UI 查看状态
- 支持 Signal 人工干预（审批、强制回滚）
- 补偿 Activity 用于回滚已完成的步骤

## 验证标准

1. `docker compose up -d` 后 temporal-server、temporal-ui、temporal-worker 都正常运行
2. 注册 demo-service 后，Temporal UI 能看到 ServiceLifecycleWorkflow 成功完成
3. 停止 demo-service 后，菜单和微应用自动变为 OFFLINE，Portal 不再显示
4. 启动 demo-service 后，Workflow 自动恢复菜单和微应用为 ONLINE
5. 权限审批 Workflow 能正确等待 Signal 并生效

## 影响范围

- 新增服务：`platform/temporal-worker`
- 新增基础设施：`temporal-server`, `temporal-ui`
- 修改服务：`registry-service`（状态机、状态字段、事件触发）、`iam-service`、`gateway-service`、`tenant-service`、`audit-service`、`notification-service`
- Portal 前端：过滤 `status = ONLINE / ACTIVE` 的菜单和微应用

## 回滚策略

由于本次是整体架构改造，在 `feat/temporal-lifecycle-state-machines` 分支完成。如果验证不通过，直接废弃分支即可，不影响 `main`。
