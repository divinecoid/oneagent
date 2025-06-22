-- Update all existing configurations to use GPT-4-turbo-preview
UPDATE user_configurations 
SET openai_model = 'gpt-4-turbo-preview'
WHERE openai_model IS NULL OR openai_model = 'gpt-3.5-turbo'; 