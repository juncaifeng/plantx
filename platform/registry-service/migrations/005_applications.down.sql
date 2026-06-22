DROP INDEX IF EXISTS idx_menus_application_id;
DROP INDEX IF EXISTS idx_micro_apps_application_id;
DROP INDEX IF EXISTS idx_registry_services_application_id;

ALTER TABLE menus DROP COLUMN IF EXISTS application_id;
ALTER TABLE micro_apps DROP COLUMN IF EXISTS application_id;
ALTER TABLE registry_services DROP COLUMN IF EXISTS application_id;

DROP TABLE IF EXISTS applications;
