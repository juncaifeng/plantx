FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN cd services/order && GOWORK=off go mod download
RUN cd services/order && GOWORK=off go build -o /bin/order-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/order-service .
EXPOSE 8080
CMD ["./order-service"]
