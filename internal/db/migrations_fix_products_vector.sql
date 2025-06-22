-- 1. Change embedding column to vector(1536)
ALTER TABLE products
    ALTER COLUMN embedding TYPE vector(1536);

-- 2. Add updated_at column if missing
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='products' AND column_name='updated_at'
    ) THEN
        ALTER TABLE products ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT NOW();
    END IF;
END$$;

-- 3. Drop old trigger and function if they exist
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP FUNCTION IF EXISTS update_updated_at_column;

-- 4. Create the function and trigger unconditionally
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 