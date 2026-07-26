package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func cmd(id string, actor string, revision uint64) Command {
	return Command{id, actor, revision, time.Now().UTC()}
}
func TestPauseRequiresBothAcknowledgementsBeforeResume(t *testing.T) {
	s, _ := New("pause-room-01", []string{key(1), key(2)})
	s, _ = s.Pause(cmd("pause-cmd-01", key(1), 0))
	if !errors.Is(s.CanSend(key(2)), ErrSuspended) {
		t.Fatal("send not suspended")
	}
	if _, err := s.Resume(cmd("resume-cmd-01", key(1), 1)); !errors.Is(err, ErrDenied) {
		t.Fatalf("resume=%v", err)
	}
	s, _ = s.Acknowledge(cmd("ack-command-01", key(2), 1))
	s, err := s.Resume(cmd("resume-cmd-02", key(1), 2))
	if err != nil || s.Status() != StatusOpen {
		t.Fatalf("stone=%#v err=%v", s, err)
	}
}
