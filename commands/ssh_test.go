package commands

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	if session, err := newSession("192.168.3.3"); session == nil && err != nil {
		t.Error("create session error", err)
	} else {
		t.Log("session created")
	}
}
