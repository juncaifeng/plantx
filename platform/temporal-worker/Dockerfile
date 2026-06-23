FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy module definitions for dependencies first (leveraging Docker cache)
COPY kit/kit-go/go.mod kit/kit-go/go.sum ./kit/kit-go/
COPY platform/registry-service/api/go.mod platform/registry-service/api/go.sum ./platform/registry-service/api/
COPY platform/temporal-worker/go.mod platform/temporal-worker/go.sum ./platform/temporal-worker/

# Download module dependencies
RUN cd platform/temporal-worker && go mod download

# Copy source code preserving monorepo structure
COPY kit/kit-go/ ./kit/kit-go/
COPY platform/registry-service/api/ ./platform/registry-service/api/
COPY platform/temporal-worker/ ./platform/temporal-worker/

# Build the worker binary
RUN cd platform/temporal-worker && CGO_ENABLED=0 go build -o /app/temporal-worker ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/temporal-worker .
CMD ["./temporal-worker"]
