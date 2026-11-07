CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    rest_seconds INTEGER NOT NULL DEFAULT 60 CHECK (rest_seconds >= 0),
    weight_unit TEXT NOT NULL DEFAULT 'kg' CHECK (weight_unit IN ('kg', 'lb'))
) STRICT;

INSERT OR IGNORE INTO settings (id) VALUES (1);

CREATE TABLE IF NOT EXISTS exercises (
    id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE IF NOT EXISTS exercise_steps (
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position >= 0),
    instruction TEXT NOT NULL,
    PRIMARY KEY (exercise_id, position)
) STRICT;

CREATE TABLE IF NOT EXISTS exercise_images (
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position >= 0),
    path TEXT NOT NULL,
    alt_text TEXT NOT NULL,
    attribution TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (exercise_id, position)
) STRICT;

CREATE TABLE IF NOT EXISTS workouts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE IF NOT EXISTS workout_exercises (
    workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position >= 0),
    exercise_id INTEGER NOT NULL REFERENCES exercises(id),
    sets INTEGER NOT NULL CHECK (sets > 0),
    reps INTEGER NOT NULL CHECK (reps > 0),
    weight REAL CHECK (weight IS NULL OR weight >= 0),
    PRIMARY KEY (workout_id, position)
) STRICT;
