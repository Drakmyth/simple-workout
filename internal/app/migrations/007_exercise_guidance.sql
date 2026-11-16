UPDATE exercise_steps
SET instruction = 'Engage your core, then extend one arm and the opposite leg while keeping your lower back gently supported by the floor.'
WHERE exercise_id = (SELECT id FROM exercises WHERE slug = 'dead-bug') AND position = 1;

UPDATE exercises
SET notes = 'Avoid arching your lower back. Only extend your arm and leg as far as you can while maintaining core engagement.'
WHERE slug = 'dead-bug';

UPDATE exercise_steps
SET instruction = 'Press the dumbbells over your chest until your arms are straight, then lower under control until your upper arms touch down gently.'
WHERE exercise_id = (SELECT id FROM exercises WHERE slug = 'dumbbell-floor-press') AND position = 2;

UPDATE exercises
SET notes = 'Keep your shoulder blades supported on the floor rather than reaching them toward the ceiling.
Pause gently when your upper arms reach the floor; do not bounce your elbows.'
WHERE slug = 'dumbbell-floor-press';
