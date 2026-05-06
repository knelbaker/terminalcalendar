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
		content = m.renderCalendarView(false)
	case StateAddEvent:
		content = m.renderAddEventForm()
	case StateDayView, StateConfirmDelete:
		content = m.renderCalendarView(true)
	case StateTodoList:
		content = m.renderTodoList()
	}

	return styleAppBox.Width(m.width).Height(m.height).Render(content)
}

// renderCalendarView generates the string output for the calendar grid.
func (m appModel) renderCalendarView(showDeleteModal bool) string {
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

	// Pad start of month with previous month dates
	prevMonth := m.currentDate.AddDate(0, -1, 0)
	daysInPrevMonth := time.Date(prevMonth.Year(), prevMonth.Month()+1, 0, 0, 0, 0, 0, prevMonth.Location()).Day()
	
	for i := 0; i < startDay; i++ {
		dayNum := daysInPrevMonth - startDay + 1 + i
		dateStr := fmt.Sprintf("%2d", dayNum)
		
		hasEvent := false
		allCompleted := true
		for _, e := range m.events {
			if e.Date.Year() == prevMonth.Year() && e.Date.Month() == prevMonth.Month() && e.Date.Day() == dayNum {
				hasEvent = true
				if !e.Completed {
					allCompleted = false
				}
			}
		}

		style := lipgloss.NewStyle().Width(4).Align(lipgloss.Center).Foreground(lipgloss.Color("241")) // dim for adjacent month
		if hasEvent {
			dotColor := "241"
			if allCompleted {
				dotColor = "42"
			}
			dateStr = dateStr + "\n" + styleEventDot.Copy().Foreground(lipgloss.Color(dotColor)).Render("•")
		} else {
			dateStr = dateStr + "\n "
		}
		currentRow = append(currentRow, style.Render(dateStr))
	}

	// Render days
	now := time.Now()
	for day := 1; day <= daysInMonth; day++ {
		dateStr := fmt.Sprintf("%2d", day)

		// Check for events
		hasEvent := false
		allCompleted := true
		for _, e := range m.events {
			if e.Date.Year() == m.currentDate.Year() &&
				e.Date.Month() == m.currentDate.Month() &&
				e.Date.Day() == day {
				hasEvent = true
				if !e.Completed {
					allCompleted = false
				}
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
			if allCompleted {
				dateStr = dateStr + "\n" + styleEventDot.Copy().Foreground(lipgloss.Color("42")).Render("•")
			} else {
				dateStr = dateStr + "\n" + styleEventDot.Render("•")
			}
		} else {
			dateStr = dateStr + "\n "
		}

		currentRow = append(currentRow, style.Render(dateStr))

		if len(currentRow) == 7 {
			gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = []string{}
		}
	}

	// Pad end of month with next month dates
	if len(currentRow) > 0 {
		nextMonth := m.currentDate.AddDate(0, 1, 0)
		dayNum := 1
		for len(currentRow) < 7 {
			dateStr := fmt.Sprintf("%2d", dayNum)
			
			hasEvent := false
			allCompleted := true
			for _, e := range m.events {
				if e.Date.Year() == nextMonth.Year() && e.Date.Month() == nextMonth.Month() && e.Date.Day() == dayNum {
					hasEvent = true
					if !e.Completed {
						allCompleted = false
					}
				}
			}

			style := lipgloss.NewStyle().Width(4).Align(lipgloss.Center).Foreground(lipgloss.Color("241")) // dim for adjacent month
			if hasEvent {
				dotColor := "241"
				if allCompleted {
					dotColor = "42"
				}
				dateStr = dateStr + "\n" + styleEventDot.Copy().Foreground(lipgloss.Color(dotColor)).Render("•")
			} else {
				dateStr = dateStr + "\n "
			}
			currentRow = append(currentRow, style.Render(dateStr))
			dayNum++
		}
		gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, gridRows...)

	instructions := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("\nh/j/k/l or Arrows: Move\nEnter: Select Day • n: Add Event • t: Todo List\ns: Sync • p: Pull • q/esc: Quit")

	var syncText string
	if m.syncStatus != "" {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		if len(m.syncStatus) > 11 && m.syncStatus[:11] == "Pull Failed" {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		} else if len(m.syncStatus) > 11 && m.syncStatus[:11] == "Sync Failed" {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red on error
		}
		syncText = "\n\n" + style.Render(m.syncStatus)
	}

	calendarBlock := lipgloss.JoinVertical(lipgloss.Center, header, "", dowRow, grid, instructions, syncText)

	// Panel Content
	var sidePanel strings.Builder

	if showDeleteModal && m.state == StateConfirmDelete && m.eventToDeleteIndex >= 0 && m.eventToDeleteIndex < len(m.events) {
		title := m.events[m.eventToDeleteIndex].Title
		sidePanel.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render("Delete Event?"))
		sidePanel.WriteString(fmt.Sprintf("\n\nAre you sure you want to delete '%s'?\n\n", title))
		sidePanel.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("y: Yes • n: No"))
	} else {
		titleStr := "Events for %s"
		if showDeleteModal && m.state == StateDayView {
			titleStr = "Events on %s"
			sidePanel.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Up/Down: Select • Space: Complete\nd/x: Delete • e: Edit • q/Esc: Back") + "\n\n")
		}

		sidePanelTitle := lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).Render(fmt.Sprintf(titleStr, m.selectedDate.Format("Jan 02")))
		sidePanel.WriteString(sidePanelTitle + "\n\n")

		foundEvent := false
		displayIndex := 0

		for _, e := range m.events {
			if e.Date.Year() == m.selectedDate.Year() && e.Date.Month() == m.selectedDate.Month() && e.Date.Day() == m.selectedDate.Day() {
				foundEvent = true

				titleStyle := lipgloss.NewStyle().Bold(true)
				if m.state == StateDayView && displayIndex == m.dayEventCursor {
					// Highlight the selected event
					titleStyle = titleStyle.Foreground(lipgloss.Color("0")).Background(lipgloss.Color("196"))
				}

				if e.Completed {
					titleStyle = titleStyle.Strikethrough(true).Foreground(lipgloss.Color("241"))
				}

				titleStr := titleStyle.Render(e.Title)
				catStr := lipgloss.NewStyle().Foreground(colorHighlight).Render(fmt.Sprintf("[%s]", e.Category))

				days := e.DaysUntil()
				var daysStr string
				
				dayStyle := lipgloss.NewStyle()
				if e.Completed {
					dayStyle = dayStyle.Strikethrough(true).Foreground(lipgloss.Color("241"))
				}
				
				if days == 0 {
					if !e.Completed {
						dayStyle = dayStyle.Foreground(lipgloss.Color("202"))
					}
					daysStr = dayStyle.Render("Today!")
				} else if days == 1 {
					if !e.Completed {
						dayStyle = dayStyle.Foreground(lipgloss.Color("42"))
					}
					daysStr = dayStyle.Render("Tomorrow")
				} else if days == -1 {
					if !e.Completed {
						dayStyle = dayStyle.Foreground(lipgloss.Color("241"))
					}
					daysStr = dayStyle.Render("Yesterday")
				} else if days < -1 {
					if !e.Completed {
						dayStyle = dayStyle.Foreground(lipgloss.Color("241"))
					}
					daysStr = dayStyle.Render(fmt.Sprintf("%d days ago", -days))
				} else {
					if !e.Completed {
						dayStyle = dayStyle.Foreground(lipgloss.Color("42"))
					}
					daysStr = dayStyle.Render(fmt.Sprintf("In %d days", days))
				}

				sidePanel.WriteString(fmt.Sprintf("%s %s\n%s\n\n", titleStr, catStr, daysStr))
				displayIndex++
			}
		}

		if !foundEvent {
			sidePanel.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No events for this day."))
		}
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

	if m.isEditing {
		b.WriteString("Edit Event\n\n")
	} else {
		b.WriteString("Add New Event\n\n")
	}

	b.WriteString(m.titleInput.View() + "\n")
	b.WriteString(m.dateInput.View() + "\n")
	b.WriteString(m.categoryInput.View() + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: Save • Esc: Cancel • Tab: Next Field"))

	return styleFormContainer.Render(b.String())
}

