package main

import (
	"encoding/json"
	"os"
)

const settingsFile = "settings.json"

type settings struct {
	Scale int `json:"scale"`
}

func clampScale(scale int) int {
	return max(-1, min(3, scale))
}

func loadScale() int {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return 0
	}
	var saved settings
	if json.Unmarshal(data, &saved) != nil {
		return 0
	}
	return clampScale(saved.Scale)
}

func saveScale(scale int) error {
	data, err := json.MarshalIndent(settings{Scale: clampScale(scale)}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFile, append(data, '\n'), 0644)
}
