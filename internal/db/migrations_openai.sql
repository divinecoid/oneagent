-- Rename DeepSeek columns to OpenAI
ALTER TABLE user_configurations 
    RENAME COLUMN deepseek_token TO openai_token;

ALTER TABLE user_configurations 
    RENAME COLUMN deepseek_token_expires TO openai_token_expires;

-- Add OpenAI model configuration
ALTER TABLE user_configurations
    ADD COLUMN openai_model VARCHAR(50) NOT NULL DEFAULT 'gpt-3.5-turbo';

-- Add OpenAI embedding model configuration
ALTER TABLE user_configurations
    ADD COLUMN openai_embedding_model VARCHAR(50) NOT NULL DEFAULT 'text-embedding-3-small'; 