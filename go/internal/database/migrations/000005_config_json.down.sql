-- Irreversible migration: config_json consolidation cannot be safely reversed.
-- This is a no-op down migration; restore from backup to reverse.
SELECT 1;
