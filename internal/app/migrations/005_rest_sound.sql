ALTER TABLE settings ADD COLUMN rest_sound INTEGER NOT NULL DEFAULT 0 CHECK (rest_sound IN (0, 1));
