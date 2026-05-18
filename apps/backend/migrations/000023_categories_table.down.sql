-- Reverse: restore test_category enum and drop categories table.
CREATE TYPE test_category AS ENUM ('tka_saintek', 'tka_soshum', 'smbt');

ALTER TABLE tests ADD COLUMN category test_category;

UPDATE tests SET category = (
    SELECT name::test_category FROM categories WHERE id = tests.category_id
);

ALTER TABLE tests ALTER COLUMN category SET NOT NULL;
ALTER TABLE tests DROP COLUMN category_id;
DROP TABLE categories;
