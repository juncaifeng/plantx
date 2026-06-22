CREATE TABLE menus (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label_key          TEXT NOT NULL,
    route              TEXT NOT NULL DEFAULT '',
    icon               TEXT NOT NULL DEFAULT '',
    parent_id          UUID REFERENCES menus(id) ON DELETE CASCADE,
    sort_order         INT NOT NULL DEFAULT 0,
    micro_app_name     TEXT NOT NULL DEFAULT '',
    require_permission TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_menus_parent_id ON menus(parent_id);
CREATE INDEX idx_menus_sort_order ON menus(sort_order);
