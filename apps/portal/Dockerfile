FROM node:20-alpine AS builder
WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.0.0 --activate

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY tsconfig.base.json tsconfig.json ./

COPY kit/kit-ui ./kit/kit-ui
COPY kit/kit-sdk-api ./kit/kit-sdk-api
COPY kit/kit-sdk-kit ./kit/kit-sdk-kit

COPY apps/portal ./apps/portal
COPY services/order/web/order-ui ./services/order/web/order-ui
COPY services/order/web/order-sdk-api ./services/order/web/order-sdk-api
COPY services/order/web/order-sdk-kit ./services/order/web/order-sdk-kit

COPY services/test-service/web/test-ui ./services/test-service/web/test-ui
COPY services/test-service/web/test-sdk-api ./services/test-service/web/test-sdk-api

COPY platform/tenant-service/web/tenant-admin-ui ./platform/tenant-service/web/tenant-admin-ui
COPY platform/iam-service/web/iam-admin-ui ./platform/iam-service/web/iam-admin-ui
COPY platform/gateway-service/web/gateway-admin-ui ./platform/gateway-service/web/gateway-admin-ui
COPY platform/audit-service/web/audit-admin-ui ./platform/audit-service/web/audit-admin-ui
COPY platform/registry-service/web/registry-admin-ui ./platform/registry-service/web/registry-admin-ui
COPY platform/registry-service/web/api-explorer-ui ./platform/registry-service/web/api-explorer-ui
COPY openapi ./openapi

RUN pnpm install --frozen-lockfile
RUN pnpm --filter @plantx/portal build
RUN pnpm --filter @plantx/order-ui build
RUN pnpm --filter @plantx/test-ui build
RUN pnpm --filter @plantx/tenant-admin-ui build
RUN pnpm --filter @plantx/iam-admin-ui build
RUN pnpm --filter @plantx/gateway-admin-ui build
RUN pnpm --filter @plantx/audit-admin-ui build
RUN pnpm --filter @plantx/registry-admin-ui build
RUN pnpm --filter @plantx/api-explorer-ui build

# Bundle micro-apps into portal dist
RUN mkdir -p apps/portal/dist/apps/order-ui && \
    cp services/order/web/order-ui/dist/order-ui.iife.js apps/portal/dist/apps/order-ui/order-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./order-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/order-ui/index.html

RUN mkdir -p apps/portal/dist/apps/test-ui && \
    cp services/test-service/web/test-ui/dist/test-ui.iife.js apps/portal/dist/apps/test-ui/test-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./test-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/test-ui/index.html

RUN mkdir -p apps/portal/dist/apps/tenant-admin-ui && \
    cp platform/tenant-service/web/tenant-admin-ui/dist/tenant-admin-ui.iife.js apps/portal/dist/apps/tenant-admin-ui/tenant-admin-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./tenant-admin-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/tenant-admin-ui/index.html

RUN mkdir -p apps/portal/dist/apps/iam-admin-ui && \
    cp platform/iam-service/web/iam-admin-ui/dist/iam-admin-ui.iife.js apps/portal/dist/apps/iam-admin-ui/iam-admin-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./iam-admin-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/iam-admin-ui/index.html

RUN mkdir -p apps/portal/dist/apps/gateway-admin-ui && \
    cp platform/gateway-service/web/gateway-admin-ui/dist/gateway-admin-ui.iife.js apps/portal/dist/apps/gateway-admin-ui/gateway-admin-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./gateway-admin-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/gateway-admin-ui/index.html

RUN mkdir -p apps/portal/dist/apps/audit-admin-ui && \
    cp platform/audit-service/web/audit-admin-ui/dist/audit-admin-ui.iife.js apps/portal/dist/apps/audit-admin-ui/audit-admin-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./audit-admin-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/audit-admin-ui/index.html

RUN mkdir -p apps/portal/dist/apps/registry-admin-ui && \
    cp platform/registry-service/web/registry-admin-ui/dist/registry-admin-ui.iife.js apps/portal/dist/apps/registry-admin-ui/registry-admin-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./registry-admin-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/registry-admin-ui/index.html

RUN mkdir -p apps/portal/dist/apps/api-explorer-ui && \
    cp platform/registry-service/web/api-explorer-ui/dist/api-explorer-ui.iife.js apps/portal/dist/apps/api-explorer-ui/api-explorer-ui.js && \
    printf '%s' '<!DOCTYPE html><html><body><script>window.process={env:{NODE_ENV:"production"}};</script><script src="./api-explorer-ui.js"></script></body></html>' \
      > apps/portal/dist/apps/api-explorer-ui/index.html

FROM nginx:alpine
COPY --from=builder /app/apps/portal/dist /usr/share/nginx/html
COPY --from=builder /app/openapi /usr/share/nginx/html/openapi
COPY apps/portal/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
