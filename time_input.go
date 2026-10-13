package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Edit only the digits, then restore the separator and cursor position. This
// makes backspace and delete work across the automatic colon like normal input.
func updateTimeInput(input textinput.Model, msg tea.Msg) (textinput.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyRunes {
		digits := []rune{}
		for _, r := range key.Runes {
			if r >= '0' && r <= '9' {
				digits = append(digits, r)
			}
		}
		if len(digits) == 0 {
			return input, nil
		}
		key.Runes = digits
		msg = key
	}
	position := input.Position()
	if position > 2 {
		position--
	}
	input.SetValue(strings.ReplaceAll(input.Value(), ":", ""))
	input.SetCursor(position)
	input, cmd := input.Update(msg)

	// Also filter clipboard content, which textinput receives asynchronously.
	position = 0
	digits := []rune{}
	for i, r := range []rune(input.Value()) {
		if r >= '0' && r <= '9' && len(digits) < 4 {
			digits = append(digits, r)
			if i < input.Position() {
				position++
			}
		}
	}
	value := string(digits)
	if len(digits) >= 2 {
		value = value[:2] + ":" + value[2:]
		if position >= 2 {
			position++
		}
	}
	input.SetValue(value)
	input.SetCursor(position)
	return input, cmd
}
