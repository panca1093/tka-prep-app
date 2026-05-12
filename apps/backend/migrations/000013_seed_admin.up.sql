-- Seed a default admin user for local development.
-- Email: admin@tkaprep.dev  Password: Admin@tkaprep123
-- Change the password immediately in any non-local environment.
INSERT INTO users (id, name, email, password_hash, role, status)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Super Admin',
    'admin@tkaprep.dev',
    '$2a$12$ZD0kFmwatbK9syVaTRW3W.vZeX6iujTCJcZVO0jhiyApBRCfA5yBW',
    'admin',
    'active'
) ON CONFLICT (email) DO NOTHING;
