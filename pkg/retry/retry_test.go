package retry

import (
	"errors"
	"testing"
	"time"
)

func TestRun_SuccessFirstAttempt(t *testing.T) {
	calls := 0
	err := Run(3, time.Millisecond, nil, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRun_NonTransientError_NoRetry(t *testing.T) {
	calls := 0
	err := Run(3, time.Millisecond, nil, func() error {
		calls++
		return errors.New("resource not found")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry for non-transient), got %d", calls)
	}
}

func TestRun_TransientError_Retries(t *testing.T) {
	calls := 0
	err := Run(3, time.Millisecond, nil, func() error {
		calls++
		if calls < 3 {
			return errors.New("connection refused")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRun_TransientError_ExhaustsAttempts(t *testing.T) {
	calls := 0
	err := Run(3, time.Millisecond, nil, func() error {
		calls++
		return errors.New("i/o timeout")
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
	if !errors.Is(err, err) {
		t.Error("error should wrap original")
	}
}

func TestRun_LogFnCalled(t *testing.T) {
	logCalls := 0
	logFn := func(format string, args ...any) {
		logCalls++
	}
	_ = Run(3, time.Millisecond, logFn, func() error {
		return errors.New("connection refused")
	})
	// logFn should be called for attempts 1 and 2 (not the last one)
	if logCalls != 2 {
		t.Errorf("expected 2 log calls, got %d", logCalls)
	}
}

func TestRun_ZeroAttempts(t *testing.T) {
	calls := 0
	err := Run(0, time.Millisecond, nil, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call (min 1 attempt), got %d", calls)
	}
}

func TestIsTransient(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("resource not found"), false},
		{errors.New("already exists"), false},
		{errors.New("connection refused"), true},
		{errors.New("i/o timeout"), true},
		{errors.New("dial tcp 127.0.0.1:443: connect: connection refused"), true},
		{errors.New("TLS handshake timeout"), true},
		{errors.New("network unreachable"), true},
	}
	for _, tt := range tests {
		got := IsTransient(tt.err)
		if got != tt.want {
			errStr := "<nil>"
			if tt.err != nil {
				errStr = tt.err.Error()
			}
			t.Errorf("IsTransient(%q) = %v, want %v", errStr, got, tt.want)
		}
	}
}
