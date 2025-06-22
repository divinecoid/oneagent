-- Rename columns
ALTER TABLE user_configurations RENAME COLUMN openai_token TO openai_api_key;
ALTER TABLE user_configurations RENAME COLUMN openai_token_expires TO openai_api_key_expires;

-- Create deleted_configs_log table
CREATE TABLE IF NOT EXISTS deleted_configs_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    config_id BIGINT,
    old_api_key TEXT,
    deleted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reason TEXT
);

-- Remove invalid API keys and log them
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN SELECT id, user_id, openai_api_key FROM user_configurations LOOP
        IF rec.openai_api_key IS NULL OR TRIM(rec.openai_api_key) = '' OR LEFT(rec.openai_api_key, 3) <> 'sk-' OR LENGTH(rec.openai_api_key) < 20 THEN
            INSERT INTO deleted_configs_log(user_id, config_id, old_api_key, reason)
            VALUES (rec.user_id, rec.id, rec.openai_api_key, 'Invalid or empty OpenAI API key during migration');
            DELETE FROM user_configurations WHERE id = rec.id;
        END IF;
    END LOOP;
END $$; 