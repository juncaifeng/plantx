#!/bin/sh
set -e

REGISTRY_SYNC_URL="${REGISTRY_SYNC_URL:-http://registry-service:8081/api/registry/v1/sync-routes}"
MICRO_APPS_URL="${MICRO_APPS_URL:-http://registry-service:8081/api/registry/v1/micro-apps}"
TOKEN_URL="${TOKEN_URL:-http://mock-auth:8080/oauth/token}"
CLIENT_ID="${CLIENT_ID:-portal}"
CLIENT_SECRET="${CLIENT_SECRET:-portal}"
SYNC_INTERVAL="${SYNC_INTERVAL:-30}"

fetch_token() {
  curl -s -X POST "$TOKEN_URL" \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    -d "grant_type=client_credentials&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET" \
    | jq -r '.access_token // empty'
}

wait_for_registry() {
  for i in $(seq 1 60); do
    TOKEN=$(fetch_token)
    if [ -n "$TOKEN" ]; then
      status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$REGISTRY_SYNC_URL" || true)
      if [ "$status" = "200" ]; then
        return 0
      fi
    fi
    echo "nginx-sync: waiting for registry..."
    sleep 2
  done
  echo "nginx-sync: registry not ready, generating fallback config"
  return 1
}

render_conf() {
  mkdir -p /etc/nginx/conf.d

  TOKEN=$(fetch_token)
  if [ -z "$TOKEN" ]; then
    routes_json='{"routes":[]}'
    micro_apps_json='{"microApps":[]}'
  else
    routes_json=$(curl -s -H "Authorization: Bearer $TOKEN" "$REGISTRY_SYNC_URL" || echo '{"routes":[]}')
    micro_apps_json=$(curl -s -H "Authorization: Bearer $TOKEN" "$MICRO_APPS_URL" || echo '{"microApps":[]}')
  fi

  # Validate JSON shape; fallback on malformed responses
  if ! echo "$routes_json" | jq -e '.routes' >/dev/null 2>&1; then
    routes_json='{"routes":[]}'
  fi
  if ! echo "$micro_apps_json" | jq -e '.microApps' >/dev/null 2>&1; then
    micro_apps_json='{"microApps":[]}'
  fi

  # Upstreams
  {
    echo "upstream mock_auth { server mock-auth:8080; }"
    echo "upstream portal { server portal:80; }"
    echo "$routes_json" | jq -r '.routes[]? | "upstream \(.name | gsub("-"; "_")) { server \(.upstreamHost); }"'
  } > /etc/nginx/conf.d/upstreams.conf

  # Rate limit zones
  {
    echo "$routes_json" | jq -r '.routes[]? | select((.policy.rateLimitRps // 0) > 0) | "limit_req_zone $binary_remote_addr zone=\(.name | gsub("-"; "_"))_lim:1m rate=\(.policy.rateLimitRps)r/s;"'
  } > /etc/nginx/conf.d/rate-limits.conf

  # Service locations
  {
    echo "$routes_json" | jq -r '
      .routes[]? |
      "    location \(.restPrefix) {\n" +
      "        proxy_pass http://\(.name | gsub("-"; "_"))\(.restPrefix);\n" +
      (if (.policy.rateLimitRps // 0) > 0 then "        limit_req zone=\(.name | gsub("-"; "_"))_lim burst=20 nodelay;\n" else "" end) +
      (if (.policy.authRequired // true) | not then "        # auth disabled by policy\n" else "" end) +
      "        proxy_set_header Host $host;\n" +
      "        proxy_set_header X-Real-IP $remote_addr;\n" +
      "    }"
    '
  } > /etc/nginx/conf.d/locations.conf

  # Micro-app locations
  {
    echo "$micro_apps_json" | jq -r '
      .microApps[]? |
      "    location \(.bundleUrl | sub("/[^/]+$"; "/")) {\n" +
      "        proxy_pass http://portal\(.bundleUrl | sub("/[^/]+$"; "/"));\n" +
      "        proxy_set_header Host $host;\n" +
      "    }"
    '
  } > /etc/nginx/conf.d/micro-app-locations.conf

  # Static shared locations
  cat > /etc/nginx/conf.d/static-locations.conf <<'EOF'
    location /auth/ {
        proxy_pass http://mock_auth/;
        proxy_set_header Host $host;
    }

    location /oauth/token {
        proxy_pass http://mock_auth/oauth/token;
        proxy_set_header Host $host;
    }

    location /openapi/ {
        proxy_pass http://portal/openapi/;
        proxy_set_header Host $host;
    }

    location / {
        proxy_pass http://portal/;
        proxy_set_header Host $host;
    }
EOF
}

# Wait for registry, but continue with fallback if unavailable
wait_for_registry || true

# Initial render
render_conf

# Validate and start nginx
nginx -t
nginx -g 'daemon off;' &
NGINX_PID=$!

cleanup() {
  kill "$NGINX_PID" 2>/dev/null || true
}
trap cleanup EXIT

echo "nginx-sync: started, reloading every ${SYNC_INTERVAL}s"

while true; do
  sleep "$SYNC_INTERVAL"
  TOKEN=$(fetch_token)
  if [ -z "$TOKEN" ]; then
    echo "nginx-sync: token refresh failed, skipping this cycle"
    continue
  fi
  render_conf
  if nginx -t 2>/dev/null; then
    nginx -s reload 2>/dev/null || true
    echo "nginx-sync: routes reloaded"
  else
    echo "nginx-sync: generated config invalid, skipping reload"
  fi
done
