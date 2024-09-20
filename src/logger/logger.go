package logger

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/charmbracelet/lipgloss"
)

var infoStyle = 
		lipgloss.
		NewStyle().
		SetString("info").
		Foreground(lipgloss.Color("#3775c4")).
		Underline(true).
		Bold(true)

var errorStyle = 
		lipgloss.
		NewStyle().
		SetString("error").
		Foreground(lipgloss.Color("#e25822")).
		Underline(true).
		Bold(true)

var warningStyle = 
		lipgloss.
		NewStyle().
		SetString("warning").
		Foreground(lipgloss.Color("#f2f27a")).
		Underline(true).
		Bold(true)

var sucessStyle = 
		lipgloss.
		NewStyle().
		SetString("success").
		Foreground(lipgloss.Color("#14b37d")).
		Underline(true).
		Bold(true)
	
var chevronRight = renderDefault[string]("›")

func renderDefault[T any](msg T) string {
	var defaultStyle = 
		lipgloss.
			NewStyle().
			SetString(fmt.Sprintf("%s", msg)).
			Foreground(lipgloss.Color("242"))
	
	return defaultStyle.Render()
}

func Info[T any](msg T) {
	fmt.Printf("%s  🔎  %s  %v\n", chevronRight, infoStyle.Render(), renderDefault(msg))
}

func Result[T any](msg T) {
	fmt.Printf("%s  🔎  %s  %v\n", chevronRight, infoStyle.Render(), renderDefault(msg))
	os.Exit(0)
}

func Error[T any](msg T) {
	fmt.Printf("%s  ❌  %s  %v\n", chevronRight, errorStyle.Render(), renderDefault(msg))
}

func Fatal[T any](msg T) {
	fmt.Printf("%s  ❌  %s  %v\n", chevronRight, errorStyle.Render(), renderDefault(msg))
	debug.PrintStack()
	os.Exit(0)
}

func Warning[T any](msg T) {
	fmt.Printf("%s  🚧  %s  %v\n", chevronRight, warningStyle.Render(), renderDefault(msg))
}

func Success[T any](msg T) {
	fmt.Printf("%s  ✅  %s  %v\n", chevronRight, sucessStyle.Render(), renderDefault(msg))
}

func Debug() {
	debug.PrintStack()
	os.Exit(0)
}

