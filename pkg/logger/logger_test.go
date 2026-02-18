package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	l := New("[test]", LevelNormal)
	if l == nil {
		t.Fatal("New() should return non-nil logger")
	}
	if l.Level != LevelNormal {
		t.Errorf("Level = %d, want %d", l.Level, LevelNormal)
	}
	if l.Prefix != "[test]" {
		t.Errorf("Prefix = %q, want %q", l.Prefix, "[test]")
	}
}

func TestWithPrefix(t *testing.T) {
	l := New("[parent]", LevelVerbose)
	child := l.WithPrefix("[child]")
	if child.Prefix != "[child]" {
		t.Errorf("child Prefix = %q, want %q", child.Prefix, "[child]")
	}
	if child.Level != LevelVerbose {
		t.Errorf("child Level = %d, want %d (inherited)", child.Level, LevelVerbose)
	}
}

func TestLevels(t *testing.T) {
	// These should not panic at any level
	levels := []Level{LevelQuiet, LevelNormal, LevelVerbose}
	for _, level := range levels {
		l := New("[test]", level)
		l.Info("info %s\n", "test")
		l.Warn("warn %s\n", "test")
		l.Error("error %s\n", "test")
		l.Debug("debug %s\n", "test")
		l.Success("success %s\n", "test")
	}
}

func TestLevelConstants(t *testing.T) {
	if LevelQuiet >= LevelNormal {
		t.Error("LevelQuiet should be less than LevelNormal")
	}
	if LevelNormal >= LevelVerbose {
		t.Error("LevelNormal should be less than LevelVerbose")
	}
}
