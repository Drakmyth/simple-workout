ALTER TABLE settings ADD COLUMN keep_screen_on INTEGER NOT NULL DEFAULT 1 CHECK (keep_screen_on IN (0, 1));
