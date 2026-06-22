#!/usr/bin/env bash
# End-to-end smoke test for PlantX Kit platform.
# Usage: ./scripts/e2e-smoke-test.sh [gateway_url]
# Prerequisites: Docker Compose stack is running.

set -euo pipefail

GATEWAY_URL="${1:-http://localhost}"
NETWORK="${NETWORK:-docker-compose_default}"
ORDER_SVC="order-service:8080"

GRPCURL="docker run --rm --network ${NETWORK} fullstorydev/grpcurl:latest -plaintext"

# Parse a string value from a JSON object (uses sed; no jq required).
json_str() {
  local key="$1"
  sed -n 's/.*"'"${key}"'": *"\([^"]*\)".*/\1/p'
}

echo "==> PlantX E2E Smoke Test against ${GATEWAY_URL}"

# 1. Health check
echo "--> 1. Gateway health"
curl -fsS "${GATEWAY_URL}/health" >/dev/null || { echo "gateway health failed"; exit 1; }
curl -fsS "${GATEWAY_URL}/ready" >/dev/null || { echo "gateway readiness failed"; exit 1; }

# 2. Portal static assets are served
echo "--> 2. Portal home and micro-app entry are reachable"
curl -fsS -o /dev/null "${GATEWAY_URL}/" || { echo "portal root failed"; exit 1; }
curl -fsS -o /dev/null "${GATEWAY_URL}/apps/order-ui/" || { echo "order-ui entry failed"; exit 1; }

# 3. Unauthorized gRPC request is rejected
echo "--> 3. gRPC without token is rejected"
MSG=$($GRPCURL "${ORDER_SVC}" plantx.order.v1.OrderService/ListOrders 2>&1 || true)
if ! echo "${MSG}" | grep -q "Unauthenticated"; then
  echo "expected Unauthenticated, got: ${MSG}"
  exit 1
fi

# 4. Unauthorized REST request is rejected
echo "--> 4. REST without token is rejected"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${GATEWAY_URL}/api/order/v1/orders" || true)
if [ "${HTTP_CODE}" != "401" ]; then
  echo "expected 401, got ${HTTP_CODE}"
  exit 1
fi

# 5. Obtain token for tenant A via the gateway (mock-auth)
echo "--> 5. Obtain token for tenant A"
TOKEN_A=$(curl -fsS -X POST "${GATEWAY_URL}/oauth/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=plantx-portal&username=demo-a&password=demo-a' | json_str access_token)
if [ -z "${TOKEN_A}" ] || [ "${TOKEN_A}" == "null" ]; then
  echo "failed to obtain token for tenant A"
  exit 1
fi

# 6. Create order as tenant A via REST
echo "--> 6. Create order as tenant A via REST"
ORDER_JSON=$(curl -fsS -X POST "${GATEWAY_URL}/api/order/v1/orders" \
  -H "Authorization: Bearer ${TOKEN_A}" \
  -H 'Content-Type: application/json' \
  -d '{"customer_name":"Alice"}')
echo "created order: ${ORDER_JSON}"
ORDER_ID=$(echo "${ORDER_JSON}" | json_str id)
if [ -z "${ORDER_ID}" ] || [ "${ORDER_ID}" == "null" ]; then
  echo "failed to create order"
  exit 1
fi

# 7. List orders as tenant A via REST
echo "--> 7. List orders as tenant A via REST"
LIST_A=$(curl -fsS "${GATEWAY_URL}/api/order/v1/orders" -H "Authorization: Bearer ${TOKEN_A}")
echo "${LIST_A}"
if ! echo "${LIST_A}" | grep -q '"id"'; then
  echo "tenant A should see at least one order"
  exit 1
fi

# 8. Get order as tenant A via REST
echo "--> 8. Get order as tenant A via REST"
GET_A=$(curl -fsS "${GATEWAY_URL}/api/order/v1/orders/${ORDER_ID}" -H "Authorization: Bearer ${TOKEN_A}")
echo "${GET_A}"
if ! echo "${GET_A}" | grep -q '"id"'; then
  echo "tenant A should retrieve the order"
  exit 1
fi

# 9. Obtain token for tenant B
echo "--> 9. Obtain token for tenant B"
TOKEN_B=$(curl -fsS -X POST "${GATEWAY_URL}/oauth/token" \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=password&client_id=plantx-portal&username=demo-b&password=demo-b' | json_str access_token)
if [ -z "${TOKEN_B}" ] || [ "${TOKEN_B}" == "null" ]; then
  echo "failed to obtain token for tenant B"
  exit 1
fi

# 10. Tenant B cannot access tenant A order via REST (expect 404)
echo "--> 10. Tenant B cannot access tenant A order via REST (expect 404)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${GATEWAY_URL}/api/order/v1/orders/${ORDER_ID}" -H "Authorization: Bearer ${TOKEN_B}" || true)
if [ "${HTTP_CODE}" != "404" ]; then
  echo "expected 404, got ${HTTP_CODE}"
  exit 1
fi

# 11. Tenant B list is empty via REST
echo "--> 11. List orders as tenant B via REST (expect empty)"
LIST_B=$(curl -fsS "${GATEWAY_URL}/api/order/v1/orders" -H "Authorization: Bearer ${TOKEN_B}")
echo "${LIST_B}"
if echo "${LIST_B}" | grep -q '"id"'; then
  echo "tenant B should not see tenant A orders"
  exit 1
fi

# 12. gRPC still works
echo "--> 12. Create order as tenant A via gRPC"
ORDER_JSON_GRPC=$($GRPCURL -H "authorization: Bearer ${TOKEN_A}" -d '{"customer_name":"Bob"}' "${ORDER_SVC}" plantx.order.v1.OrderService/CreateOrder)
echo "created order via gRPC: ${ORDER_JSON_GRPC}"
if ! echo "${ORDER_JSON_GRPC}" | grep -q '"id"'; then
  echo "gRPC create order failed"
  exit 1
fi

echo "==> E2E smoke test passed"
