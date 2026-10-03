package services

import (
	"path/filepath"
	"testing"

	"dnd-backend/internal/repository"
)

// Writes sharing a key collapse into the latest snapshot while the writer is
// busy; distinct keys still all run. This is what keeps a minion wave from
// turning into one fsync per hit.
func TestPersisterCoalescesWritesPerKey(t *testing.T) {
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPersister(repo)

	blocking := make(chan struct{})
	entered := make(chan struct{})
	p.Enqueue("char:arya", func() {
		close(entered)
		<-blocking // hold the writer busy while we queue more writes
	})
	<-entered

	// Ten snapshots for the same key while the writer is busy → exactly one
	// pending job (last write wins).
	runs := 0
	for i := 0; i < 10; i++ {
		p.Enqueue("char:arya", func() { runs++ })
	}
	// A different key is independent and must still run.
	other := false
	p.Enqueue("world", func() { other = true })

	close(blocking)
	p.Close() // flushes everything, then stops; sync with the writer goroutine

	if runs != 1 {
		t.Errorf("same-key writes coalesced to %d runs, want 1", runs)
	}
	if !other {
		t.Fatal("pending job for a distinct key was dropped on Close")
	}
}

// Enqueue after Close is a no-op (no panic, no write).
func TestPersisterEnqueueAfterCloseIsSafe(t *testing.T) {
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPersister(repo)
	p.Close()
	p.Close() // idempotent
	p.Enqueue("char:x", func() { t.Error("job enqueued after Close must not run") })
}

// Close is idempotent and flushes pending writes first (restart tests rely
// on this to observe the last snapshot).
func TestPersisterCloseFlushesPending(t *testing.T) {
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	p := NewPersister(repo)

	ran := false
	p.Enqueue("char:x", func() { ran = true })
	p.Close()
	if !ran {
		t.Fatal("pending write must be flushed on Close")
	}
	p.Close()
}
