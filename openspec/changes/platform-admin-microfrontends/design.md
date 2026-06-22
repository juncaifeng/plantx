## Context

The PlantX Kit platform has a working portal (`apps/portal`) and one business micro-app (`order-ui`). The platform services (`tenant-service`, `iam-service`, `audit-service`, `notification-service`) are empty Go modules. Operators currently have no UI to register services or manage routes, tenants, users, or audit logs.

## Goals / Non-Goals

**Goals:**

- Provide a minimal but runnable admin center split into per-service micro-frontends.
- Implement backend CRUD for tenants, users/roles, gateway service/route registry, and audit log querying.
- Keep each admin micro-app self-contained and loadable by the existing portal via Qiankun.
- Reuse the existing auth/tenant/authorization stack.

**Non-Goals:**

- Full-featured IAM (no OAuth2 provider, no session management, no password reset).
- Real-time route synchronization to nginx (routes will be stored and exposed via API; nginx config is still static in this change).
- Multi-cluster or Kubernetes-specific service discovery.
- Production-grade UI polish.

## Decisions

### 1. One admin micro-app per platform service

**Choice:** Each platform service gets its own `{service}-admin-ui` package under `platform/{service}/web/`.

**Rationale:**
- Matches the chosen option B and the existing `services/order/web/order-ui` pattern.
- Allows independent iteration and ownership per platform capability.
- The portal shell remains thin and only registers entries.

### 2. Use in-memory repositories for the first milestone

**Choice:** All new platform services use in-memory repositories with optional Postgres fallback via the existing `kit-go/db` abstraction.

**Rationale:**
- Faster to implement and verify without writing migrations.
- Postgres support can be added later by replacing the repository implementation.

### 3. Reuse grpc-gateway for admin REST APIs

**Choice:** Each platform service exposes gRPC and uses grpc-gateway for REST, following the same pattern as `order-service`.

**Rationale:**
- Consistent with the contract-first DDD approach.
- Admin UIs reuse the shared `@plantx/kit-sdk-api` client.

### 4. Portal loads admin micro-apps lazily

**Choice:** `apps/portal` registers admin micro-apps under `/admin/tenants`, `/admin/iam`, `/admin/gateway`, `/admin/audit` and loads them on demand.

**Rationale:**
- Keeps initial portal bundle small.
- Follows Qiankun best practice.

### 5. Service registry is the source of truth for the admin view, not nginx

**Choice:** `gateway-service` stores service definitions and route mappings; the admin UI reads from it. nginx config is manually aligned in this change.

**Rationale:**
- Avoids the complexity of hot-reloading nginx in the first iteration.
- Provides a clear API surface for future dynamic gateway implementations.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Many new packages increase build time and Docker image size. | Use a single admin Dockerfile that builds all admin UIs into one nginx image, or build per service only when needed. |
| Platform services share patterns that could duplicate code. | Extract common admin table/list components into `kit/admin-ui` later if needed; for now keep each app simple. |
| nginx config and gateway-service data can drift. | Document that routes must be added to both until dynamic gateway is implemented. |

## Migration Plan

1. Implement backend platform services.
2. Implement admin micro-frontends.
3. Register admin routes in `apps/portal`.
4. Update Docker Compose and nginx config.
5. Add smoke tests for service registration and tenant listing.

## Open Questions

- Should platform services use separate Postgres schemas or one shared PlantX database?
- Should admin UIs share a common layout component, or each use Antd Layout independently?
