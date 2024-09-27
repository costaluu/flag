package utils

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/constants"
)

func AccentTextUnderLine(message string) string {
	return lipgloss.
		NewStyle().
		Padding(0, 1).
		SetString(message).
		Background(lipgloss.Color(constants.AccentColor)).
		Foreground(lipgloss.Color("255")).
		Bold(true).Render()
}