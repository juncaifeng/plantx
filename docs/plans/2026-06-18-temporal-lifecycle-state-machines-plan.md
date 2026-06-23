# PlantX Temporal 生命周期状态机实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在 PlantX 平台引入 Temporal 生命周期编排引擎，并在 registry-service 中加入状态机，实现服务注册/注销时菜单、微应用、权限自动上下线。

**Architecture:** Temporal Server + UI 作为基础设施；新增 `platform/temporal-worker` 模块承载 Workflows 和 Activities；`registry-service` 在 Menu/MicroApp/Service 中增加 `status` 字段和内部状态机；服务状态变化时触发 Temporal Workflow 完成级联操作；Portal 前端只展示 `ONLINE` 状态的菜单和微应用。

**Tech Stack:** Temporal (Go SDK), Go + sqlc + protobuf, PostgreSQL, NATS, Docker Compose, React + TypeScript

---

## 前置信息

- 设计文档：`docs/plans/2026-06-18-temporal-lifecycle-state-machines-design.md`
- 当前分支：`feat/temporal-lifecycle-state-machines`
- 关键路径：
  - `deployments/docker-compose/docker-compose.yml`
  - `platform/registry-service/api/registry.proto`
  - `platform/registry-service/internal/domain/registry.go`
  - `platform/registry-service/internal/infra/repo/postgres.go`
  - `platform/registry-service/internal/interfaces/grpc/handler.go`
  - `platform/registry-service/migrations/`
  - `platform/registry-service/internal/infra/sqlc/`
  - `platform/temporal-worker/` (新建)
  - `apps/portal/src/useMenus.ts`, `useMicroApps.ts`, `MicroAppPage.tsx`

---

## Task 1: 添加 Temporal 基础设施到 docker-compose

**Files:**
- Modify: `deployments/docker-compose/docker-compose.yml`

**Step 1: 在 docker-compose.yml 末尾新增 services**

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
```

**Step 2: 验证 compose 配置语法**

```bash
cd deployments/docker-compose && docker compose config > /dev/null
```

Expected: 无错误（可能仍有 version 字段 warning，属于已有问题）。

**Step 3: 提交**

```bash
git add deployments/docker-compose/docker-compose.yml
git commit -m "infra: add temporal-server and temporal-ui to compose"
```

---

## Task 2: 创建 platform/temporal-worker 模块骨架

**Files:**
- Create: `platform/temporal-worker/go.mod`
- Create: `platform/temporal-worker/cmd/main.go`
- Create: `platform/temporal-worker/Dockerfile`
- Create: `platform/temporal-worker/internal/workflows/service_lifecycle.go`
- Create: `platform/temporal-worker/internal/activities/registry.go`
- Create: `platform/temporal-worker/internal/worker/worker.go`

**Step 1: 创建 go.mod**

```go
module github.com/plantx/platform/temporal-worker

go 1.22

require (
	go.temporal.io/sdk v1.26.0
	github.com/plantx/platform/registry-service/api v0.0.0
)

replace github.com/plantx/platform/registry-service/api => ../registry-service/api
```

**Step 2: 创建 worker.go**

```go
package worker

import (
	"context"
	"fmt"

	"github.com/plantx/platform/temporal-worker/internal/activities"
	"github.com/plantx/platform/temporal-worker/internal/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func Start(ctx context.Context, temporalHost string) error {
	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		return fmt.Errorf("dial temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, "plantx-platform", worker.Options{})
	w.RegisterWorkflow(workflows.ServiceLifecycleWorkflow)
	w.RegisterActivity(activities.UpdateMenuStatus)
	w.RegisterActivity(activities.UpdateMicroAppStatus)
	w.RegisterActivity(activities.AuditLog)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("run worker: %w", err)
	}
	return nil
}
```

**Step 3: 创建 cmd/main.go**

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/plantx/platform/temporal-worker/internal/worker"
)

func main() {
	host := os.Getenv("TEMPORAL_HOST")
	if host == "" {
		host = "localhost:7233"
	}
	if err := worker.Start(context.Background(), host); err != nil {
		log.Fatalf("worker failed: %v", err)
	}
}
```

