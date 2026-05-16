CREATE TYPE gender AS ENUM ('male', 'female', 'other');
ALTER TABLE users ADD COLUMN gender gender;
ALTER TABLE users ADD COLUMN phone varchar(20);
ALTER TABLE users ADD COLUMN avatar_url text;
