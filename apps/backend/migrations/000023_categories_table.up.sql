-- Replace hardcoded test_category enum with a categories table (admin-manageable).
-- 1. Create the categories table
CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Seed existing enum values
INSERT INTO categories (name) VALUES
    ('tka_saintek'),
    ('tka_soshum'),
    ('smbt');

-- 3. Add category_id FK to tests (nullable first for backfill)
ALTER TABLE tests ADD COLUMN category_id UUID REFERENCES categories(id);

-- 4. Backfill from old enum column
UPDATE tests SET category_id = (
    SELECT id FROM categories WHERE name = tests.category::text
);

-- 5. Make NOT NULL and drop old column + type
ALTER TABLE tests ALTER COLUMN category_id SET NOT NULL;
ALTER TABLE tests DROP COLUMN category;
DROP TYPE test_category;
