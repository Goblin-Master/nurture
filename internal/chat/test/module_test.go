package test

import (
	"testing"
	"time"

	"nurture/internal/chat"
)

func TestModuleCloseStopsHub(t *testing.T) {
	module := chat.NewModule(chat.Deps{})

	module.Close()

	select {
	case <-module.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for module shutdown")
	}
}
