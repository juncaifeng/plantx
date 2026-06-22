## Context

The backend milestone is verified: `order-service` starts via Docker Compose, authenticates with `mock-auth`, authorizes with OPA, and exposes gRPC with tenant isolation. The frontend side only has unconnected pieces:

- `kit/kit-ui`: shared layout primitives.
- `services/order/web/order-ui`: a Qiankun micro-app skeleton with no runtime entry.
- `services/order/web/order-sdk-*`: SDKs that assume a REST API which does not exist yet.
- No portal shell to load the micro-app.

This design connects those pieces into a runnable browser console.

## Goals / Non-Goals

**Goals:**

- Provide a browser console (`apps/console`) that can log in and load `order-ui` as a Qiankun micro-app.
- Expose `order-service` gRPC methods as REST JSON endpoints through the nginx gateway.
- Keep the existing auth flow intact (gateway → mock-auth JWT → order-service gRPC interceptors).
- Add a Docker Compose service for the portal and update the smoke test to cover the browser path.

**Non-Goals:**

- Production-grade UI/UX polish (console is a functional scaffold).
- Multiple independent micro-frontend deployments in this change (only `order-ui` is wired).
- Replacing gRPC with REST internally; REST is a gateway layer only.
- SSR or advanced code-splitting.

## Decisions

### 1. REST gateway: grpc-gateway v2 generated from proto annotations

**Choice:** Add `google.api.http` annotations to `services/order/api/order.proto` and generate a grpc-gateway handler.

**Rationale:**
- Keeps the contract-first workflow: REST paths and request/response shapes are derived from the same proto used for gRPC and SDK generation.
- The generated gateway handler translates HTTP/JSON to gRPC calls locally, reusing existing interceptors for auth, tenant, and authorization.
- It is the standard Go ecosystem tool for this pattern.

**Alternatives considered:**
- Hand-written HTTP adapter in `order-service`: faster to write but diverges from contract-first principles.
- Direct gRPC-web: would require an envoy/nginx gRPC-web filter and a different client stack; grpc-gateway is simpler for the current React SDK.

### 2. Kit server integration: optional `GatewayRegister` hook

**Choice:** Extend `kit/kit-go/server.Options` with an optional `GatewayRegister func(ctx, mux, conn) error`. The server creates a `*runtime.ServeMux`, dials its own gRPC port on localhost, runs the register function, and mounts the resulting mux under `/api/` on the HTTP port.

**Rationale:**
- The portal and other services can register gateways without changing kit internals for each service.
- `/api/` is a stable convention already used by nginx.
- Mounting under the kit HTTP server keeps observability and readiness unified.

### 3. Portal architecture: React + Vite + Qiankun + React Router

**Choice:** Create `apps/console` as a pnpm workspace package. The portal builds into a static site, served by a dedicated `portal` container behind the gateway.

**Rationale:**
- Vite is already the expected frontend build tool for the project.
- Qiankun is listed in `order-ui` dependencies and matches the stated micro-frontend strategy.
- React Router handles top-level navigation between the portal shell and loaded micro-apps.

### 4. Micro-app serving: portal image bundles order-ui

**Choice:** The `portal` Docker image copies both `apps/console/dist` and `services/order/web/order-ui/dist`, serving them from the same origin under `/apps/order-ui/`.

**Rationale:**
- Avoids adding another compose service just for the first micro-app.
- Qiankun can load the micro-app entry JS from the same origin, avoiding CORS.
- Future services can add their own static mounts or independent services without changing the shell design.

### 5. Auth context propagation

**Choice:**
- Portal login form posts `grant_type=password` to `/oauth/token` through the gateway.
- Access token is stored in memory (and optionally `localStorage` for page reloads).
- `@plantx/kit-sdk-api` attaches `Authorization: Bearer <token>` to every request.
- Portal passes decoded token claims (`user`, `tenant`, `permissions`) into Qiankun `mount` props, so `order-ui` can use `useKitContext`.

**Rationale:**
- Reuses the existing mock-auth/OAuth2 password grant used by the gRPC smoke test.
- Passing context through props keeps micro-apps decoupled from token management.

### 6. Nginx routing

**Choice:** Update `deployments/docker-compose/nginx.conf`:
- `/api/order/` → `order-service:8081/api/order/`
- `/oauth/token` → `mock-auth:8080/oauth/token`
- `/auth/` → `mock-auth:8080/`
- `/apps/order-ui/` → `portal:80/apps/order-ui/`
- `/` → `portal:80/`

**Rationale:**
- All browser traffic goes through one gateway port, matching the existing stack.
- REST prefix per service (`/api/order/`) scopes future service additions.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| grpc-gateway version may require Go >1.24 if latest is used. | Pin `github.com/grpc-ecosystem/grpc-gateway/v2` to a version compatible with the project's Go 1.24 toolchain (v2.26.x). |
| Qiankun + Vite dev/build can have module loading issues. | Build both portal and micro-app as ES modules with deterministic entry file names; load via `<script type="module" src="...">` in Qiankun. |
| Portal and gateway are separate nginx instances; micro-app static paths must match build base. | Set Vite `base` correctly (`/` for portal, `/apps/order-ui/` for order-ui) and use relative Qiankun entry URLs. |
| Decoding JWT in the browser without signature verification leaks trust to the gateway. | The portal only uses token claims for UI display; the gateway and order-service re-validate the token signature and expiry. |
| Local smoke test becomes heavier (requires browser or static fetch). | Provide a curl-based login + REST API smoke path and optionally a Playwright test. |

## Migration Plan

1. Merge the change and regenerate `order` proto gateway code.
2. Update Docker images for `order-service` and build the new `portal` image.
3. Run `docker-compose up -d --build` and verify `/health`, `/ready`, `/oauth/token`, `/api/order/v1/orders`.
4. Open `http://localhost` in a browser, log in as `demo-a`, and create/list orders.
5. Run the updated smoke test.

## Open Questions

- Should the portal decode the JWT locally or call a future `/userinfo` endpoint? (For this change: local decode is sufficient.)
- Should each micro-app be built into its own container in the future? (Yes, but out of scope for this change.)
