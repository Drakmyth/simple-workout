package app

import (
	"net/http"
	"strconv"
	"strings"
)

type workoutsPage struct {
	Workouts []WorkoutSummary
}

type workoutExerciseView struct {
	WorkoutExercise
	Weight  string
	CanUp   bool
	CanDown bool
}

type workoutEditPage struct {
	Workout   Workout
	Items     []workoutExerciseView
	Exercises []Exercise
	Unit      string
}

func (a *App) workoutList(w http.ResponseWriter, _ *http.Request) {
	workouts, err := a.workouts()
	if err != nil {
		a.serverError(w, "load workouts", err)
		return
	}
	a.render(w, "workouts.html", workoutsPage{Workouts: workouts})
}

func (a *App) workoutEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	workout, err := a.workout(id)
	if isNotFound(err) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, "load workout", err)
		return
	}
	exercises, err := a.exercises()
	if err != nil {
		a.serverError(w, "load exercises", err)
		return
	}
	unit, err := a.weightUnit()
	if err != nil {
		a.serverError(w, "load weight unit", err)
		return
	}
	items := make([]workoutExerciseView, len(workout.Exercises))
	for position, exercise := range workout.Exercises {
		weight := ""
		if exercise.Weight != nil {
			weight = strconv.FormatFloat(*exercise.Weight, 'f', -1, 64)
		}
		items[position] = workoutExerciseView{
			WorkoutExercise: exercise,
			Weight:          weight,
			CanUp:           position > 0,
			CanDown:         position < len(workout.Exercises)-1,
		}
	}
	a.render(w, "workout-edit.html", workoutEditPage{Workout: workout, Items: items, Exercises: exercises, Unit: unit})
}

func (a *App) workoutRename(w http.ResponseWriter, r *http.Request) {
	a.workoutAction(w, r, func(id int64) error { return a.renameWorkout(id, r.FormValue("name")) })
}

func (a *App) workoutExerciseAdd(w http.ResponseWriter, r *http.Request) {
	a.workoutAction(w, r, func(id int64) error {
		exerciseID, err := strconv.ParseInt(r.FormValue("exercise_id"), 10, 64)
		if err != nil {
			return err
		}
		return a.addWorkoutExercise(id, exerciseID)
	})
}

func (a *App) workoutExerciseUpdate(w http.ResponseWriter, r *http.Request) {
	position, ok := pathPosition(w, r)
	if !ok {
		return
	}
	a.workoutAction(w, r, func(id int64) error {
		sets, err := strconv.Atoi(r.FormValue("sets"))
		if err != nil {
			return err
		}
		reps, err := strconv.Atoi(r.FormValue("reps"))
		if err != nil {
			return err
		}
		var weight *float64
		if value := strings.TrimSpace(r.FormValue("weight")); value != "" {
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			weight = &parsed
		}
		return a.updateWorkoutExercise(id, position, sets, reps, weight)
	})
}

func (a *App) workoutExerciseRemove(w http.ResponseWriter, r *http.Request) {
	position, ok := pathPosition(w, r)
	if !ok {
		return
	}
	a.workoutAction(w, r, func(id int64) error { return a.removeWorkoutExercise(id, position) })
}

func (a *App) workoutExerciseMove(w http.ResponseWriter, r *http.Request) {
	position, ok := pathPosition(w, r)
	if !ok {
		return
	}
	direction := map[string]int{"up": -1, "down": 1}[r.FormValue("direction")]
	a.workoutAction(w, r, func(id int64) error { return a.moveWorkoutExercise(id, position, direction) })
}

func (a *App) workoutAction(w http.ResponseWriter, r *http.Request, action func(int64) error) {
	id, ok := pathInt(w, r, "id")
	if !ok {
		return
	}
	if err := action(id); err != nil {
		http.Error(w, "Invalid workout update", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/workouts/"+strconv.FormatInt(id, 10)+"/edit", http.StatusSeeOther)
}

func pathPosition(w http.ResponseWriter, r *http.Request) (int, bool) {
	value, err := strconv.Atoi(r.PathValue("position"))
	if err != nil || value < 0 {
		http.Error(w, "Invalid exercise position", http.StatusBadRequest)
		return 0, false
	}
	return value, true
}

func pathInt(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	value, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || value < 1 {
		http.Error(w, "Invalid identifier", http.StatusBadRequest)
		return 0, false
	}
	return value, true
}
