---
name: development-workflow
description: >
  Provides the PlantX development workflow conventions including Conventional Commits,
  code generation with buf/sqlc, CI/CD pipeline behavior, SDK release via git tags,
  and local development commands. Use when a developer or agent needs to understand
  how to contribute code, run generators, or release packages in the PlantX monorepo.
metadata:
  author: PlantX Platform Team
  version: "1.2"
  updated: "2026-06-24"
---

# PlantX Development Workflow

This skill documents the standardized development process for the PlantX monorepo.

## 1. Source Control

### Conventional Commits

All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```text
<type>(<scope>): <short description>

<body>

<footer>
```

Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `build`.

Examples:

```bash
git commit -m "feat(order): add order cancellation endpoint"
git commit -m "fix(gateway): correct route prefix matching"
git commit -m "docs: update development workflow"
```

### Commitlint

Commit messages in the branch/PR range are validated by `.github/workflows/commitlint.yml`.

## 2. Code Generation

### Canonical Generation Command

After modifying any proto file or SQLC query, run the canonical generator:

```bash
bash scripts/generate.sh
```

This script runs `buf generate` with the Go/gRPC and OpenAPI templates, plus `sqlc generate`.

Proto files live in multiple locations:

- `proto/plantx/kit/` — shared kit types (authz, context, event)
- `platform/<service>/api/<service>.proto` — platform service definitions
- `services/<service>/api/<service>.proto` — business service definitions

### Protobuf

For local, single-service Go/gRPC generation you can also use Makefile targets, but `scripts/generate.sh` is the canonical command:

```bash
# Canonical
bash scripts/generate.sh

# Makefile targets for specific services (do not run sqlc/OpenAPI)
make generate-order
make generate-test
```

Generated outputs:

- Go gRPC/gateway code: `platform/<service>/api/`, `services/<service>/api/`
- OpenAPI specs: `openapi/<service-short>.yaml`
- TypeScript SDK in `kit/kit-sdk-api/src/` is currently hand-maintained. The `ts` plugin is configured in `buf.gen.yaml` but not yet wired into `scripts/generate.sh`.

### SQLC

After modifying migrations or queries under a service `internal/infra/sqlc/` directory:

```bash
sqlc generate
```

Query directories are configured in `sqlc.yaml`.

### Rule

Generated code must not be hand-edited. The `Generate Check` workflow verifies consistency between source definitions and generated artifacts.

## 3. Continuous Integration

The `CI` workflow (`.github/workflows/ci.yml`) runs:

```
generate
  → lint-go ─┐
  → test-go ─┼→ build-images → publish-images
  → lint-web─┘
```

`lint-go`, `test-go`, and `lint-web` run in parallel after `generate`.

| Job | Purpose |
|-----|---------|
| `generate` | Runs `scripts/generate.sh` and stores artifacts |
| `lint-go` | Runs golangci-lint inside every main Go module |
| `test-go` | Runs Go unit tests across kit and service modules |
| `lint-web` | Lints TypeScript packages |
| `build-images` | Builds Docker images for services |
| `publish-images` | Pushes images to the registry |

### Go Lint Notes

Because the repository uses `go.work`, lint must run inside each main module:

```bash
for dir in $(go list -m -f '{{if .Main}}{{.Dir}}{{end}}' all); do
  (cd "$dir" && golangci-lint run ./...)
done
```

### Docker Build Notes

The `CI` workflow uses the repository root as the build context and disables the Go workspace:

```dockerfile
COPY . .
RUN cd services/order && GOWORK=off go build -o /bin/order-service ./cmd
```

A separate `.github/workflows/release.yml` exists for tag-based releases and currently uses `context: ./services/order`. Be aware of this difference when debugging release builds.

## 4. SDK Release

SDK packages (`kit/*`) are released by pushing a git tag. The tag version becomes the version for all published kit packages.

```bash
# Tag format must be v<semver>, for example:
git tag v0.2.0
git push origin v0.2.0
```

The `Release SDK` workflow (`.github/workflows/release-sdk.yml`) then:

1. Builds all `kit/*` packages.
2. Bumps every published `kit/*` package version to the tag version.
3. Publishes TypeScript SDK packages to npmjs.org.
4. Creates Go module tags for `kit/kit-go`, `kit/kit-go/gateway`, and `kit/kit-cli`.
5. Commits the `package.json` version bumps back to `main`.

### Versioning Rules

- Use a single monorepo tag (`v0.2.0`) for all kit packages.
- Tag `v0.2.0` will:
  - Set npm packages `@plantx/kit-sdk-api`, `@plantx/kit-sdk-kit`, and `@plantx/kit-ui` to `0.2.0`.
  - Create Go module tags `kit/kit-go/v0.2.0`, `kit/kit-go/gateway/v0.2.0`, and `kit/kit-cli/v0.2.0`.
- If a package should not be published, ensure it is excluded from the `ci:publish` script in `package.json`.

### Go SDK Consumption

Go modules do not use npm. After the workflow pushes the Go module tags, external consumers can fetch a released version:

```bash
go get github.com/plantx/kit/kit-go@v0.2.0
go get github.com/plantx/kit/kit-go/gateway@v0.2.0
```

Inside the PlantX monorepo, services continue to use `replace` directives in `go.mod` to develop against the local kit-go source.

### NPM Token

`NPM_TOKEN` must be a **Granular Access Token** with publish permission for the `@plantx` scope and 2FA bypass. Ordinary access tokens will fail with `403 Forbidden`.

## 5. Local Commands

```bash
# Install dependencies (CI uses --frozen-lockfile)
pnpm install

# Generate code (canonical)
bash scripts/generate.sh

# Run Go tests across all modules
go test ./kit/kit-go/... ./kit/kit-go/gateway/... ./services/order/... ./kit/kit-cli/...

# Run Go lint per module
for dir in $(go list -m -f '{{if .Main}}{{.Dir}}{{end}}' all); do
  (cd "$dir" && golangci-lint run ./...)
done

# Build SDKs
pnpm -r --filter './kit/**' run build

# Bump kit package versions locally (dry-run example)
pnpm -r --filter './kit/**' exec npm version 0.2.0 --no-git-tag-version
```

## 6. Pull Request Guidelines

1. Keep PRs small and focused.
2. Regenerate and commit generated code before opening a PR.
3. For user-visible SDK changes, push a new `v<semver>` tag after merging to release.
4. Ensure all CI checks pass before merging.
5. Avoid direct pushes to `main`; use pull requests.

## 7. Documentation Sync

Changes to CI/CD, build scripts, or release procedures must be reflected in:

- `AGENTS.md`
- `skills/development-workflow/SKILL.md` (this file)

## 8. Related Files

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/generate-check.yml`
- `.github/workflows/release-sdk.yml`
- `.github/workflows/commitlint.yml`
- `.golangci.yml`
- `buf.gen.yaml`
- `buf.go.gen.yaml`
- `buf.openapi.gen.yaml`
- `sqlc.yaml`
- `scripts/generate.sh`
- `package.json`
- `AGENTS.md`
