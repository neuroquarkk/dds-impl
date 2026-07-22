CREATE TABLE IF NOT EXISTS configs (
    key VARCHAR(255) PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- This table will always hold a single row
-- Stores MD5 hash of current configuration
CREATE TABLE IF NOT EXISTS config_metadata (
    id SERIAL PRIMARY KEY CHECK (id = 1),
    hash TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Change Detection Trigger
-- Every time a row is inserted, updated or deleted
-- it calculates the hash of entire config
CREATE OR REPLACE FUNCTION recompute_config_hash() RETURNS trigger AS $$
BEGIN
    INSERT INTO config_metadata (id, hash, updated_at)
    SELECT 1, md5(string_agg(key || '=' || value, ',' ORDER BY key)), NOW()
    FROM configs
    ON CONFLICT (id) DO UPDATE SET
        hash = EXCLUDED.hash,
        updated_at = EXCLUDED.updated_at;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER configs_hash_trigger
AFTER INSERT OR UPDATE OR DELETE ON configs
FOR EACH STATEMENT
EXECUTE FUNCTION recompute_config_hash();

INSERT INTO configs (key, value) VALUES
    ('DB_URL', 'postgres://postgres:postgres@localhost:5432/toy_db'),
    ('MAX_CONNECTIONS', '10'),
    ('CACHE_ENABLED', 'true')
ON CONFLICT (key) DO NOTHING;
