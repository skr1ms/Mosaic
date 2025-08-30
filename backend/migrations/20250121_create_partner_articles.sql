-- Create partner_articles table for storing product SKUs
CREATE TABLE IF NOT EXISTS partner_articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES partners(id) ON DELETE CASCADE,
    size VARCHAR(10) NOT NULL,
    style VARCHAR(20) NOT NULL,
    marketplace VARCHAR(20) NOT NULL,
    sku VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_partner_articles_partner_id ON partner_articles(partner_id);
CREATE INDEX IF NOT EXISTS idx_partner_articles_size_style ON partner_articles(size, style);
CREATE INDEX IF NOT EXISTS idx_partner_articles_marketplace ON partner_articles(marketplace);
CREATE INDEX IF NOT EXISTS idx_partner_articles_sku ON partner_articles(sku);
CREATE UNIQUE INDEX IF NOT EXISTS idx_partner_articles_unique ON partner_articles(partner_id, size, style, marketplace);

-- Add trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_partner_articles_updated_at BEFORE UPDATE
    ON partner_articles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();