**Step 4: 创建 Dockerfile**

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/temporal-worker ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/temporal-worker .
CMD ["./temporal-worker"]
```

**Step 5: 提交**

```bash
git add platform/temporal-worker/
git commit -m "feat(temporal): scaffold temporal-worker module"
```

---

## Task 3: 在 registry proto 中为 Menu/MicroApp/Service 增加 status

**Files:**
- Modify: `platform/registry-service/api/registry.proto`

**Step 1: 新增枚举和字段**

```protobuf
enum ResourceStatus {
  RESOURCE_STATUS_UNSPECIFIED = 0;
  RESOURCE_STATUS_DRAFT = 1;
  RESOURCE_STATUS_PENDING = 2;
  RESOURCE_STATUS_ONLINE = 3;
  RESOURCE_STATUS_OFFLINE = 4;
  RESOURCE_STATUS_UPDATING = 5;
}

message MicroApp {
  string name = 1;
  string route = 2;
  string bundle_url = 3;
  string menu_label_key = 4;
  string require_permission = 5;
  string application_id = 6;
  string application_key = 7;
  string upstream = 8;
  ResourceStatus status = 9;
}

message Service {
  string id = 1;
  string name = 2;
  string grpc_host = 3;
  string rest_prefix = 4;
  repeated Route routes = 5;
  repeated MicroApp micro_apps = 6;
  string application_id = 7;
  string application_key = 8;
  ResourceStatus status = 9;
}

message Menu {
  string id = 1;
  string label_key = 2;
  string route = 3;
  string icon = 4;
  string parent_id = 5;
  int32 sort_order = 6;
  string micro_app_name = 7;
  string require_permission = 8;
  string application_id = 9;
  string application_key = 10;
  ResourceStatus status = 11;
}
```

**Step 2: 更新 CreateMenuRequest / UpdateMenuRequest**

```protobuf
message CreateMenuRequest {
  string label_key = 1;
  string route = 2;
  string icon = 3;
  string parent_id = 4;
  int32 sort_order = 5;
  string micro_app_name = 6;
  string require_permission = 7;
  string application_id = 8;
  string application_key = 9;
  ResourceStatus status = 10;
}

message UpdateMenuRequest {
  string id = 1;
  string label_key = 2;
  string route = 3;
  string icon = 4;
  string parent_id = 5;
  int32 sort_order = 6;
  string micro_app_name = 7;
  string require_permission = 8;
  string application_id = 9;
  string application_key = 10;
  ResourceStatus status = 11;
}
```

**Step 3: 重新生成 Go 代码**

```bash
buf generate --template buf.go.gen.yaml --path platform/registry-service/api/registry.proto
```

（若生成文件不在 platform 下，手动移动到 `platform/registry-service/api/`）

**Step 4: 提交**

```bash
git add platform/registry-service/api/registry.proto
git commit -m "feat(registry): add ResourceStatus to Menu, MicroApp and Service"
```

---

## Task 4: 数据库迁移添加 status 列

**Files:**
- Create: `platform/registry-service/migrations/007_add_status.up.sql`
- Create: `platform/registry-service/migrations/007_add_status.down.sql`

**Step 1: up migration**

```sql
ALTER TABLE micro_apps ADD COLUMN status TEXT NOT NULL DEFAULT 'ONLINE';
ALTER TABLE menus ADD COLUMN status TEXT NOT NULL DEFAULT 'ONLINE';
ALTER TABLE registry_services ADD COLUMN status TEXT NOT NULL DEFAULT 'ONLINE';
```

**Step 2: down migration**

```sql
ALTER TABLE micro_apps DROP COLUMN status;
ALTER TABLE menus DROP COLUMN status;
ALTER TABLE registry_services DROP COLUMN status;
```

**Step 3: 提交**

```bash
git add platform/registry-service/migrations/
git commit -m "feat(registry): add status columns for Menu, MicroApp and Service"
```

---

## Task 5: 更新 sqlc queries 包含 status

**Files:**
- Modify: `platform/registry-service/internal/infra/sqlc/queries.sql`

**Step 1: 修改所有涉及 micro_apps / menus / services 的 query**

- `UpsertService` 增加 `status`
- `UpsertMicroApp` 增加 `status`
- `UpdateMicroApp` 增加 `status`
- `CreateMenu` 增加 `status`
- `UpdateMenu` 增加 `status`

例如：

```sql
-- name: UpsertService :one
INSERT INTO registry_services (name, grpc_host, rest_prefix, application_id, status)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (name) DO UPDATE SET
    grpc_host = EXCLUDED.grpc_host,
    rest_prefix = EXCLUDED.rest_prefix,
    application_id = EXCLUDED.application_id,
    status = EXCLUDED.status,
    updated_at = now()
