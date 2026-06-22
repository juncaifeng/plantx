## 1. Backend REST gateway

- [x] 1.1 Add `google.api.http` annotations to `services/order/api/order.proto` for `CreateOrder`, `GetOrder`, and `ListOrders`.
- [x] 1.2 Add `github.com/grpc-ecosystem/grpc-gateway/v2` and `google.golang.org/genproto/googleapis/api` dependencies to `services/order/go.mod` and `kit/kit-go/go.mod` at Go 1.24-compatible versions.
- [x] 1.3 Add `protoc-gen-grpc-gateway` generation step to the order service code-generation script or Makefile.
- [x] 1.4 Generate `order.pb.gw.go` and commit it under `services/order/api/`.
- [x] 1.5 Extend `kit/kit-go/server.Options` with `GatewayRegister func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error` and mount the resulting gateway mux under `/api/` on the HTTP server.
- [x] 1.6 In `services/order/cmd/main.go`, call the new `GatewayRegister` hook with the generated `RegisterOrderServiceHandlerFromEndpoint` function.
- [x] 1.7 Update `deployments/docker-compose/nginx.conf` to route `/api/order/` to `order-service:8081/api/order/`.
- [x] 1.8 Rebuild and verify REST endpoints with curl: 401 without token, 200 with token, 404 cross-tenant.

## 2. Frontend console portal

- [x] 2.1 Create `apps/console` package with `package.json`, `vite.config.ts`, `tsconfig.json`, and React + React Router + Qiankun dependencies.
- [x] 2.2 Implement `apps/console/src/main.tsx` that bootstraps the portal, configures Qiankun micro-apps, and renders the shell layout.
- [x] 2.3 Implement `apps/console/src/pages/LoginPage.tsx` that posts credentials to `/oauth/token` and stores the access token.
- [x] 2.4 Implement `apps/console/src/pages/HomePage.tsx` and a navigation component with logout and an "Orders" link.
- [x] 2.5 Implement `@plantx/kit-sdk-api` authenticated fetch/axios client that reads the token from context and attaches `Authorization: Bearer`.
- [x] 2.6 Update `services/order/web/order-ui/src/index.ts` to consume Qiankun props and pass them into `KitProvider`.
- [x] 2.7 Update `services/order/web/order-ui/src/OrderPage.tsx` to use the shared API client from context instead of creating its own.
- [x] 2.8 Update `services/order/web/order-ui/package.json` scripts to build a micro-app bundle with a deterministic entry file name.
- [x] 2.9 Add a Vite build for `apps/console` that copies/bundles the built `order-ui` dist under `/apps/order-ui/`.

## 3. Deployment and smoke tests

- [x] 3.1 Add a `portal` service to `deployments/docker-compose/docker-compose.yml` using a Dockerfile in `apps/console` that serves the static dist.
- [x] 3.2 Create `apps/console/Dockerfile` that installs pnpm dependencies, builds the portal and order-ui, and serves them with `nginx:alpine`.
- [x] 3.3 Update `deployments/docker-compose/nginx.conf` to proxy `/` and `/apps/order-ui/` to the `portal` service.
- [x] 3.4 Update `scripts/e2e-smoke-test.sh` to include a browser/REST login flow: login via `/oauth/token`, then call `/api/order/v1/orders` and `/api/order/v1/orders` POST.
- [x] 3.5 Run `docker-compose up -d --build` and verify the full flow: portal home → login → Orders page loads → create order → list orders → tenant isolation.
- [x] 3.6 Update `AGENTS.md` or README with instructions for running the console and adding new micro-frontends.
