FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN cd demo_app/backend && GOWORK=off go mod download
RUN cd demo_app/backend && GOWORK=off go build -o /bin/demo-service ./cmd

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/demo-service .
COPY --from=builder /app/demo_app/backend/config ./config
EXPOSE 8080 8081
CMD ["./demo-service"]