// renderTodoList generates the string output for the upcoming To-Do list
func (m appModel) renderTodoList() string {
	var b strings.Builder
	
	b.WriteString(lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).Render("Upcoming Todo List"))
	b.WriteString("\n\n")

	if len(m.todoIndices) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("No upcoming tasks! You're all caught up.\n\n"))
	} else {
		// Only display a window around the cursor, e.g. 10 items max
		startIdx := m.todoCursor - 4
		if startIdx < 0 { startIdx = 0 }
		endIdx := startIdx + 8
		if endIdx > len(m.todoIndices) { endIdx = len(m.todoIndices) }
		
		for i := startIdx; i < endIdx; i++ {
			actualIdx := m.todoIndices[i]
			e := m.events[actualIdx]
			
			prefix := "  "
			titleStyle := lipgloss.NewStyle().Bold(true)
			if i == m.todoCursor {
				prefix = "> "
				titleStyle = titleStyle.Foreground(lipgloss.Color("0")).Background(lipgloss.Color("196"))
			}

			if e.Completed {
				titleStyle = titleStyle.Strikethrough(true).Foreground(lipgloss.Color("241"))
			}
			
			days := e.DaysUntil()
			var daysStr string
			
			dayStyle := lipgloss.NewStyle()
			if e.Completed {
				dayStyle = dayStyle.Strikethrough(true).Foreground(lipgloss.Color("241"))
			}
			
			if days == 0 {
				if !e.Completed {
					dayStyle = dayStyle.Foreground(lipgloss.Color("202"))
				}
				daysStr = dayStyle.Render("Today!")
			} else if days == 1 {
				if !e.Completed {
					dayStyle = dayStyle.Foreground(lipgloss.Color("42"))
				}
				daysStr = dayStyle.Render("Tomorrow")
			} else {
				if !e.Completed {
					dayStyle = dayStyle.Foreground(lipgloss.Color("42"))
				}
				daysStr = dayStyle.Render(fmt.Sprintf("In %d days", days))
			}
			
			catStyle := lipgloss.NewStyle().Foreground(colorHighlight)
			if e.Completed {
				catStyle = catStyle.Strikethrough(true).Foreground(lipgloss.Color("241"))
			}
			catStr := catStyle.Render(fmt.Sprintf("[%s]", e.Category))
			
			dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			if e.Completed {
				dateStyle = dateStyle.Strikethrough(true)
			}
			dateFmt := dateStyle.Render(e.Date.Format("Jan 02"))

			b.WriteString(fmt.Sprintf("%s%s %s - %s (%s)\n\n", prefix, titleStyle.Render(e.Title), catStr, daysStr, dateFmt))
		}
	}
	
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Up/Down: Select • Space: Complete • t/q/Esc: Back"))
	
	return styleFormContainer.Render(b.String())
}
