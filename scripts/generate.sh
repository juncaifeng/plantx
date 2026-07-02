#!/usr/bin/env bash
set -euo pipefail

echo "==> Generating Kit proto code"
protoc \
  --proto_path=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  plantx/kit/authz.proto \
  plantx/kit/context.proto \
  plantx/kit/event.proto

# Move generated files from plantx/kit/ to kit/kit-go/proto/*/
mkdir -p kit/kit-go/proto/authz kit/kit-go/proto/context kit/kit-go/proto/event
mv plantx/kit/authz.pb.go kit/kit-go/proto/authz/ 2>/dev/null || true
mv plantx/kit/context.pb.go kit/kit-go/proto/context/ 2>/dev/null || true
mv plantx/kit/event.pb.go kit/kit-go/proto/event/ 2>/dev/null || true
rmdir -p plantx/kit 2>/dev/null || true

echo "==> Ensuring sqlc output directories exist"
mkdir -p \
  services/order/internal/infra/sqlc \
  platform/registry-service/internal/infra/sqlc \
  platform/iam-service/internal/infra/sqlc

echo "==> Generating sqlc code"
sqlc generate

echo "==> Generating Go/gRPC code from service proto"
npx buf generate --template buf.go.gen.yaml

# buf.go.gen.yaml outputs source-relative files at the repo root because the
# module roots are proto/, platform/, and services/. Move them to the correct
# service directories before cleaning up the stray root directories.
echo "==> Moving generated Go files to service directories"
for svc in audit-service gateway-service iam-service registry-service tenant-service; do
  if [ -d "$svc/api" ]; then
    mkdir -p "platform/$svc/api"
    mv "$svc/api"/*.pb.go "$svc/api"/*.pb.gw.go "platform/$svc/api/" 2>/dev/null || true
  fi
done
if [ -d "order/api" ]; then
  mkdir -p services/order/api
  mv order/api/*.pb.go order/api/*.pb.gw.go services/order/api/ 2>/dev/null || true
fi
if [ -d "test-service/api" ]; then
  mkdir -p services/test-service/api
  mv test-service/api/*.pb.go test-service/api/*.pb.gw.go services/test-service/api/ 2>/dev/null || true
fi

echo "==> Generating demo app proto code"
protoc \
  --proto_path=. \
  --proto_path=proto \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=. \
  --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  demo_app/backend/api/demo.proto

echo "==> Generating OpenAPI specs"
npx buf generate --template buf.openapi.gen.yaml

# Normalize source-relative OpenAPI outputs to openapi/<service-short>.yaml
echo "==> Normalizing OpenAPI spec paths"
mkdir -p openapi
rm -f openapi/*.yaml

mv openapi/audit-service/api/audit.openapi.yaml openapi/audit.yaml 2>/dev/null || true
mv openapi/gateway-service/api/gateway.openapi.yaml openapi/gateway.yaml 2>/dev/null || true
mv openapi/iam-service/api/iam.openapi.yaml openapi/iam.yaml 2>/dev/null || true
mv openapi/registry-service/api/registry.openapi.yaml openapi/registry.yaml 2>/dev/null || true
mv openapi/tenant-service/api/tenant.openapi.yaml openapi/tenant.yaml 2>/dev/null || true
mv openapi/order/api/order.openapi.yaml openapi/order.yaml 2>/dev/null || true
mv openapi/test-service/api/test.openapi.yaml openapi/test.yaml 2>/dev/null || true

# Remove leftover generated files (shared proto specs and merged output)
rm -f openapi/openapi.yaml
find openapi -mindepth 2 -type f -delete
find openapi -mindepth 1 -type d -empty -delete

# Clean stray source-relative output at repo root (buf may create top-level dirs)
for d in audit-service gateway-service iam-service registry-service tenant-service order test-service google plantx; do
  rm -rf "$d"
done

echo "==> Generate complete"
