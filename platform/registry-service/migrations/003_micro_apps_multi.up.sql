ALTER TABLE micro_apps DROP CONSTRAINT micro_apps_pkey;

ALTER TABLE micro_apps ADD COLUMN id UUID NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE micro_apps ADD PRIMARY KEY (id);

ALTER TABLE micro_apps ADD CONSTRAINT micro_apps_service_id_name_key UNIQUE (service_id, name);

CREATE INDEX idx_micro_apps_service_id ON micro_apps(service_id);
