INSERT INTO workouts (id, name) VALUES
    (1, 'Workout A'),
    (2, 'Workout B');

INSERT INTO workout_exercises (workout_id, position, exercise_id, sets, reps, weight) VALUES
    (1, 0, 1, 2, 10, 2.5),
    (1, 1, 2, 2, 10, 2.5),
    (1, 2, 3, 2, 10, 2.5),
    (1, 3, 4, 2, 12, 2.5),
    (1, 4, 5, 2, 6, NULL),
    (2, 0, 6, 2, 10, 2.5),
    (2, 1, 7, 2, 10, NULL),
    (2, 2, 3, 2, 10, 3.0),
    (2, 3, 8, 2, 12, 2.5),
    (2, 4, 9, 2, 15, NULL);
