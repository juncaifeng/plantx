FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy workspace and the two modules we need.
COPY go.work go.work.sum ./
COPY kit/kit-go/go.mod kit/kit-go/go.sum ./kit/kit-go/
COPY services/order/go.mod services/order/go.sum ./services/order/

# Download dependencies before copying full source for layer caching.
RUN cd services/order && go mod download

COPY kit/kit-go ./kit/kit-go
COPY services/order ./services/order

RUN cd services/order && go build -o /bin/order-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/order-service .
EXPOSE 8080
CMD ["./order-service"]
