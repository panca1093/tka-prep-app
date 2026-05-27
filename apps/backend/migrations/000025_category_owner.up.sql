ALTER TABLE categories ADD COLUMN description TEXT;
ALTER TABLE categories ADD COLUMN created_by UUID REFERENCES users(id);
