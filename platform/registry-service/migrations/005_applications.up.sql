CREATE TABLE applications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    label_key   TEXT NOT NULL,
    icon        TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'ACTIVE',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed a default application for backwards compatibility.
INSERT INTO applications (key, name, label_key, status, sort_order)
VALUES ('platform', 'Platform', 'nav.platform', 'ACTIVE', 0)
ON CONFLICT (key) DO NOTHING;

ALTER TABLE registry_services
    ADD COLUMN application_id UUID REFERENCES applications(id) ON DELETE SET NULL;

UPDATE registry_services
    SET application_id = (SELECT id FROM applications WHERE key = 'platform')
    WHERE application_id IS NULL;

ALTER TABLE micro_apps
    ADD COLUMN application_id UUID REFERENCES applications(id) ON DELETE SET NULL;

UPDATE micro_apps
    SET application_id = (SELECT id FROM applications WHERE key = 'platform')
    WHERE application_id IS NULL;

ALTER TABLE menus
    ADD COLUMN application_id UUID REFERENCES applications(id) ON DELETE SET NULL;

UPDATE menus
    SET application_id = (SELECT id FROM applications WHERE key = 'platform')
    WHERE application_id IS NULL;

CREATE INDEX idx_registry_services_application_id ON registry_services(application_id);
CREATE INDEX idx_micro_apps_application_id ON micro_apps(application_id);
CREATE INDEX idx_menus_application_id ON menus(application_id);
