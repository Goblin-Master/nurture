package zapx

import (
	"testing"

	"go.uber.org/zap"
)

func TestOrNop(t *testing.T) {
	if OrNop(nil) == nil {
		t.Fatal("expected nop logger")
	}
	logger := zap.NewNop().Sugar()
	if OrNop(logger) != logger {
		t.Fatal("expected original logger")
	}
}
