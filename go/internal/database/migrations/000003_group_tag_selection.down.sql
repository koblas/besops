-- SQLite 3.35+ supports DROP COLUMN
ALTER TABLE "group" DROP COLUMN tag_ids_json;

DROP INDEX IF EXISTS idx_monitor_tag_tag_id;

ALTER TABLE monitor DROP COLUMN group_tag_ids_json;
