FROM golang:1.24-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
COPY kit/kit-go/go.mod kit/kit-go/go.sum ./kit/kit-go/
COPY kit/kit-go/gateway/go.mod kit/kit-go/gateway/go.sum ./kit/kit-go/gateway/
COPY platform/registry-service/go.mod platform/registry-service/go.sum ./platform/registry-service/
COPY platform/registry-service/api/go.mod platform/registry-service/api/go.sum ./platform/registry-service/api/
COPY platform/gateway-service/go.mod platform/gateway-service/go.sum ./platform/gateway-service/
COPY services/test-service/go.mod services/test-service/go.sum ./services/test-service/
RUN cd services/test-service && go mod download
COPY kit/kit-go/ ./kit/kit-go/
COPY platform/registry-service/ ./platform/registry-service/
COPY platform/gateway-service/ ./platform/gateway-service/
COPY services/test-service/ ./services/test-service/
RUN cd services/test-service && go build -o /bin/test-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/test-service .
EXPOSE 8080
CMD ["./test-service"]
