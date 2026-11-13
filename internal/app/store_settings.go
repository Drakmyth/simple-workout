package app

import "fmt"

func (s *Store) playbackSettings() (PlaybackSettings, error) {
	var settings PlaybackSettings
	if err := s.db.QueryRow(`SELECT rest_seconds, weight_unit, image_preference, keep_screen_on
		FROM settings WHERE id = 1`).Scan(&settings.RestSeconds, &settings.WeightUnit, &settings.ImagePreference, &settings.KeepScreenOn); err != nil {
		return PlaybackSettings{}, fmt.Errorf("query playback settings: %w", err)
	}
	return settings, nil
}

func (s *Store) updatePlaybackSettings(restSeconds int, imagePreference string, keepScreenOn bool) error {
	if restSeconds < 0 || !validImagePreference(imagePreference) {
		return fmt.Errorf("invalid playback settings")
	}
	if _, err := s.db.Exec(`UPDATE settings SET rest_seconds = ?, image_preference = ?, keep_screen_on = ? WHERE id = 1`, restSeconds, imagePreference, keepScreenOn); err != nil {
		return fmt.Errorf("update playback settings: %w", err)
	}
	return nil
}

func (s *Store) weightUnit() (string, error) {
	var unit string
	if err := s.db.QueryRow("SELECT weight_unit FROM settings WHERE id = 1").Scan(&unit); err != nil {
		return "", fmt.Errorf("query weight unit: %w", err)
	}
	return unit, nil
}

func (s *Store) imagePreference() (string, error) {
	var preference string
	if err := s.db.QueryRow("SELECT image_preference FROM settings WHERE id = 1").Scan(&preference); err != nil {
		return "", fmt.Errorf("query image preference: %w", err)
	}
	return preference, nil
}

func (s *Store) setImagePreference(preference string) error {
	if !validImagePreference(preference) {
		return fmt.Errorf("invalid image preference %q", preference)
	}
	if _, err := s.db.Exec("UPDATE settings SET image_preference = ? WHERE id = 1", preference); err != nil {
		return fmt.Errorf("update image preference: %w", err)
	}
	return nil
}
