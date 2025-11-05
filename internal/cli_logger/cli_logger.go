package clilogger

import (
	"fmt"
	"log/slog"
)

func LogError(s string) {
	slog.Error(fmt.Sprintln("\t🚫\t\b\b\b\b\b\b", s))
}

func LogSuccess(s string) {
	slog.Info(fmt.Sprintln("\t✅\t\b\b\b\b\b\b", s))
}

func LogInfo(s string) {
	slog.Info(fmt.Sprintln("\tℹ️\t\b\b\b\b\b\b", s))
}

func LogCLI(s string) {
	slog.Info(fmt.Sprintln("\t▶️\t\b\b\b\b\b\b", s))
}

func LogAlert(s string) {
	slog.Info(fmt.Sprintln("\t⚠️\t\b\b\b\b\b\b", s))
}
