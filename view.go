package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the application's UI based on the current state.
func (m appModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var content string
	switch m.state {
	case StateCalendar:
		content = m.renderCalendarView()
	case StateAddEvent:
		content = m.renderAddEventForm()
	}

	return styleAppBox.Width(m.width).Height(m.height).Render(content)
}

// renderCalendarView generates the string output for the calendar grid.
func (m appModel) renderCalendarView() string {
	header := styleCalendarHeader.Render(m.currentDate.Format("January 2006"))

	// Build days of week header
	daysOfWeek := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	var dowStrs []string
	for _, d := range daysOfWeek {
		dowStrs = append(dowStrs, lipgloss.NewStyle().Width(4).Align(lipgloss.Center).Render(d))
	}
	dowRow := lipgloss.JoinHorizontal(lipgloss.Top, dowStrs...)

	// Determine calendar bounds
	firstOfMonth := time.Date(m.currentDate.Year(), m.currentDate.Month(), 1, 0, 0, 0, 0, m.currentDate.Location())
	startDay := int(firstOfMonth.Weekday())
	daysInMonth := time.Date(m.currentDate.Year(), m.currentDate.Month()+1, 0, 0, 0, 0, 0, m.currentDate.Location()).Day()

	// Build grid
	var gridRows []string
	var currentRow []string

	// Pad start of month
	for i := 0; i < startDay; i++ {
		currentRow = append(currentRow, lipgloss.NewStyle().Width(4).Render(""))
	}

	// Render days
	now := time.Now()
	for day := 1; day <= daysInMonth; day++ {
		dateStr := fmt.Sprintf("%2d", day)

		// Check for events
		hasEvent := false
		for _, e := range m.events {
			if e.Date.Year() == m.currentDate.Year() &&
				e.Date.Month() == m.currentDate.Month() &&
				e.Date.Day() == day {
				hasEvent = true
				break
			}
		}

		style := lipgloss.NewStyle().Width(4).Align(lipgloss.Center)

		isToday := now.Year() == m.currentDate.Year() && now.Month() == m.currentDate.Month() && now.Day() == day
		isSelected := m.selectedDate.Year() == m.currentDate.Year() && m.selectedDate.Month() == m.currentDate.Month() && m.selectedDate.Day() == day

		if isSelected {
			style = styleSelectedDate.Width(4).Align(lipgloss.Center)
		} else if isToday {
			style = styleCurrentDate.Width(4).Align(lipgloss.Center)
		}

		if hasEvent {
			// Add a dot under the number
			dateStr = dateStr + "\n" + styleEventDot.Render("•")
		} else {
			dateStr = dateStr + "\n "
		}

		currentRow = append(currentRow, style.Render(dateStr))

		if len(currentRow) == 7 {
			gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	// Pad end of month
	if len(currentRow) > 0 {
		for len(currentRow) < 7 {
			currentRow = append(currentRow, lipgloss.NewStyle().Width(4).Render(""))
		}
		gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, gridRows...)

	instructions := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\nh/j/k/l or Arrows: Move\nn: Add Event • q/esc: Quit")

	calendarBlock := lipgloss.JoinVertical(lipgloss.Center, header, "", dowRow, grid, instructions)

	// Side Panel for Events
	var sidePanel strings.Builder
	sidePanelTitle := lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).Render(fmt.Sprintf("Events for %s", m.selectedDate.Format("Jan 02")))
	sidePanel.WriteString(sidePanelTitle + "\n\n")

	foundEvent := false
	for _, e := range m.events {
		if e.Date.Year() == m.selectedDate.Year() && e.Date.Month() == m.selectedDate.Month() && e.Date.Day() == m.selectedDate.Day() {
			foundEvent = true

			titleStr := lipgloss.NewStyle().Bold(true).Render(e.Title)
			catStr := lipgloss.NewStyle().Foreground(colorHighlight).Render(fmt.Sprintf("[%s]", e.Category))

			days := e.DaysUntil()
			var daysStr string
			if days == 0 {
				daysStr = lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Render("Today!")
			} else if days < 0 {
				daysStr = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fmt.Sprintf("%d days ago", -days))
			} else {
				daysStr = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(fmt.Sprintf("In %d days", days))
			}

			sidePanel.WriteString(fmt.Sprintf("%s %s\n%s\n\n", titleStr, catStr, daysStr))
		}
	}

	if !foundEvent {
		sidePanel.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No events for this day."))
	}

	sidePanelBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(30).
		Height(15).
		Render(sidePanel.String())

	return lipgloss.JoinHorizontal(lipgloss.Top, calendarBlock, "    ", sidePanelBlock)
}

// renderAddEventForm generates the string output for the event creation form.
func (m appModel) renderAddEventForm() string {
	var b strings.Builder

	b.WriteString("Add New Event\n\n")

	b.WriteString(m.titleInput.View() + "\n")
	b.WriteString(m.dateInput.View() + "\n")
	b.WriteString(m.categoryInput.View() + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: Save • Esc: Cancel • Tab: Next Field"))

	return styleFormContainer.Render(b.String())
}
