package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Isolate initialModel's cleanup and Update's saves from the user's events.
func isolatedModel(t *testing.T) appModel {
	t.Helper()
	t.Chdir(t.TempDir())
	return initialModel(false)
}

func press(m appModel, key tea.KeyType) appModel {
	next, _ := m.Update(tea.KeyMsg{Type: key})
	return next.(appModel)
}

func letter(m appModel, key rune) appModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
	return next.(appModel)
}

func TestCreateOptionalTime(t *testing.T) {
	for _, value := range []string{"", "00:00", "14:30", "23:59"} {
		t.Run("time="+value, func(t *testing.T) {
			m := isolatedModel(t)
			selected := m.selectedDate
			m = letter(m, 'n')
			m.titleInput.SetValue("Dentist")
			m.categoryInput.SetValue("Personal")
			m.timeInput.SetValue(value)
			m = press(m, tea.KeyEnter)
			if m.state != StateCalendar || len(m.events) != 1 {
				t.Fatalf("event not saved: state=%v error=%s", m.state, m.formError)
			}
			e := m.events[0]
			if e.Time != value || e.Date.Format("2006-01-02") != selected.Format("2006-01-02") || e.DaysUntil() != 0 || !m.selectedDate.Equal(selected) {
				t.Fatalf("unexpected event/date: %+v", e)
			}
			loaded, err := loadEvents("events.json")
			if err != nil || len(loaded) != 1 || loaded[0].Time != value {
				t.Fatalf("time did not survive save/load: %+v, %v", loaded, err)
			}
		})
	}
}

func TestInvalidTimePreservesFormAndEvent(t *testing.T) {
	for _, value := range []string{"24:00", "12:60", "9:30", "12:3", "14:30:00", "noon", " 14:30", "-1:00"} {
		t.Run(value, func(t *testing.T) {
			m := isolatedModel(t)
			m.events = []Event{{Title: "Original", Date: m.selectedDate, Time: "10:00"}}
			m = press(m, tea.KeyEnter)
			m = letter(m, 'e')
			m.titleInput.SetValue("Changed")
			m.categoryInput.SetValue("Personal")
			date := m.dateInput.Value()
			m.timeInput.SetValue(value)
			m = press(m, tea.KeyEnter)
			if m.state != StateAddEvent || m.formError == "" || !strings.Contains(m.renderAddEventForm(), "HH:MM") {
				t.Fatal("invalid time did not keep form open with an error")
			}
			if m.titleInput.Value() != "Changed" || m.categoryInput.Value() != "Personal" || m.dateInput.Value() != date || m.timeInput.Value() != value || m.events[0].Title != "Original" || m.events[0].Time != "10:00" {
				t.Fatal("validation changed inputs or the saved event")
			}
			if _, err := os.Stat("events.json"); !os.IsNotExist(err) {
				t.Fatal("invalid form wrote to storage")
			}
		})
	}
}

func TestEditAddChangeClearTime(t *testing.T) {
	m := isolatedModel(t)
	m.events = []Event{{Title: "Appointment", Date: m.selectedDate, Completed: true}}
	for _, value := range []string{"14:30", "00:00", ""} {
		previous := m.events[0].Time
		m = press(m, tea.KeyEnter)
		m = letter(m, 'e')
		if m.timeInput.Value() != previous {
			t.Fatal("edit did not prefill time")
		}
		m.timeInput.SetValue(value)
		m = press(m, tea.KeyEnter)
		if len(m.events) != 1 || m.events[0].Time != value || !m.events[0].Completed {
			t.Fatalf("edit failed: %+v", m.events)
		}
		loaded, err := loadEvents("events.json")
		if err != nil || len(loaded) != 1 || loaded[0].Time != value {
			t.Fatalf("edit did not persist: %+v, %v", loaded, err)
		}
	}
}

