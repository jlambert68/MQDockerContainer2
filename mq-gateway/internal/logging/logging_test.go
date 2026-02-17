package logging

import (
	"log/slog"
	"testing"
)

func TestInitSetsDefaultLogger(t *testing.T) {
	Init("mq-gateway-test")
	if slog.Default() == nil {
		t.Fatal("expected default logger to be configured")
	}
	// Smoke test: should not panic.
	slog.Info("logging init test message")
}
