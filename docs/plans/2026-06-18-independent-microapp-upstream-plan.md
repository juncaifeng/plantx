# 独立微前端 Upstream 支持实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 `MicroApp` 支持指定独立 `upstream`，业务团队微前端可独立部署，Demo 不再依赖把 bundle 复制进 Portal 容器。

**Architecture:** 在 registry proto 与数据库中新增 `upstream` 字段；`nginx-sync.sh` 根据该字段把 `/apps/<name>/` 代理到指定主机；Demo 增加独立 `demo-ui` nginx 容器并注册 `upstream: "demo-ui:80"`。

**Tech Stack:** buf + protobuf, Go + sqlc, shell + nginx, Docker Compose, React + qiankun

---

## 前置信息

- 设计文档：`docs/plans/2026-06-18-independent-microapp-upstream-design.md`
- 关键文件路径：
  - proto: `platform/registry-service/api/registry.proto`
  - 生成代码: `platform/registry-service/api/*.pb.go`, `platform/registry-service/api/*.gw.pb.go`
  - DB migration: `platform/registry-service/migrations/`
  - sqlc: `platform/registry-service/internal/infra/sqlc/`
  - repo/handler: `platform/registry-service/internal/infra/repo/postgres.go`, `platform/registry-service/internal/interfaces/grpc/handler.go`
  - nginx sync: `deployments/docker-compose/nginx-sync.sh`
  - demo backend: `demo_app/backend/cmd/main.go`
  - demo frontend: `demo_app/frontend/`
  - compose: `deployments/docker-compose/docker-compose.yml`

---

## Task 1: 修改 registry proto，新增 upstream 字段

**Files:**
- Modify: `platform/registry-service/api/registry.proto:308-316`

**Step 1: 编辑 proto**

在 `MicroApp` 消息中新增 `upstream` 字段：

```protobuf
message MicroApp {
  string name = 1;
  string route = 2;
  string bundle_url = 3;
  string menu_label_key = 4;
  string require_permission = 5;
  string application_id = 6;
  string application_key = 7;
  string upstream = 8;
}
```

**Step 2: 更新 Register/Update 请求**

`RegisterMicroAppRequest` 已经通过嵌套 `MicroApp` 透传所有字段，无需改动。

`UpdateMicroAppRequest` 当前是逐个字段，需要新增 `upstream`：

```protobuf
message UpdateMicroAppRequest {
  string name = 1;
  string route = 2;
  string bundle_url = 3;
  string menu_label_key = 4;
  string require_permission = 5;
  string upstream = 6;
}
```

**Step 3: 提交**

```bash
git add platform/registry-service/api/registry.proto
git commit -m "feat(registry): add upstream field to MicroApp proto"
```

---

## Task 2: 重新生成 registry Go 代码

**Files:**
- Generated: `platform/registry-service/api/registry.pb.go`
- Generated: `platform/registry-service/api/registry.pb.gw.go`

**Step 1: 运行 buf generate**

```bash
cd E:/git/plantx
buf generate --template buf.gen.yaml platform/registry-service/api/registry.proto
```

或运行项目已有生成脚本：

```bash
make generate
```

**Step 2: 验证生成文件包含 upstream 字段**

```bash
grep -n "Upstream" platform/registry-service/api/registry.pb.go | head -20
```

Expected: 能看到 `GetUpstream()`, `Upstream` 字段等。

**Step 3: 提交**

```bash
git add platform/registry-service/api/
git commit -m "chore(registry): regenerate Go code for upstream field"
```

---

## Task 3: 数据库 migration 添加 upstream 列

**Files:**
- Create: `platform/registry-service/migrations/006_micro_apps_upstream.up.sql`
- Create: `platform/registry-service/migrations/006_micro_apps_upstream.down.sql`

**Step 1: 创建 up migration**

```sql
ALTER TABLE micro_apps ADD COLUMN upstream TEXT;
```

**Step 2: 创建 down migration**

```sql
ALTER TABLE micro_apps DROP COLUMN upstream;
```

**Step 3: 提交**

```bash
git add platform/registry-service/migrations/
git commit -m "feat(registry): add upstream column to micro_apps"
```

---

## Task 4: 更新 sqlc queries

**Files:**
- Modify: `platform/registry-service/internal/infra/sqlc/queries.sql:23-40`
- Modify: `platform/registry-service/internal/infra/sqlc/queries.sql:51-59`
- Generated: `platform/registry-service/internal/infra/sqlc/*.sql.go`

**Step 1: 修改 UpsertMicroApp query**