func TestTimeNavigationAndReset(t *testing.T) {
	m := isolatedModel(t)
	m = letter(m, 'n')
	m = press(m, tea.KeyTab)
	m = press(m, tea.KeyTab)
	if !m.timeInput.Focused() || m.titleInput.Focused() || m.dateInput.Focused() || m.categoryInput.Focused() {
		t.Fatal("time field is not exclusively focused")
	}
	for _, r := range "1430" {
		m = letter(m, r)
	}
	if m.timeInput.Value() != "14:30" {
		t.Fatal("time field did not receive typing")
	}
	m = press(m, tea.KeyTab)
	if !m.categoryInput.Focused() || m.timeInput.Focused() {
		t.Fatal("tab did not move to category")
	}
	m = press(m, tea.KeyTab)
	m = press(m, tea.KeyShiftTab)
	if !m.categoryInput.Focused() {
		t.Fatal("reverse tab did not wrap")
	}
	m = press(m, tea.KeyShiftTab)
	if !m.timeInput.Focused() {
		t.Fatal("reverse tab skipped time")
	}
	m = press(m, tea.KeyEsc)
	m = press(m, tea.KeyEnter)
	if m.timeInput.Value() != "" || m.timeInput.Focused() || m.isEditing || m.formError != "" {
		t.Fatal("new form retained previous time or state")
	}
}

func TestTimeAutoColonAndDigitFiltering(t *testing.T) {
	m := isolatedModel(t)
	m = letter(m, 'n')
	m.titleInput.SetValue("Lunch")
	m = press(m, tea.KeyTab)
	m = press(m, tea.KeyTab)
	for i, want := range []string{"1", "12:", "12:3", "12:30"} {
		m = letter(m, rune("1230"[i]))
		if m.timeInput.Value() != want {
			t.Fatalf("digit %d: got %q, want %q", i, m.timeInput.Value(), want)
		}
	}
	for _, r := range "x :/９5" {
		m = letter(m, r)
		if m.timeInput.Value() != "12:30" {
			t.Fatalf("unexpected input accepted: %q", m.timeInput.Value())
		}
	}
	if strings.Contains(m.renderAddEventForm(), "optional") {
		t.Fatal("form still says optional")
	}
	m = press(m, tea.KeyEnter)
	loaded, err := loadEvents("events.json")
	if err != nil || len(loaded) != 1 || loaded[0].Time != "12:30" {
		t.Fatalf("digit entry did not persist formatted time: %+v, %v", loaded, err)
	}
}

func TestTimeAutoColonDeletionAndPaste(t *testing.T) {
	m := isolatedModel(t)
	m = letter(m, 'n')
	m = press(m, tea.KeyTab)
	m = press(m, tea.KeyTab)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("12:30abc"), Paste: true})
	m = next.(appModel)
	if m.timeInput.Value() != "12:30" {
		t.Fatalf("paste got %q", m.timeInput.Value())
	}
	for _, want := range []string{"12:3", "12:", "1", ""} {
		m = press(m, tea.KeyBackspace)
		if m.timeInput.Value() != want {
			t.Fatalf("backspace got %q, want %q", m.timeInput.Value(), want)
		}
	}
	for _, r := range "1230" {
		m = letter(m, r)
	}
	m = press(m, tea.KeyHome)
	m = press(m, tea.KeyRight)
	m = press(m, tea.KeyDelete)
	m = letter(m, '4')
	if m.timeInput.Value() != "14:30" {
		t.Fatalf("editing hours got %q", m.timeInput.Value())
	}
}

func TestLegacyStorageAndTimeDisplay(t *testing.T) {
	m := isolatedModel(t)
	path := filepath.Join(t.TempDir(), "events.json")
	if err := os.WriteFile(path, []byte(`[{"title":"Legacy","date":"2026-10-13T00:00:00Z","category":"Personal","completed":false}]`), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadEvents(path)
	if err != nil || len(loaded) != 1 || loaded[0].Time != "" {
		t.Fatalf("legacy events failed: %+v, %v", loaded, err)
	}
	if err := saveEvents(path, loaded); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || strings.Contains(string(data), `"time"`) {
		t.Fatal("date-only event gained a time field")
	}
	m.events = []Event{{Title: "Dentist", Date: time.Now(), Time: "00:00"}}
	m.todoIndices = []int{0}
	for _, state := range []AppState{StateCalendar, StateDayView, StateTodoList} {
		m.state = state
		m.width, m.height = 100, 40
		if !strings.Contains(m.View(), "00:00") {
			t.Fatalf("time missing in state %v", state)
		}
		m.events[0].Time = ""
		if strings.Contains(m.View(), "00:00") {
			t.Fatalf("date-only event shows midnight in state %v", state)
		}
		m.events[0].Time = "00:00"
	}
}
