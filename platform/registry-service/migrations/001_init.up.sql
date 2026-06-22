CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE registry_services (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    grpc_host   TEXT NOT NULL,
    rest_prefix TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE micro_apps (
    service_id         UUID PRIMARY KEY REFERENCES registry_services(id) ON DELETE CASCADE,
    name               TEXT NOT NULL,
    route              TEXT NOT NULL,
    bundle_url         TEXT NOT NULL,
    menu_label_key     TEXT NOT NULL,
    require_permission TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_registry_services_name ON registry_services(name);
