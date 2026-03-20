package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// main is the entry point for the application.
// It should:
// 1. Initialize the storage mechanism (e.g., ensure events.json exists).
// 2. Load existing events from storage to populate the initial model.
// 3. Initialize the bubbletea program with the initial model.
// 4. Run the bubbletea program and handle any fatal errors.
func main() {
	autoSync := flag.Bool("autosync", false, "Automatically pull on start and sync on app exit")
	flag.Parse()

	p := tea.NewProgram(initialModel(*autoSync), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
