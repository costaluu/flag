package logger

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/styles"
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
	
var chevronRight = styles.SecondaryTextStyle[string]("›")

func Info[T any](msg T) {
	fmt.Printf("%s  🔎  %s  %v\n", chevronRight, styles.InfoTextStyle("info"), styles.SecondaryTextStyle(msg))
}

func Result[T any](msg T) {
	fmt.Printf("%s  🔎  %s  %v\n", chevronRight, styles.InfoTextStyle("info"), styles.SecondaryTextStyle(msg))
	os.Exit(0)
}

func Error[T any](msg T) {
	fmt.Printf("%s  ❌  %s  %v\n", chevronRight, styles.ErrorTextStyle("error"), styles.SecondaryTextStyle(msg))
}

func Fatal[T any](msg T) {
	fmt.Printf("%s  ❌  %s  %v\n", chevronRight, styles.ErrorTextStyle("fatal"), styles.SecondaryTextStyle(msg))
	debug.PrintStack()
	os.Exit(0)
}

func Warning[T any](msg T) {
	fmt.Printf("%s  🚧  %s  %v\n", chevronRight, styles.WarningTextStyle("warning"), styles.SecondaryTextStyle(msg))
}

func Success[T any](msg T) {
	fmt.Printf("%s  ✅  %s  %v\n", chevronRight, styles.SuccessTextStyle("success"), styles.SecondaryTextStyle(msg))
}

func Debug() {
	debug.PrintStack()
	os.Exit(0)
}

