package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/constants"
)

func AccentTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Underline(true).
		Foreground(lipgloss.Color(constants.AccentColor)).
		Bold(true).Render()
}

func SecondaryTextStyle[T any](msg T) string {
	var defaultStyle = 
		lipgloss.
			NewStyle().
			SetString(fmt.Sprintf("%s", msg)).
			Foreground(lipgloss.Color("242"))
	
	return defaultStyle.Render()
}

func InfoTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("#3775c4")).
		Underline(true).
		Bold(true).Render()
}

func ErrorTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("#e25822")).
		Underline(true).
		Bold(true).Render()
}

func WarningTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("#f2f27a")).
		Underline(true).
		Bold(true).Render()
}

func SuccessTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("#14b37d")).
		Underline(true).
		Bold(true).Render()
}

func RedTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("160")).
		Underline(true).
		Bold(true).Render()
}

func BlueTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("27")).
		Underline(true).
		Bold(true).Render()
}

func GreenTextStyle(message string) string {
	return lipgloss.
		NewStyle().
		SetString(message).
		Foreground(lipgloss.Color("42")).
		Underline(true).
		Bold(true).Render()
}