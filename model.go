package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// AppState represents the current view of the application.
// We need to keep track of whether we are viewing the calendar or adding an event.
type AppState int

const (
	StateCalendar AppState = iota
	StateAddEvent
	StateDayView
	StateConfirmDelete
	StateTodoList
)

// Event represents a single event in the calendar.
// It includes logic fields to compute transient properties like "Days until Date".
type Event struct {
	Title     string    `json:"title"`
	Date      time.Time `json:"date"`
	Category  string    `json:"category"`
	Completed bool      `json:"completed"`
	// Note: "Days until Date" should be a computed helper method, not necessarily saved to JSON.
}

// DaysUntil is a helper method to calculate how many days from now this event occurs.
// Returns a negative number if the event is in the past.
func (e *Event) DaysUntil() int {
	now := time.Now()
	
	// Create normalized dates exactly at midnight UTC to completely ignore 
	// daylight saving time and duration shifts when doing math.
	// We use e.Date's Year/Month/Day directly because it was parsed as a literal calendar date.
	startOfNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfEvent := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, time.UTC)
	
	return int(startOfEvent.Sub(startOfNow).Hours() / 24)
}

// appModel is the core state of the Bubble Tea program.
type appModel struct {
	// events is a slice or map to hold loaded calendar events.
	events []Event

	// state tracks the current active view (calendar or form).
	state AppState

	// currentDate tracks the currently viewed month/year on the calendar grid.
	currentDate time.Time

	// selectedDate tracks the user's cursor on the calendar.
	selectedDate time.Time

	// dayEventIndices stores the global indices for all events on the selectedDate currently being viewed.
	dayEventIndices []int

	// dayEventCursor tracks which index in dayEventIndices we are currently focused on.
	dayEventCursor int

	todoIndices []int
	todoCursor  int

	// eventToDeleteIndex tracks the final target index of the event to delete when the confirmation modal is active.
	eventToDeleteIndex int

	isEditing        bool
	eventToEditIndex int

	// syncStatus displays messages about Git background syncing.
	syncStatus string

	// width and height track the current terminal dimensions.
	width  int
	height int

	autoSync bool

	titleInput    textinput.Model
	dateInput     textinput.Model
	categoryInput textinput.Model
	focusIndex    int
}

// initialModel returns the starting state of the application.
func initialModel(autoSync bool) appModel {
	ti := textinput.New()
	ti.Placeholder = "Event Title"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	di := textinput.New()
	di.Placeholder = "YYYY-MM-DD"
	di.CharLimit = 10
	di.Width = 30

	ci := textinput.New()
	ci.Placeholder = "Category"
	ci.CharLimit = 50
	ci.Width = 30

	events, _ := loadEvents("events.json")

	now := time.Now()
	twoMonthsAgo := now.AddDate(0, -2, 0)

	var validEvents []Event
	var purged bool
	for _, e := range events {
		// If the event date is strictly before 2 months ago today
		if e.Date.Before(twoMonthsAgo) {
			purged = true
		} else {
			validEvents = append(validEvents, e)
		}
	}

	if purged {
		events = validEvents
		_ = saveEvents("events.json", events)
	}

	return appModel{
		events:             events,
		state:              StateCalendar,
		currentDate:        now,
		selectedDate:       now,
		dayEventIndices:    []int{},
		dayEventCursor:     0,
		todoIndices:        []int{},
		todoCursor:         0,
		eventToDeleteIndex: -1,
		isEditing:          false,
		eventToEditIndex:   -1,
		syncStatus:         "",
		titleInput:         ti,
		dateInput:          di,
		categoryInput:      ci,
		focusIndex:         0,
		autoSync:           autoSync,
	}
}

// Init initializes the application. This is called once when the program starts.
func (m appModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, textinput.Blink)

	if m.autoSync {
		cmds = append(cmds, func() tea.Msg {
			err := pullEventsWithGit()
			if err != nil {
				return pullCompleteMsg{err: err}
			}
			newEvents, readErr := loadEvents("events.json")
			return pullCompleteMsg{err: readErr, newEvents: newEvents}
		})
	}

	return tea.Batch(cmds...)
}
