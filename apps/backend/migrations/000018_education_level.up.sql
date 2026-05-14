CREATE TYPE education_level AS ENUM ('sd', 'smp', 'sma', 'smk');

ALTER TABLE users ADD COLUMN education_level education_level;
ALTER TABLE tests ADD COLUMN education_level education_level;