RETURNING *;
```

**Step 2: 重新生成 sqlc**

```bash
cd E:/git/plantx && sqlc generate
```

**Step 3: 提交**

```bash
git add platform/registry-service/internal/infra/sqlc/queries.sql
git commit -m "feat(registry): include status in sqlc queries"
```

---

## Task 6: 更新 registry-service domain 状态机

**Files:**
- Create: `platform/registry-service/internal/domain/status.go`
- Modify: `platform/registry-service/internal/domain/registry.go`

**Step 1: 创建 status.go**

```go
package domain

type ResourceStatus string

const (
	ResourceStatusDraft    ResourceStatus = "DRAFT"
	ResourceStatusPending  ResourceStatus = "PENDING"
	ResourceStatusOnline   ResourceStatus = "ONLINE"
	ResourceStatusOffline  ResourceStatus = "OFFLINE"
	ResourceStatusUpdating ResourceStatus = "UPDATING"
)

var menuStatusTransitions = map[ResourceStatus][]ResourceStatus{
	ResourceStatusDraft:    {ResourceStatusPending},
	ResourceStatusPending:  {ResourceStatusOnline, ResourceStatusOffline, ResourceStatusDraft},
	ResourceStatusOnline:   {ResourceStatusUpdating, ResourceStatusOffline},
	ResourceStatusUpdating: {ResourceStatusOnline, ResourceStatusOffline},
	ResourceStatusOffline:  {ResourceStatusPending, ResourceStatusDraft},
}

func CanTransitionMenu(from, to ResourceStatus) bool {
	if from == "" || from == to {
		return true
	}
	allowed, ok := menuStatusTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
```

**Step 2: 修改 domain/registry.go**

在 `MicroApp`、`Service`、`Menu` 结构体中增加 `Status ResourceStatus`。

**Step 3: 提交**

```bash
git add platform/registry-service/internal/domain/
git commit -m "feat(registry): add ResourceStatus and menu transition rules"
```

---

## Task 7: 更新 repo 层传递 status

**Files:**
- Modify: `platform/registry-service/internal/infra/repo/postgres.go`

**Step 1: 修改所有 mapper 和 query 调用**

- `UpsertService` / `serviceToDomain` 处理 status
- `UpsertMicroApp` / `UpdateMicroApp` / `toDomainMicroApp` 处理 status
- `CreateMenu` / `UpdateMenu` / `toDomainMenu` 处理 status

例如：

```go
func toDomainMicroApp(row sqlc.MicroApp) *domain.MicroApp {
	m := &domain.MicroApp{
		Name:              row.Name,
		Route:             row.Route,
		BundleURL:         row.BundleUrl,
		MenuLabelKey:      row.MenuLabelKey,
		RequirePermission: row.RequirePermission,
		Status:            domain.ResourceStatus(row.Status),
	}
	// ...
}
```

**Step 2: 提交**

```bash
git add platform/registry-service/internal/infra/repo/postgres.go
git commit -m "feat(registry): persist and load resource status"
```

---

## Task 8: 更新 grpc handler 透传 status

**Files:**
- Modify: `platform/registry-service/internal/interfaces/grpc/handler.go`

**Step 1: 修改 toProto / toDomain mapper**

- `toProtoMicroApp` 增加 `Status`
- `toDomainMicroApp` 增加 `Status`
- `toProtoService` 增加 `Status`
- `toProtoMenu` 增加 `Status`
- `toDomainMenu` 增加 `Status`

**Step 2: 编译 registry-service**

```bash
cd platform/registry-service && go build ./...
```

Expected: 无错误。

**Step 3: 提交**

```bash
git add platform/registry-service/internal/interfaces/grpc/handler.go
git commit -m "feat(registry): expose status in grpc handler"
```

---

## Task 9: 实现 temporal-worker 的 ServiceLifecycleWorkflow

**Files:**
- Modify: `platform/temporal-worker/internal/workflows/service_lifecycle.go`
- Modify: `platform/temporal-worker/internal/activities/registry.go`

**Step 1: 创建 ServiceLifecycleWorkflow**

```go
package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type LifecycleEvent string

const (
	ServiceRegistered   LifecycleEvent = "REGISTERED"
	ServiceDeregistered LifecycleEvent = "DEREGISTERED"
	ServiceUnhealthy    LifecycleEvent = "UNHEALTHY"
	ServiceHealthy      LifecycleEvent = "HEALTHY"
)

func ServiceLifecycleWorkflow(ctx workflow.Context, serviceName string, event LifecycleEvent) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	switch event {
	case ServiceRegistered, ServiceHealthy:
		// 上线流程
		if err := workflow.ExecuteActivity(ctx, "SetServiceStatus", serviceName, "ONLINE").Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "PublishServiceMenus", serviceName).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "PublishServiceMicroApps", serviceName).Get(ctx, nil); err != nil {
			return err
		}
	case ServiceDeregistered, ServiceUnhealthy:
		// 下线流程
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMenus", serviceName).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMicroApps", serviceName).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "SetServiceStatus", serviceName, "OFFLINE").Get(ctx, nil); err != nil {
			return err
		}
	}

	return workflow.ExecuteActivity(ctx, "WriteAuditLog", serviceName, string(event)).Get(ctx, nil)
}
```

**Step 2: 创建 Activities**

```go
package activities

