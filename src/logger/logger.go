package logger

import (
	"fmt"
	"os"
	"runtime/debug"
)

func Info[T any](msg T) {
	fmt.Printf("🔎 %v\n", msg)
}

func Result[T any](msg T) {
	fmt.Printf("🔎 %v\n", msg)
	os.Exit(0)
}

func Error[T any](msg T) {
	fmt.Printf("❌ %v\n", msg)
	debug.PrintStack()
}

func Fatal[T any](msg T) {
	fmt.Printf("❌ %v\n", msg)
	debug.PrintStack()
	os.Exit(0)
}

func Warning[T any](msg T) {
	fmt.Printf("🚨  %v\n", msg)
}

func Success[T any](msg T) {
	fmt.Printf("✅ %v\n", msg)
}

func Debug() {
	debug.PrintStack()
	os.Exit(0)
}

