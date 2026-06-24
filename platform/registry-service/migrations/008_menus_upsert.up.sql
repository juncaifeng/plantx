-- Deduplicate menus by application, label and route before adding the unique constraint.
DELETE FROM menus a
USING menus b
WHERE a.id < b.id
  AND a.application_id IS NOT NULL
  AND a.application_id = b.application_id
  AND a.label_key = b.label_key
  AND a.route = b.route;

-- Enforce uniqueness so repeated service registrations can upsert menus.
ALTER TABLE menus
    ADD CONSTRAINT idx_menus_application_label_route UNIQUE (application_id, label_key, route);
