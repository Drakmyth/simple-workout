ALTER TABLE settings ADD COLUMN keep_image_visible INTEGER NOT NULL DEFAULT 0 CHECK (keep_image_visible IN (0, 1));
