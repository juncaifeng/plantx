[Unit]
Description=PlantX Order Service
After=network.target

[Service]
Type=simple
User=plantx
Group=plantx
ExecStart=/opt/plantx/bin/order-service
Restart=always
RestartSec=5
Environment="ORDER_GRPC_PORT=8080"
Environment="ORDER_HTTP_PORT=8081"
Environment="ORDER_DATABASE_DSN=postgres://plantx:{{ .DbPassword }}@localhost:5432/plantx?sslmode=disable"
Environment="ORDER_NATS_URL=nats://localhost:4222"
Environment="ORDER_TRACING_ENABLED=true"
Environment="MAXKEY_ISSUER={{ .MaxKeyIssuer }}"
Environment="MAXKEY_JWKS_URL={{ .MaxKeyJWKS }}"
Environment="OPA_URL={{ .OPAURL }}"
Environment="OPA_DECISION_PATH=plantx/authz/allow"

[Install]
WantedBy=multi-user.target
