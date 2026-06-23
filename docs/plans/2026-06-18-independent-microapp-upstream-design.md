# 独立微前端 Upstream 支持设计

## 目标

让 PlantX 平台的 `MicroApp` 支持指定独立的 `upstream` 主机。业务团队开发的微前端可以独立部署、独立 serve bundle，不再需要在构建期将 JS 文件打包进 Portal 容器。

## 背景

当前平台（`apps/portal/Dockerfile` 与 `deployments/docker-compose/nginx-sync.sh`）把所有微前端在构建期复制进 Portal 的 `dist`，网关统一把 `/apps/<name>/` 代理回 Portal。这导致：

- 业务团队无法独立部署前端。
- 停止后端服务后，只要 Portal 中还残留 bundle 文件，前端页面仍然能打开。
- 不符合 qiankun 独立微前端的架构理念。

## 设计

### 1. 数据模型

在 `platform/registry-service/api/registry.proto` 的 `MicroApp` 消息中新增可选字段：

```protobuf
message MicroApp {
  string name = 1;
  string route = 2;
  string bundle_url = 3;
  string menu_label_key = 4;
  string require_permission = 5;
  string application_id = 6;
  string application_key = 7;
  string upstream = 8;   // NEW: 独立部署时指向 bundle 所在主机，如 "demo-ui:80"
}
```

`RegisterMicroAppRequest` / `UpdateMicroAppRequest` 透传该字段。

### 2. 持久化

- 新增 migration：`ALTER TABLE micro_apps ADD COLUMN upstream TEXT;`
- 更新 `platform/registry-service/internal/infra/sqlc/queries.sql`：
  - `UpsertMicroApp` 增加 `upstream`
  - `UpdateMicroApp` 增加 `upstream`
- 更新 domain、repo mapper、grpc handler mapper，使 `upstream` 在 proto / DB / domain 之间正确传递。

### 3. 网关路由

修改 `deployments/docker-compose/nginx-sync.sh` 的微应用 location 生成逻辑：

- 若 `upstream` 为空，保持现状：`proxy_pass http://portal<path>;`
- 若 `upstream` 非空，先生成 `upstream <sanitized_name> { server <upstream>; }`，再生成 `proxy_pass http://<sanitized_name><path>;`

其中 `<sanitized_name>` 为 `name | gsub("-"; "_")`，与已有 service upstream 命名规则一致。

### 4. Demo 改造

- 新增 `demo_app/frontend/Dockerfile`：基于 `nginx:alpine` 独立 serve `demo-ui.iife.js`。
- 新增 `demo_app/frontend/nginx.conf`。
- 在 `deployments/docker-compose/docker-compose.yml` 增加 `demo-ui` 服务。
- 修改 `demo_app/backend/cmd/main.go` 注册微应用时设置：
  - `bundle_url: "/apps/demo-ui/demo-ui.js"`
  - `upstream: "demo-ui:80"`
- 删除之前把 `demo-ui.iife.js` 复制到 Portal 容器的临时 hack。

### 5. 向后兼容

- `upstream` 为可选字段，未设置时行为与改造前完全一致。
- Portal Dockerfile 继续可以托管平台内部微应用（order-ui、iam-admin-ui 等），无需改动。

## 验证标准

1. 启动平台 + `demo-service` + `demo-ui`。
2. Portal 中 Demo 菜单、页面正常显示。
3. 停止 `demo-ui` 容器后刷新 Portal，Demo 页面应无法加载（404 或网络错误）。
4. 停止 `demo-service` 后，菜单记录仍保留在 registry 中（菜单生命周期属于另一个独立话题）。

## 未包含在本设计中的问题

- 服务停止时是否自动清理注册的菜单/微应用：当前 `DeregisterService` 只删除 service 记录，菜单不会级联删除。如需支持，需要单独设计。
- `upstream` 格式：本设计限定为 `host:port`（与现有 service `grpc_host` / `upstreamHost` 风格一致），暂不支持带 scheme 的完整 URL 或 CDN 外链。
