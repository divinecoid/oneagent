-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    email_verify_token VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expires TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);

-- Create index for session expiration
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Create index for user email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Create index for reset token
CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token);

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