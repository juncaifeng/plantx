CREATE TABLE route_policies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id      UUID NOT NULL UNIQUE REFERENCES registry_services(id) ON DELETE CASCADE,
    rate_limit_rps  INT NOT NULL DEFAULT 0,
    auth_required   BOOLEAN NOT NULL DEFAULT true,
    canary_weight   INT NOT NULL DEFAULT 0,
    canary_host     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_route_policies_service_id ON route_policies(service_id);