```sql
-- name: UpsertMicroApp :one
INSERT INTO micro_apps (
    service_id,
    name,
    route,
    bundle_url,
    menu_label_key,
    require_permission,
    application_id,
    upstream
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (service_id, name) DO UPDATE SET
    route = EXCLUDED.route,
    bundle_url = EXCLUDED.bundle_url,
    menu_label_key = EXCLUDED.menu_label_key,
    require_permission = EXCLUDED.require_permission,
    application_id = EXCLUDED.application_id,
    upstream = EXCLUDED.upstream,
    updated_at = now()
RETURNING *;
```

**Step 2: 修改 UpdateMicroApp query**

```sql
-- name: UpdateMicroApp :one
UPDATE micro_apps SET
    route = $2,
    bundle_url = $3,
    menu_label_key = $4,
    require_permission = $5,
    upstream = $6,
    updated_at = now()
WHERE name = $1
RETURNING *;
```

**Step 3: 重新生成 sqlc**

```bash
cd platform/registry-service
sqlc generate
```

如果没有 sqlc CLI，安装：`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

**Step 4: 验证生成代码**

```bash
grep -n "Upstream" internal/infra/sqlc/*.sql.go | head -20
```

**Step 5: 提交**

```bash
git add platform/registry-service/internal/infra/sqlc/
git commit -m "feat(registry): include upstream in micro-app sqlc queries"
```

---

## Task 5: 更新 domain model

**Files:**
- Modify: `platform/registry-service/internal/domain/registry.go:32-40`

**Step 1: 在 MicroApp 结构体增加 Upstream 字段**

```go
type MicroApp struct {
	Name              string
	Route             string
	BundleURL         string
	MenuLabelKey      string
	RequirePermission string
	ApplicationID     string
	ApplicationKey    string
	Upstream          string
}
```

**Step 2: 提交**

```bash
git add platform/registry-service/internal/domain/registry.go
git commit -m "feat(registry): add Upstream to MicroApp domain model"
```

---

## Task 6: 更新 repo mapper

**Files:**
- Modify: `platform/registry-service/internal/infra/repo/postgres.go:246-258`
- Modify: `platform/registry-service/internal/infra/repo/postgres.go:287-298`
- Modify: `platform/registry-service/internal/infra/repo/postgres.go:532-544`

**Step 1: UpsertMicroApp 增加 Upstream**

```go
row, err := r.queries.UpsertMicroApp(ctx, sqlc.UpsertMicroAppParams{
	ServiceID:         svc.ID,
	Name:              name,
	Route:             microApp.Route,
	BundleUrl:         microApp.BundleURL,
	MenuLabelKey:      microApp.MenuLabelKey,
	RequirePermission: microApp.RequirePermission,
	ApplicationID:     appID,
	Upstream:          nullString(microApp.Upstream),
})
```

**Step 2: UpdateMicroApp 增加 Upstream**

```go
row, err := r.queries.UpdateMicroApp(ctx, sqlc.UpdateMicroAppParams{
	Name:              name,
	Route:             microApp.Route,
	BundleUrl:         microApp.BundleURL,
	MenuLabelKey:      microApp.MenuLabelKey,
	RequirePermission: microApp.RequirePermission,
	Upstream:          nullString(microApp.Upstream),
})
```

**Step 3: toDomainMicroApp 增加 Upstream**

```go
func toDomainMicroApp(row sqlc.MicroApp) *domain.MicroApp {
	m := &domain.MicroApp{
		Name:              row.Name,
		Route:             row.Route,
		BundleURL:         row.BundleUrl,
		MenuLabelKey:      row.MenuLabelKey,
		RequirePermission: row.RequirePermission,
	}
	if row.ApplicationID.Valid {
		m.ApplicationID = row.ApplicationID.UUID.String()
	}
	if row.Upstream.Valid {
		m.Upstream = row.Upstream.String
	}
	return m
}
```

**Step 4: 提交**

```bash
git add platform/registry-service/internal/infra/repo/postgres.go
git commit -m "feat(registry): persist and load micro-app upstream"
```

---

## Task 7: 更新 grpc handler mapper

**Files:**
- Modify: `platform/registry-service/internal/interfaces/grpc/handler.go:159-171`
- Modify: `platform/registry-service/internal/interfaces/grpc/handler.go:380-411`

**Step 1: UpdateMicroApp 透传 upstream**

```go
func (h *Handler) UpdateMicroApp(ctx context.Context, req *api.UpdateMicroAppRequest) (*api.MicroApp, error) {
	app, err := h.registry.UpdateMicroApp(ctx, req.GetName(), toDomainMicroApp(&api.MicroApp{
		Name:              req.GetName(),
		Route:             req.GetRoute(),
		BundleUrl:         req.GetBundleUrl(),
		MenuLabelKey:      req.GetMenuLabelKey(),
		RequirePermission: req.GetRequirePermission(),
		Upstream:          req.GetUpstream(),
	}, ""))
	if err != nil {
		return nil, err
	}
	return toProtoMicroApp(app), nil
}
```

**Step 2: toDomainMicroApp 透传 upstream**

```go
func toDomainMicroApp(m *api.MicroApp, applicationID string) *domain.MicroApp {
	if m == nil {
		return nil
	}
	appID := applicationID
	if appID == "" {
		appID = m.GetApplicationId()
	}
	return &domain.MicroApp{
		Name:              m.GetName(),
		Route:             m.GetRoute(),
		BundleURL:         m.GetBundleUrl(),
		MenuLabelKey:      m.GetMenuLabelKey(),
		RequirePermission: m.GetRequirePermission(),
		ApplicationID:     appID,
		Upstream:          m.GetUpstream(),
	}
}
```

**Step 3: toProtoMicroApp 透传 upstream**

```go
func toProtoMicroApp(m *domain.MicroApp) *api.MicroApp {
	if m == nil {
		return nil
	}
	return &api.MicroApp{
		Name:              m.Name,
		Route:             m.Route,
		BundleUrl:         m.BundleURL,
		MenuLabelKey:      m.MenuLabelKey,
		RequirePermission: m.RequirePermission,
		ApplicationId:     m.ApplicationID,
		ApplicationKey:    m.ApplicationKey,
		Upstream:          m.Upstream,
	}
}
```

**Step 4: 提交**

```bash
git add platform/registry-service/internal/interfaces/grpc/handler.go
git commit -m "feat(registry): expose upstream in grpc handler"
```

---

## Task 8: 编译 registry-service

**Files:**
- Build: `platform/registry-service/`

**Step 1: 编译验证**

```bash
cd platform/registry-service
go build ./...
```

Expected: 无错误。

**Step 2: 提交**

```bash
git commit --allow-empty -m "test(registry): compile check passed"
```

---

## Task 9: 修改 nginx-sync.sh 支持独立 upstream

**Files:**
- Modify: `deployments/docker-compose/nginx-sync.sh:54-89`

**Step 1: 修改 upstreams.conf 生成**

将第 54-59 行改为同时生成 micro-app 的 upstream：

```sh
# Upstreams
{
  echo "upstream mock_auth { server mock-auth:8080; }"
  echo "upstream portal { server portal:80; }"
  echo "$routes_json" | jq -r '.routes[]? | "upstream \(.name | gsub("-"; "_")) { server \(.upstreamHost); }"'
  echo "$micro_apps_json" | jq -r '
    .microApps[]? |
    select(.upstream and (.upstream | length) > 0) |
    "upstream \(.name | gsub("-"; "_")) { server \(.upstream); }"
  '
} > /etc/nginx/conf.d/upstreams.conf
```

**Step 2: 修改 micro-app locations 生成**

将第 80-89 行改为根据 upstream 选择 proxy target：

```sh
# Micro-app locations
{
  echo "$micro_apps_json" | jq -r '
    .microApps[]? |
    (if (.upstream // "") | length > 0 then "\(.name | gsub("-"; "_"))" else "portal" end) as $target |
    "    location \(.bundleUrl | sub("/[^/]+$"; "/")) {\n" +
    "        proxy_pass http://\($target)\(.bundleUrl | sub("/[^/]+$"; "/"));\n" +
    "        proxy_set_header Host $host;\n" +
    "    }"
  '
} > /etc/nginx/conf.d/micro-app-locations.conf
```

**Step 3: 本地测试脚本语法**

```bash
sh -n deployments/docker-compose/nginx-sync.sh
```

**Step 4: 提交**

```bash
git add deployments/docker-compose/nginx-sync.sh
git commit -m "feat(gateway): proxy micro-apps to custom upstream when set"
```

---

## Task 10: 为 demo-ui 创建独立 Dockerfile 和 nginx 配置

**Files:**
- Create: `demo_app/frontend/Dockerfile`
- Create: `demo_app/frontend/nginx.conf`

**Step 1: 创建 Dockerfile**

```dockerfile
FROM nginx:alpine
COPY dist/demo-ui.iife.js /usr/share/nginx/html/apps/demo-ui/demo-ui.js
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**Step 2: 创建 nginx.conf**

```nginx
server {
    listen 80;
    server_name localhost;

    location /apps/demo-ui/ {
        root /usr/share/nginx/html;
        try_files $uri $uri/ =404;
        add_header Content-Type application/javascript;
    }

    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ =404;
    }
}
```

**Step 3: 提交**

```bash
git add demo_app/frontend/Dockerfile demo_app/frontend/nginx.conf
git commit -m "feat(demo): add standalone demo-ui container"
```

---

## Task 11: 在 docker-compose 中增加 demo-ui 服务

**Files:**
- Modify: `deployments/docker-compose/docker-compose.yml`

**Step 1: 找到 demo-service 服务定义，在其后添加 demo-ui**

```yaml
  demo-ui:
    build:
      context: ../../demo_app/frontend
      dockerfile: Dockerfile
    container_name: docker-compose-demo-ui-1
    networks:
      - plantx
    restart: unless-stopped
```

**Step 2: 提交**

```bash
git add deployments/docker-compose/docker-compose.yml
git commit -m "feat(demo): add demo-ui service to compose"
```

---

## Task 12: 修改 demo-service 注册微应用时设置 upstream

**Files:**
- Modify: `demo_app/backend/cmd/main.go`

**Step 1: 在 seed/registration 代码中设置 upstream**

找到注册 micro app 的代码，例如：

```go
microApp := &registry.MicroApp{
    Name:              "demo-ui",
    Route:             "/demo/*",
    BundleUrl:         "/apps/demo-ui/demo-ui.js",
    MenuLabelKey:      "nav.demo.home",
    RequirePermission: "demo:item:read",
    ApplicationId:     appID,
    Upstream:          "demo-ui:80",
}
```

**Step 2: 提交**

```bash
git add demo_app/backend/cmd/main.go
git commit -m "feat(demo): register demo-ui with independent upstream"
```

---

## Task 13: 清理 Portal 中的 demo-ui 临时文件

**Files:**
- None (runtime cleanup)

**Step 1: 删除 Portal 容器内的 demo-ui 文件**

```bash
docker exec docker-compose-portal-1 rm -rf /usr/share/nginx/html/apps/demo-ui
docker exec docker-compose-portal-1 nginx -s reload
```

**Step 2: 检查是否还有残留**

```bash
docker exec docker-compose-portal-1 ls -la /usr/share/nginx/html/apps/
```

Expected: 没有 `demo-ui` 目录。

---

## Task 14: 重建并验证

**Step 1: 停止并重建相关服务**

```bash
cd deployments/docker-compose
docker compose down -v
docker compose up -d --build
```

**Step 2: 等待服务启动**

```bash
docker compose ps
```

Expected: `demo-service` 和 `demo-ui` 都 healthy / running。

**Step 3: 检查 registry 中的 micro-app**

```bash
curl -s http://localhost/api/registry/v1/micro-apps | jq
```

Expected: `demo-ui` 的 `upstream` 为 `"demo-ui:80"`，`bundleUrl` 为 `"/apps/demo-ui/demo-ui.js"`。

**Step 4: 检查 nginx 配置**

```bash
docker exec docker-compose-gateway-1 cat /etc/nginx/conf.d/micro-app-locations.conf | grep -A5 demo-ui
docker exec docker-compose-gateway-1 cat /etc/nginx/conf.d/upstreams.conf | grep demo_ui
```

Expected: 看到 `upstream demo_ui { server demo-ui:80; }` 和 `location /apps/demo-ui/ { proxy_pass http://demo_ui/apps/demo-ui/; ... }`。

**Step 5: 浏览器验证 Portal Demo 页面**

访问 `http://localhost`，登录后进入 Demo 应用，确认页面正常。

**Step 6: 验证停止 demo-ui 后页面无法加载**

```bash
docker compose stop demo-ui
```

刷新 Portal 中的 Demo 页面，Expected: 页面空白/404/网络错误，不再能显示 Demo 内容。

**Step 7: 恢复 demo-ui**

```bash
docker compose start demo-ui
```

---

## Task 15: 文档更新

**Files:**
- Modify: `demo_app/README.md` 或创建 `demo_app/ARCHITECTURE.md`

**Step 1: 说明独立部署方式**

补充说明：
- demo-ui 是独立容器
- demo-service 通过 registry 注册 `upstream`
- Portal 通过 qiankun 加载独立部署的 bundle

**Step 2: 提交**

```bash
git add demo_app/README.md
git commit -m "docs(demo): document independent micro-frontend deployment"
```

---

## 回滚/清理命令

```bash
cd deployments/docker-compose
docker compose down -v
```

---

## 已知限制

- 服务停止不会自动清理 registry 中的菜单记录，需单独设计生命周期管理。
- `upstream` 当前限定为 `host:port`，暂不支持带 scheme 的完整 URL。
