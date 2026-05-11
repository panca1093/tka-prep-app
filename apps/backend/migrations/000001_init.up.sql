-- Enable UUID generation for all entities defined in spec.md section 6.
-- Every table uses gen_random_uuid() as default for the id column.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
