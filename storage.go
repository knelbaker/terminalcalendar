package main

import (
	"encoding/json"
	"errors"
	"os"
)

// loadEvents reads the events.json file from the disk and parses it into a slice of Event structs.
func loadEvents(filename string) ([]Event, error) {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return []Event{}, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// saveEvents takes a slice of Event structs and writes it to the events.json file.
func saveEvents(filename string, events []Event) error {
	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
