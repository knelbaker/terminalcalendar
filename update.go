package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages (keypresses, window resizing, etc.)
// and updates the model state accordingly.
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit keys
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		switch m.state {
		case StateCalendar:
			switch msg.String() {
			case "n":
				// Switch to add event state
				m.state = StateAddEvent
				m.titleInput.SetValue("")
				// Pre-fill date input with the selected date
				m.dateInput.SetValue(m.selectedDate.Format("2006-01-02"))
				m.categoryInput.SetValue("")
				m.focusIndex = 0
				m.titleInput.Focus()
				m.dateInput.Blur()
				m.categoryInput.Blur()
			case "right", "l":
				m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
				if m.selectedDate.Month() != m.currentDate.Month() {
					m.currentDate = m.currentDate.AddDate(0, 1, 0)
				}
			case "left", "h":
				m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
				if m.selectedDate.Month() != m.currentDate.Month() {
					m.currentDate = m.currentDate.AddDate(0, -1, 0)
				}
			case "down", "j":
				m.selectedDate = m.selectedDate.AddDate(0, 0, 7)
				if m.selectedDate.Month() != m.currentDate.Month() {
					m.currentDate = m.currentDate.AddDate(0, 1, 0)
				}
			case "up", "k":
				m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
				if m.selectedDate.Month() != m.currentDate.Month() {
					m.currentDate = m.currentDate.AddDate(0, -1, 0)
				}
			}

		case StateAddEvent:
			switch msg.String() {
			case "esc":
				m.state = StateCalendar
			case "tab", "shift+tab":
				s := msg.String()

				if s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > 2 {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = 2
				}

				m.titleInput.Blur()
				m.dateInput.Blur()
				m.categoryInput.Blur()

				switch m.focusIndex {
				case 0:
					m.titleInput.Focus()
				case 1:
					m.dateInput.Focus()
				case 2:
					m.categoryInput.Focus()
				}
			case "enter":
				// Process form
				parsedDate, err := time.Parse("2006-01-02", m.dateInput.Value())
				if err == nil && m.titleInput.Value() != "" {
					newEvent := Event{
						Title:    m.titleInput.Value(),
						Date:     parsedDate,
						Category: m.categoryInput.Value(),
					}
					m.events = append(m.events, newEvent)
					
					// Save to JSON
					_ = saveEvents("events.json", m.events)
					
					// Return to calendar
					m.state = StateCalendar
				}
			}

			// Pass key to focused input
			switch m.focusIndex {
			case 0:
				m.titleInput, cmd = m.titleInput.Update(msg)
			case 1:
				m.dateInput, cmd = m.dateInput.Update(msg)
			case 2:
				m.categoryInput, cmd = m.categoryInput.Update(msg)
			}
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}
