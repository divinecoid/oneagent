-- Add seller_id to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS seller_id BIGINT REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_products_seller_id ON products(seller_id); 