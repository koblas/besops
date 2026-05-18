ALTER TABLE monitor ADD COLUMN group_tag_ids_json TEXT DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_monitor_tag_tag_id ON monitor_tag(tag_id);

ALTER TABLE "group" ADD COLUMN tag_ids_json TEXT DEFAULT NULL;
