## 1. kit-go 统一服务注册

- [x] 1.1 新增 `kit/kit-go/gateway/client.go`，封装 `gateway-service` 的 gRPC 客户端。
- [x] 1.2 新增 `kit/kit-go/gateway/register.go`，实现 `AutoRegister(serviceName string, opts ...Option)`。
- [x] 1.3 修改 `kit/kit-go/server/server.go`，新增 `ServiceName` 和 `GatewayRegistrar` 选项，启动后自动注册、关闭时注销。
- [x] 1.4 修改 `services/order/cmd/main.go`，启用 `GatewayRegistrar` 验证已有服务。
- [ ] 1.5 验证 `gateway-service` `/api/gateway/v1/services` 包含 `order-service`。

## 2. gateway-service 微前端注册能力

- [x] 2.1 扩展 `platform/gateway-service/api/gateway.proto`，在 `Service` 消息中加入 `MicroApp` 元数据字段。
- [x] 2.2 在 `gateway.proto` 中新增 `RegisterMicroApp` 和 `ListMicroApps` RPC。
- [x] 2.3 重新生成 `gateway.pb.go`、`gateway_grpc.pb.go`、`gateway.pb.gw.go`。
- [x] 2.4 修改 `platform/gateway-service/cmd/main.go`，实现微前端注册表与查询 API。
- [x] 2.5 验证 `GET /api/gateway/v1/micro-apps` 返回已注册微应用列表（通过 `kit-go/gateway` 集成测试覆盖）。

## 3. portal 接入 qiankun

- [x] 3.1 修改 `apps/portal/src/MicroAppPage.tsx`，使用 `qiankun` 的 `loadMicroApp` 加载子应用。
- [x] 3.2 确保 qiankun props 包含 `user/tenant/permissions/apiClient/locale`。
- [x] 3.3 在组件卸载或路由离开时调用 `microApp.unmount()`。
- [x] 3.4 验证 order-ui 在 qiankun 下仍可正常加载和交互。

## 4. portal 动态微前端路由与菜单

- [x] 4.1 在 `kit/kit-sdk-kit` 中新增 `MicroAppManifest` 类型与 `useMicroApps` hook。
- [x] 4.2 在 `apps/portal/src/microApps.ts` 中静态声明 order-ui 和 test-ui 的 manifest。
- [x] 4.3 修改 `apps/portal/src/App.tsx`，基于 manifest 动态渲染业务路由。
- [x] 4.4 修改 `apps/portal/src/Layout.tsx`，基于 manifest 动态渲染业务菜单。
- [x] 4.5 支持 `requirePermission` 过滤菜单项。

## 5. kit-cli 修复与 TS SDK 生成

- [x] 5.1 修复 `kit/kit-cli/internal/scaffold/service.go` 中 `mainStub` 的编译错误（`fmt.Println` 与 `srv.Run(nil)`）。
- [x] 5.2 为脚手架生成的服务添加 `sqlc.yaml`。
- [x] 5.3 修改 `kit/kit-cli/cmd/generate.go`，调用 `buf generate` 并为每个服务生成前端 TS SDK。
- [x] 5.4 更新 `buf.gen.yaml`，配置 TypeScript 生成插件。
- [x] 5.5 验证 `kit new service demo` 生成后可立即 `go build ./cmd`。

## 6. 平台 admin 修复

- [x] 6.1 修改 `apps/portal/src/Layout.tsx`，admin 菜单同时检查 `admin` 角色和 `platform:admin` 权限。
- [x] 6.2 修改 `kit/kit-go/server/server.go` 拦截器，发布审计事件到 `event.Bus`。
- [x] 6.3 修改 `platform/audit-service`，订阅审计事件并持久化。
- [x] 6.4 重构 `platform/tenant-service` 为 DDD 分层结构。
- [x] 6.5 重构 `platform/iam-service` 为 DDD 分层结构。
- [x] 6.6 重构 `platform/gateway-service` 为 DDD 分层结构。

## 7. Demo 测试服务

- [x] 7.1 创建 `services/test-service/api/test.proto`。
- [x] 7.2 生成 `test.pb.go`、`test_grpc.pb.go`、`test.pb.gw.go`。
- [x] 7.3 实现 `internal/domain/test.go`、`internal/app/service.go`、`internal/interfaces/grpc/handler.go`。
- [x] 7.4 编写 `services/test-service/cmd/main.go`，只写业务逻辑，通过 kit 自动注册。
- [x] 7.5 创建 `services/test-service/web/test-ui`，实现 qiankun 子应用与 `TestPage`。
- [x] 7.6 生成 `test-sdk-api` 并验证 `test-ui` 使用生成的 `TestServiceClient`。
- [x] 7.7 添加 `services/test-service/Dockerfile`。

## 8. 部署集成

- [x] 8.1 修改 `deployments/docker-compose/docker-compose.yml`，新增 `test-service`。
- [x] 8.2 修改 `apps/portal/Dockerfile`，复制 `test-ui` 构建产物到 `dist/apps/test-ui/`。
- [x] 8.3 修改 `deployments/docker-compose/nginx.conf`，添加 `/apps/test-ui/` 代理。
- [x] 8.4 更新 `Makefile` 的 `generate`、`clean` 目标覆盖平台服务与 test-service。

## 9. 验证

- [x] 9.1 运行 `pnpm -r run typecheck` 与 `pnpm -r run build`。
- [x] 9.2 运行 `make build-go` 与 `make test-go`。
- [ ] 9.3 本地 `apps/portal` dev 模式验证 qiankun 加载 test-ui。
- [ ] 9.4 Docker Compose 全量启动验证 `http://localhost/` 菜单出现「测试服务」并可调用 `/api/test/v1/echo`。
- [ ] 9.5 验证 admin 菜单对 `platform:admin` 权限用户可见。
- [ ] 9.6 验证 `audit-service` 能查询到 kit 拦截器产生的审计日志。
