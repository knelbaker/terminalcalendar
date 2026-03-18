package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// syncEventsWithGit shells out to git to automatically add, commit, and push events.json
func syncEventsWithGit() error {
	// git add events.json
	cmdAdd := exec.Command("git", "add", "events.json")
	if err := cmdAdd.Run(); err != nil {
		return err
	}

	// git commit -m "Auto-sync calendar events"
	// We ignore commit errors since it will fail if there are no changes to the file,
	// which is perfectly fine—we still want to attempt a push just in case!
	cmdCommit := exec.Command("git", "commit", "-m", "Auto-sync calendar events")
	_ = cmdCommit.Run()

	// git push
	cmdPush := exec.Command("git", "push")
	out, err := cmdPush.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}

	return nil
}
