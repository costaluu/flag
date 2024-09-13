package logger

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/costaluu/flag/constants"
)

func Info[T any](msg T) {
	fmt.Printf("%s %v\n", constants.InfoMark.Render(), msg)
}

func Result[T any](msg T) {
	fmt.Printf("%s %v\n", constants.InfoMark.Render(), msg)
	os.Exit(0)
}

func Error[T any](msg T) {
	fmt.Printf("%s %v\n", constants.XMark.Render(), msg)
	debug.PrintStack()
}

func Fatal[T any](msg T) {
	fmt.Printf("%s %v\n", constants.XMark.Render(), msg)
	debug.PrintStack()
	os.Exit(0)
}

func Warning[T any](msg T) {
	fmt.Printf("%s  %v\n", constants.WarningMark.Render(), msg)
}

func Success[T any](msg T) {
	fmt.Printf("%s %v\n", constants.CheckMark.Render(), msg)
}

func Debug() {
	debug.PrintStack()
	os.Exit(0)
}

