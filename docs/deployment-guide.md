# Deployment Guide

PlantX supports three deployment shapes: Docker Compose, Kubernetes (Helm), and bare metal (systemd + Nginx).

## Docker Compose

```bash
cd deployments/docker-compose
docker compose up -d
```

Services: PostgreSQL, NATS, MaxKey, OPA, Nginx gateway, order-service.

## Kubernetes Helm

```bash
cd deployments/k8s/plantx
helm dependency update
helm install plantx . --set orderService.databaseDSN=...
```

## Bare Metal / Systemd

1. Build the binary: `go build -o /opt/plantx/bin/order-service ./services/order/cmd`
2. Copy `deployments/systemd/*.service` to `/etc/systemd/system/`.
3. Copy `deployments/nginx/nginx.conf` to `/etc/plantx/nginx/`.
4. Enable and start services:

```bash
systemctl enable plantx-order.service plantx-gateway.service
systemctl start plantx-order.service plantx-gateway.service
```

## APISIX

For APISIX deployments use `deployments/nginx/apisix.yaml` as a route template.

## Rollback

- Docker Compose: `docker compose down && git checkout <tag> && docker compose up -d`
- K8s: `helm rollback plantx <revision>`
- Systemd: revert binary and `systemctl restart`.
