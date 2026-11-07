ALTER TABLE settings ADD COLUMN image_preference TEXT NOT NULL DEFAULT 'alternating'
    CHECK (image_preference IN ('male', 'female', 'random', 'alternating'));

DROP TABLE exercise_images;
CREATE TABLE exercise_images (
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    variant TEXT NOT NULL CHECK (variant IN ('male', 'female')),
    path TEXT NOT NULL,
    alt_text TEXT NOT NULL,
    attribution TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (exercise_id, variant)
) STRICT;

INSERT INTO exercises (id, slug, name, notes) VALUES
    (1, 'goblet-squat', 'Goblet squat', 'Use a load that allows a controlled depth with both heels planted.'),
    (2, 'dumbbell-floor-press', 'Dumbbell floor press', 'Pause gently when the upper arms reach the floor; do not bounce the elbows.'),
    (3, 'one-arm-dumbbell-row', 'One-arm dumbbell row', 'Complete all repetitions on one side before switching sides.'),
    (4, 'biceps-curl', 'Biceps curl', 'Keep the upper arms quiet and stop before the shoulders roll forward.'),
    (5, 'dead-bug', 'Dead bug', 'Move only as far as you can while keeping the lower back gently supported.'),
    (6, 'dumbbell-romanian-deadlift', 'Dumbbell Romanian deadlift', 'The movement comes from the hips; stop when hamstring tension limits the hinge.'),
    (7, 'incline-push-up', 'Incline push-up', 'Use a stable surface. A higher surface makes the exercise easier.'),
    (8, 'standing-hammer-curl', 'Standing hammer curl', 'Keep the palms facing inward throughout the repetition.'),
    (9, 'standing-calf-raise', 'Standing calf raise', 'Use support for balance only and pause briefly at the top.');

INSERT INTO exercise_steps (exercise_id, position, instruction) VALUES
    (1, 0, 'Stand with feet about shoulder-width apart and hold one dumbbell vertically at the chest.'),
    (1, 1, 'Brace the torso, sit the hips down, and let the knees track in line with the toes.'),
    (1, 2, 'Keep the heels planted, then drive through the feet to stand tall.'),
    (2, 0, 'Lie on the floor with knees bent, feet planted, and a dumbbell in each hand.'),
    (2, 1, 'Set the upper arms lightly on the floor with forearms vertical and wrists neutral.'),
    (2, 2, 'Press over the chest, then lower under control until the upper arms touch down gently.'),
    (3, 0, 'Take a staggered stance and support one hand on a stable bench.'),
    (3, 1, 'Let the working arm hang vertically while keeping the back neutral and hips square.'),
    (3, 2, 'Pull the elbow toward the lower ribs, pause, and lower the dumbbell without rotating.'),
    (4, 0, 'Stand tall with a dumbbell in each hand, palms forward, and elbows close to the torso.'),
    (4, 1, 'Curl the dumbbells toward the shoulders without swinging or moving the elbows forward.'),
    (4, 2, 'Lower slowly until the arms are straight while keeping the wrists neutral.'),
    (5, 0, 'Lie on your back with arms vertical and hips and knees bent to 90 degrees.'),
    (5, 1, 'Brace gently, then extend one arm and the opposite leg without arching the lower back.'),
    (5, 2, 'Return to the start and repeat on the other side with the torso still.'),
    (6, 0, 'Stand with feet hip-width apart and hold the dumbbells against the front of the thighs.'),
    (6, 1, 'Soften the knees and push the hips backward while keeping the spine neutral.'),
    (6, 2, 'Keep the dumbbells close to the legs, then squeeze the glutes to stand tall.'),
    (7, 0, 'Place both hands on a stable waist-high surface and walk the feet back.'),
    (7, 1, 'Hold a straight line from head through heels and lower the chest toward the surface.'),
    (7, 2, 'Keep the elbows 30 to 45 degrees from the torso and press back to straight arms.'),
    (8, 0, 'Stand tall with a dumbbell in each hand, palms inward, and elbows close to the torso.'),
    (8, 1, 'Curl both dumbbells while keeping the palms facing each other and the torso still.'),
    (8, 2, 'Lower under control to straight arms without shrugging or bending the wrists.'),
    (9, 0, 'Stand on a flat floor with feet hip-width apart and one hand lightly supported.'),
    (9, 1, 'Rise through the balls of both feet while keeping the knees straight and ankles aligned.'),
    (9, 2, 'Pause at the top, then lower the heels slowly to the floor.');

INSERT INTO exercise_images (exercise_id, variant, path, alt_text) VALUES
    (1, 'male', '/static/images/exercises/goblet-squat-male.webp', 'Male model showing the standing and lowered positions of a goblet squat.'),
    (1, 'female', '/static/images/exercises/goblet-squat-female.webp', 'Female model showing the standing and lowered positions of a goblet squat.'),
    (2, 'male', '/static/images/exercises/dumbbell-floor-press-male.webp', 'Male model showing the lowered and pressed positions of a dumbbell floor press.'),
    (2, 'female', '/static/images/exercises/dumbbell-floor-press-female.webp', 'Female model showing the lowered and pressed positions of a dumbbell floor press.'),
    (3, 'male', '/static/images/exercises/one-arm-dumbbell-row-male.webp', 'Male model showing the lowered and pulled positions of a one-arm dumbbell row.'),
    (3, 'female', '/static/images/exercises/one-arm-dumbbell-row-female.webp', 'Female model showing the lowered and pulled positions of a one-arm dumbbell row.'),
    (4, 'male', '/static/images/exercises/biceps-curl-male.webp', 'Male model showing the lowered and curled positions of a biceps curl.'),
    (4, 'female', '/static/images/exercises/biceps-curl-female.webp', 'Female model showing the lowered and curled positions of a biceps curl.'),
    (5, 'male', '/static/images/exercises/dead-bug-male.webp', 'Male model showing the starting and extended positions of a dead bug.'),
    (5, 'female', '/static/images/exercises/dead-bug-female.webp', 'Female model showing the starting and extended positions of a dead bug.'),
    (6, 'male', '/static/images/exercises/dumbbell-romanian-deadlift-male.webp', 'Male model showing the standing and hinged positions of a dumbbell Romanian deadlift.'),
    (6, 'female', '/static/images/exercises/dumbbell-romanian-deadlift-female.webp', 'Female model showing the standing and hinged positions of a dumbbell Romanian deadlift.'),
    (7, 'male', '/static/images/exercises/incline-push-up-male.webp', 'Male model showing the raised and lowered positions of an incline push-up.'),
    (7, 'female', '/static/images/exercises/incline-push-up-female.webp', 'Female model showing the raised and lowered positions of an incline push-up.'),
    (8, 'male', '/static/images/exercises/standing-hammer-curl-male.webp', 'Male model showing the lowered and curled positions of a standing hammer curl.'),
    (8, 'female', '/static/images/exercises/standing-hammer-curl-female.webp', 'Female model showing the lowered and curled positions of a standing hammer curl.'),
    (9, 'male', '/static/images/exercises/standing-calf-raise-male.webp', 'Male model showing the lowered and raised positions of a standing calf raise.'),
    (9, 'female', '/static/images/exercises/standing-calf-raise-female.webp', 'Female model showing the lowered and raised positions of a standing calf raise.');
