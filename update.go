package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// syncCompleteMsg is used to report the result of Git push back to the main thread
type syncCompleteMsg struct{ err error }

// pullCompleteMsg is used to report the result of Git pull back to the main thread
type pullCompleteMsg struct {
	err       error
	newEvents []Event
}

// Update handles incoming messages (keypresses, window resizing, etc.)
// and updates the model state accordingly.
func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case syncCompleteMsg:
		if msg.err != nil {
			m.syncStatus = "Sync Failed: " + msg.err.Error()
		} else {
			m.syncStatus = "Synced!"
		}
		return m, nil

	case pullCompleteMsg:
		if msg.err != nil {
			m.syncStatus = "Pull Failed: " + msg.err.Error()
		} else {
			m.syncStatus = "Pull Complete!"
			m.events = msg.newEvents
		}
		return m, nil

	case tea.KeyMsg:
		// Global quit keys
		if msg.String() == "ctrl+c" {
			if m.autoSync {
				_ = syncEventsWithGit()
			}
			return m, tea.Quit
		}

		switch m.state {
		case StateCalendar:
			switch msg.String() {
			case "q", "esc":
				if m.autoSync {
					_ = syncEventsWithGit()
				}
				return m, tea.Quit
			case "enter":
				// Gather ALL events on the selected date
				m.dayEventIndices = []int{}
				for i, e := range m.events {
					if e.Date.Year() == m.selectedDate.Year() &&
						e.Date.Month() == m.selectedDate.Month() &&
						e.Date.Day() == m.selectedDate.Day() {
						m.dayEventIndices = append(m.dayEventIndices, i)
					}
				}
				
				if len(m.dayEventIndices) > 0 {
					// There are events on this day, jump into Day View to browse them
					m.dayEventCursor = 0
					m.state = StateDayView
				} else {
					// There are zero events on this day, so pop open the Add Event Form
					m.state = StateAddEvent
					m.titleInput.SetValue("")
					// Pre-fill date input with the selected date
					m.dateInput.SetValue(m.selectedDate.Format("2006-01-02"))
					m.categoryInput.SetValue("")
					m.focusIndex = 0
					m.titleInput.Focus()
					m.dateInput.Blur()
					m.categoryInput.Blur()
				}
			case "n":
				// Explicitly switch to add event state regardless of whether day has events
				m.state = StateAddEvent
				m.isEditing = false
				m.eventToEditIndex = -1
				m.titleInput.SetValue("")
				m.dateInput.SetValue(m.selectedDate.Format("2006-01-02"))
				m.categoryInput.SetValue("")
				m.focusIndex = 0
				m.titleInput.Focus()
				m.dateInput.Blur()
				m.categoryInput.Blur()
			case "s":
				m.syncStatus = "Syncing..."
				return m, func() tea.Msg {
					err := syncEventsWithGit()
					return syncCompleteMsg{err: err}
				}
			case "p":
				m.syncStatus = "Pulling..."
				return m, func() tea.Msg {
					err := pullEventsWithGit()
					if err != nil {
						return pullCompleteMsg{err: err}
					}
					// If pull succeeded, read the newly updated file
					newEvents, readErr := loadEvents("events.json")
					return pullCompleteMsg{err: readErr, newEvents: newEvents}
				}
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
				m.isEditing = false
				m.eventToEditIndex = -1
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
					if m.isEditing && m.eventToEditIndex >= 0 && m.eventToEditIndex < len(m.events) {
						// Update existing event
						m.events[m.eventToEditIndex].Title = m.titleInput.Value()
						m.events[m.eventToEditIndex].Date = parsedDate
						m.events[m.eventToEditIndex].Category = m.categoryInput.Value()
						m.isEditing = false
						m.eventToEditIndex = -1
					} else {
						// Create new event
						newEvent := Event{
							Title:    m.titleInput.Value(),
							Date:     parsedDate,
							Category: m.categoryInput.Value(),
						}
						m.events = append(m.events, newEvent)
					}

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

		case StateDayView:
			switch msg.String() {
			case "esc", "q", "n", "h", "left":
				m.state = StateCalendar
			case "up", "k":
				m.dayEventCursor--
				if m.dayEventCursor < 0 {
					m.dayEventCursor = len(m.dayEventIndices) - 1
				}
			case "down", "j":
				m.dayEventCursor++
				if m.dayEventCursor >= len(m.dayEventIndices) {
					m.dayEventCursor = 0
				}
			case "d", "x", "delete":
				m.eventToDeleteIndex = m.dayEventIndices[m.dayEventCursor]
				m.state = StateConfirmDelete
			case "e":
				// Switch to edit mode
				m.isEditing = true
				m.eventToEditIndex = m.dayEventIndices[m.dayEventCursor]
				
				// Populate form inputs
				targetEvent := m.events[m.eventToEditIndex]
				m.titleInput.SetValue(targetEvent.Title)
				m.dateInput.SetValue(targetEvent.Date.Format("2006-01-02"))
				m.categoryInput.SetValue(targetEvent.Category)
				
				m.focusIndex = 0
				m.titleInput.Focus()
				m.dateInput.Blur()
				m.categoryInput.Blur()
				m.state = StateAddEvent
			case " ":
				// Toggle Completed status
				actualIndex := m.dayEventIndices[m.dayEventCursor]
				m.events[actualIndex].Completed = !m.events[actualIndex].Completed
				_ = saveEvents("events.json", m.events)
			}

		case StateConfirmDelete:
			switch msg.String() {
			case "y", "enter": // Confirm
				if m.eventToDeleteIndex >= 0 && m.eventToDeleteIndex < len(m.events) {
					m.events = append(m.events[:m.eventToDeleteIndex], m.events[m.eventToDeleteIndex+1:]...)
					_ = saveEvents("events.json", m.events)
				}
				m.eventToDeleteIndex = -1
				
				// Once the item is deleted out of the master m.events slice, we need to completely rebuild 
				// the local day array to stay synchronized and see if we should remain in DayView.
				m.dayEventIndices = []int{}
				for i, e := range m.events {
					if e.Date.Year() == m.selectedDate.Year() &&
						e.Date.Month() == m.selectedDate.Month() &&
						e.Date.Day() == m.selectedDate.Day() {
						m.dayEventIndices = append(m.dayEventIndices, i)
					}
				}
				
				if len(m.dayEventIndices) > 0 {
					// Clamp cursor to avoid out-of-bounds panics
					if m.dayEventCursor >= len(m.dayEventIndices) {
						m.dayEventCursor = len(m.dayEventIndices) - 1
					}
					m.state = StateDayView
				} else {
					// We've deleted the last event on this day, boot back to root.
					m.state = StateCalendar
				}
			case "n", "esc": // Cancel
				m.state = StateDayView
				m.eventToDeleteIndex = -1
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}
