package session

import (
	"testing"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/tmux"
)

// A confirmed-dead row whose session the fresh listing still lacks must stay
// dead without forking tmux (has-session + list-panes per ghost per 30s blew
// the 250 ms status-pass budget at ~130 ghosts).
func TestUpdateStatus_GhostRecheckAnsweredFromSessionCache(t *testing.T) {
	tmux.SeedSessionCacheForTest(t, "some-other-session")
	i := &Instance{
		Tool:           "claude",
		Status:         StatusError,
		CreatedAt:      time.Now().Add(-time.Hour),
		tmuxSession:    &tmux.Session{Name: "ghost-recheck-missing", SocketName: tmux.DefaultSocketName()},
		lastErrorCheck: time.Now().Add(-time.Minute),
	}

	before := tmux.SubprocessStarts()
	if err := i.UpdateStatus(); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if n := tmux.SubprocessStarts() - before; n != 0 {
		t.Errorf("ghost recheck spawned %d tmux subprocesses, want 0", n)
	}
	if got := i.GetStatusThreadSafe(); got != StatusError {
		t.Errorf("Status = %q, want %q", got, StatusError)
	}
	if time.Since(i.lastErrorCheck) > time.Second {
		t.Error("lastErrorCheck not refreshed; the next tick would recheck again")
	}
}

// The cache only short-circuits a negative: a ghost that reappears in the
// listing is revived through the normal path.
func TestUpdateStatus_GhostRevivesWhenSessionCacheHasIt(t *testing.T) {
	tmux.SeedSessionCacheForTest(t, "ghost-recheck-revived")
	i := &Instance{
		Tool:           "claude",
		Status:         StatusStopped,
		CreatedAt:      time.Now().Add(-time.Hour),
		tmuxSession:    &tmux.Session{Name: "ghost-recheck-revived", SocketName: tmux.DefaultSocketName()},
		lastErrorCheck: time.Now().Add(-time.Minute),
	}
	_ = i.UpdateStatus()
	if !i.lastErrorCheck.IsZero() {
		t.Error("revived session kept its ghost recheck timestamp; cache hit was ignored")
	}
}
