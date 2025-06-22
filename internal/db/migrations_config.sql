-- Create user configurations table
CREATE TABLE IF NOT EXISTS user_configurations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    deepseek_token VARCHAR(1024),
    whatsapp_token VARCHAR(1024),
    whatsapp_number VARCHAR(50) NOT NULL,
    basic_prompt TEXT NOT NULL,
    max_chat_reply_count INTEGER NOT NULL,
    max_chat_reply_chars INTEGER NOT NULL,
    deepseek_token_expires TIMESTAMP,
    whatsapp_token_expires TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT NOT NULL REFERENCES users(id)
);

-- Create product knowledge data table
CREATE TABLE IF NOT EXISTS product_knowledge_data (
    id BIGSERIAL PRIMARY KEY,
    configuration_id BIGINT NOT NULL REFERENCES user_configurations(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT NOT NULL REFERENCES users(id)
);

-- Create configuration history table
CREATE TABLE IF NOT EXISTS configuration_history (
    id BIGSERIAL PRIMARY KEY,
    configuration_id BIGINT NOT NULL REFERENCES user_configurations(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL,
    changed_fields JSONB NOT NULL,
    changed_at TIMESTAMP NOT NULL,
    changed_by BIGINT NOT NULL REFERENCES users(id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_configurations_user_id ON user_configurations(user_id);
CREATE INDEX IF NOT EXISTS idx_product_knowledge_configuration_id ON product_knowledge_data(configuration_id);
CREATE INDEX IF NOT EXISTS idx_configuration_history_configuration_id ON configuration_history(configuration_id);
CREATE INDEX IF NOT EXISTS idx_configuration_history_changed_at ON configuration_history(changed_at);