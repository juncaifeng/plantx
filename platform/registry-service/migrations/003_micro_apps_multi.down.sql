DROP INDEX IF EXISTS idx_micro_apps_service_id;

ALTER TABLE micro_apps DROP CONSTRAINT IF EXISTS micro_apps_service_id_name_key;

ALTER TABLE micro_apps DROP CONSTRAINT IF EXISTS micro_apps_pkey;

ALTER TABLE micro_apps DROP COLUMN IF EXISTS id;

ALTER TABLE micro_apps ADD PRIMARY KEY (service_id);
