package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Define your visual styles and palettes here.
// Keeping them global/package level helps maintain a consistent design system.
var (
	// colorHighlight represents a primary accent color (e.g., for selected days or buttons).
	colorHighlight = lipgloss.Color("#ebdbb2")

	// styleAppBox is the main container style for the application to center content.
	styleAppBox = lipgloss.NewStyle().
			Align(lipgloss.Center, lipgloss.Center)

	// styleCalendarHeader styles the current month/year text.
	styleCalendarHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorHighlight)

	// styleEventDot is the style for the dot indicator underneath dates with events.
	styleEventDot = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ee5555"))

	// styleFormContainer styles the box wrapping the "Add Event" inputs.
	styleFormContainer = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorHighlight).
				Padding(1, 4)

	// styleCurrentDate styles today's actual date.
	styleCurrentDate = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("250"))

	// styleSelectedDate styles the user's cursor on the calendar grid.
	styleSelectedDate = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(colorHighlight)
)
