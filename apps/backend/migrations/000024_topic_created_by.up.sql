ALTER TABLE topics ADD COLUMN created_by UUID REFERENCES users(id);
