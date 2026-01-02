package axy

import "log/slog"

// logger is the package-wide logger used by default Base hooks.
var logger *slog.Logger = slog.Default()

// SetLogger overrides the package logger.
//
// If not set, slog.Default() is used.
func SetLogger(l *slog.Logger) {
	logger = l
}