import (
	"context"

	registryapi "github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
)

type RegistryActivities struct {
	RegistryAddr string
}

func (a *RegistryActivities) getClient(ctx context.Context) (registryapi.RegistryServiceClient, error) {
	conn, err := grpc.Dial(a.RegistryAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return registryapi.NewRegistryServiceClient(conn), nil
}

func (a *RegistryActivities) PublishServiceMenus(ctx context.Context, serviceName string) error {
	// 调用 registry gRPC 更新该服务下所有菜单为 ONLINE
	return nil
}

func (a *RegistryActivities) UnpublishServiceMenus(ctx context.Context, serviceName string) error {
	// 调用 registry gRPC 更新该服务下所有菜单为 OFFLINE
	return nil
}

func (a *RegistryActivities) PublishServiceMicroApps(ctx context.Context, serviceName string) error {
	return nil
}

func (a *RegistryActivities) UnpublishServiceMicroApps(ctx context.Context, serviceName string) error {
	return nil
}

func (a *RegistryActivities) SetServiceStatus(ctx context.Context, serviceName, status string) error {
	return nil
}

func (a *RegistryActivities) WriteAuditLog(ctx context.Context, serviceName, event string) error {
	return nil
}
```

**Step 3: 提交**

```bash
git add platform/temporal-worker/
git commit -m "feat(temporal): add ServiceLifecycleWorkflow and registry activities"
```

---

## Task 10: 在 docker-compose 中添加 temporal-worker 服务

**Files:**
- Modify: `deployments/docker-compose/docker-compose.yml`

**Step 1: 添加 temporal-worker service**

```yaml
  temporal-worker:
    build:
      context: ../..
      dockerfile: platform/temporal-worker/Dockerfile
    container_name: docker-compose-temporal-worker-1
    environment:
      - TEMPORAL_HOST=temporal-server:7233
      - REGISTRY_SERVICE_GRPC_ADDR=registry-service:8080
    depends_on:
      - temporal-server
      - registry-service
    networks:
      - plantx
```

**Step 2: 提交**

```bash
git add deployments/docker-compose/docker-compose.yml
git commit -m "infra: add temporal-worker to compose"
```

---

## Task 11: 在 registry-service 中触发 Temporal Workflow

**Files:**
- Create: `platform/registry-service/internal/infra/temporal/client.go`
- Modify: `platform/registry-service/internal/app/registry.go`

**Step 1: 创建 Temporal client wrapper**

```go
package temporal

import (
	"context"
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

func NewClient() (client.Client, error) {
	host := os.Getenv("TEMPORAL_HOST")
	if host == "" {
		host = "localhost:7233"
	}
	return client.Dial(client.Options{HostPort: host})
}
```

**Step 2: 修改 app/registry.go 的 RegisterService / DeregisterService**

```go
func (r *Registry) RegisterService(ctx context.Context, name, grpcHost, restPrefix, applicationID string) (*domain.Service, error) {
	svc, err := r.repo.RegisterService(ctx, name, grpcHost, restPrefix, applicationID)
	if err != nil {
		return nil, err
	}
	// 触发 Temporal Workflow
	if r.temporalClient != nil {
		_, _ = r.temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
			ID:        fmt.Sprintf("service-lifecycle-%s-registered", name),
			TaskQueue: "plantx-platform",
		}, "ServiceLifecycleWorkflow", name, "REGISTERED")
	}
	return svc, nil
}
```

**Step 3: 提交**

```bash
git add platform/registry-service/internal/infra/temporal/client.go
git add platform/registry-service/internal/app/registry.go
git commit -m "feat(registry): trigger Temporal workflow on service registration"
```

---

## Task 12: 实现 Activity 真正调用 registry gRPC

**Files:**
- Modify: `platform/temporal-worker/internal/activities/registry.go`

**Step 1: 实现 Publish/Unpublish 逻辑**

```go
func (a *RegistryActivities) UnpublishServiceMenus(ctx context.Context, serviceName string) error {
	conn, err := grpc.Dial(a.RegistryAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := registryapi.NewRegistryServiceClient(conn)

	// 1. 通过 service name 找到 service id
	// 2. 列出该 service 的菜单
	// 3. 逐个更新为 OFFLINE
	return nil
}
```

由于当前 registry proto 没有 "ListMenusByServiceID" 或 "UpdateMenuStatus" 便捷接口，可能需要新增 `UpdateMenuStatus` RPC 或 `BatchUpdateMenuStatus` RPC。

**Step 2: 提交**

```bash
git add platform/temporal-worker/internal/activities/registry.go
git commit -m "feat(temporal): implement registry activities for menu/micro-app status"
```

---

## Task 13: 新增 BatchUpdateMenuStatus / BatchUpdateMicroAppStatus RPC（如需要）

**Files:**
- Modify: `platform/registry-service/api/registry.proto`
- Modify: `platform/registry-service/internal/interfaces/grpc/handler.go`
- Modify: `platform/registry-service/internal/app/registry.go`
- Modify: `platform/registry-service/internal/domain/registry.go`
- Modify: `platform/registry-service/internal/infra/repo/postgres.go`

**Step 1: proto 新增 RPC**

```protobuf
rpc BatchUpdateMenuStatus(BatchUpdateMenuStatusRequest) returns (MenuList) {
  option (google.api.http) = {
    post: "/api/registry/v1/menus/batch-status"
    body: "*"
  };
}

message BatchUpdateMenuStatusRequest {
  repeated string ids = 1;
  ResourceStatus status = 2;
}
```

同理 `BatchUpdateMicroAppStatus`。

**Step 2: 实现 repo/app/handler**

**Step 3: 重新生成代码并编译**

**Step 4: 提交**

```bash
git add platform/registry-service/
git commit -m "feat(registry): add batch status update RPCs for menus and micro-apps"
```

---

## Task 14: 更新 Portal 前端过滤状态

**Files:**
- Modify: `apps/portal/src/useMenus.ts`
- Modify: `apps/portal/src/useMicroApps.ts`
- Modify: `apps/portal/src/MicroAppPage.tsx`（或对应组件）

**Step 1: useMenus.ts 过滤 ONLINE**

```ts
const activeMenus = data?.menus?.filter((m) => m.status === 'ONLINE') ?? [];
```

**Step 2: useMicroApps.ts 过滤 ONLINE**

```ts
const activeMicroApps = data?.microApps?.filter((m) => m.status === 'ONLINE') ?? [];
```

**Step 3: 提交**

```bash
git add apps/portal/src/useMenus.ts apps/portal/src/useMicroApps.ts
git commit -m "feat(portal): only show ONLINE menus and micro-apps"
```

---

## Task 15: 构建并运行验证

**Step 1: 构建 temporal-worker**

```bash
cd platform/temporal-worker
go build ./...
```

**Step 2: 启动平台**

```bash
cd deployments/docker-compose
docker compose down -v
docker compose up -d --build
```

**Step 3: 等待 Temporal 启动**

```bash
docker logs -f docker-compose-temporal-server-1
```

Expected: Server started。

**Step 4: 访问 Temporal UI**

```bash
open http://localhost:8083
```

**Step 5: 注册 demo-service**

```bash
curl -X POST http://localhost/api/registry/v1/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"demo-service","grpc_host":"demo-service:8080","rest_prefix":"/api/demo/v1"}'
```

**Step 6: 检查 Temporal UI Workflow**

Expected: 看到 `ServiceLifecycleWorkflow` 成功执行。

**Step 7: 停止 demo-service**

```bash
cd deployments/docker-compose && docker compose stop demo-service
```

**Step 8: 检查菜单状态**

```bash
curl -s http://localhost/api/registry/v1/menus | jq
```

Expected: demo 相关菜单 status 变为 `OFFLINE`。

**Step 9: 启动 demo-service**

```bash
cd deployments/docker-compose && docker compose start demo-service
```

Expected: 菜单恢复 `ONLINE`。

---

## Task 16: 文档更新

**Files:**
- Modify: `docs/plans/2026-06-18-temporal-lifecycle-state-machines-design.md`

**Step 1: 补充实际验证结果**

**Step 2: 提交**

```bash
git add docs/plans/
git commit -m "docs: record Temporal integration verification results"
```

---

## 回滚

```bash
git checkout main
# 删除 feat/temporal-lifecycle-state-machines 分支即可
```
