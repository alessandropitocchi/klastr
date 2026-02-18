package logger

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
)

// Level controls the verbosity of log output.
type Level int

const (
	LevelQuiet  Level = iota // only errors
	LevelNormal              // info + warn + error (default)
	LevelVerbose             // everything including debug
)

// color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// Logger provides structured, leveled logging with an optional prefix.
type Logger struct {
	Level  Level
	Prefix string // e.g. "[argocd]"
	color  bool
}

// New creates a Logger with the given prefix and level.
// Color output is auto-detected based on whether stdout is a TTY.
func New(prefix string, level Level) *Logger {
	return &Logger{
		Level:  level,
		Prefix: prefix,
		color:  isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()),
	}
}

// WithPrefix returns a new Logger with a different prefix, keeping the same level and color settings.
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		Level:  l.Level,
		Prefix: prefix,
		color:  l.color,
	}
}

func (l *Logger) printf(color, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	prefix := l.Prefix
	if prefix != "" {
		prefix += " "
	}
	if l.color && color != "" {
		fmt.Printf("%s%s%s%s", color, prefix, msg, colorReset)
	} else {
		fmt.Printf("%s%s", prefix, msg)
	}
}

// Info logs informational messages (visible at Normal and Verbose levels).
func (l *Logger) Info(format string, args ...any) {
	if l.Level >= LevelNormal {
		l.printf("", format, args...)
	}
}

// Warn logs warning messages (visible at Normal and Verbose levels).
func (l *Logger) Warn(format string, args ...any) {
	if l.Level >= LevelNormal {
		l.printf(colorYellow, format, args...)
	}
}

// Error logs error messages (always visible).
func (l *Logger) Error(format string, args ...any) {
	l.printf(colorRed, format, args...)
}

// Debug logs debug messages (visible only at Verbose level).
func (l *Logger) Debug(format string, args ...any) {
	if l.Level >= LevelVerbose {
		l.printf(colorCyan, format, args...)
	}
}

// Success logs success messages with a check mark (visible at Normal and Verbose levels).
func (l *Logger) Success(format string, args ...any) {
	if l.Level >= LevelNormal {
		l.printf(colorGreen, format, args...)
	}
}
