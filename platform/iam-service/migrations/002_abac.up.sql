CREATE TABLE attributes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         TEXT NOT NULL UNIQUE,
    value_type  TEXT NOT NULL DEFAULT 'string',
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE conditions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL UNIQUE,
    attribute_key TEXT NOT NULL,
    operator      TEXT NOT NULL,
    value         TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    effect      TEXT NOT NULL DEFAULT 'allow',
    priority    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE policy_permissions (
    policy_id    UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    permission   TEXT NOT NULL,
    PRIMARY KEY (policy_id, permission)
);

CREATE TABLE policy_conditions (
    policy_id    UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    condition_id UUID NOT NULL REFERENCES conditions(id) ON DELETE CASCADE,
    PRIMARY KEY (policy_id, condition_id)
);

CREATE INDEX idx_attributes_key ON attributes(key);
CREATE INDEX idx_conditions_attribute_key ON conditions(attribute_key);
CREATE INDEX idx_policy_permissions_policy_id ON policy_permissions(policy_id);
CREATE INDEX idx_policy_conditions_policy_id ON policy_conditions(policy_id);
