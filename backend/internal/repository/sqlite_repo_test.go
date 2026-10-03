package repository

import (
	"path/filepath"
	"testing"
)

// The repository must open in WAL mode: it turns each commit into a cheap
// append (the old rollback journal cost a full fsync per write, ~3ms) and
// lets readers overlap the single writer.
func TestJournalModeWAL(t *testing.T) {
	repo, err := NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	var mode string
	if err := repo.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want wal", mode)
	}

	var timeout int
	if err := repo.db.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil {
		t.Fatalf("PRAGMA busy_timeout: %v", err)
	}
	if timeout < 1000 {
		t.Errorf("busy_timeout = %d, want >= 1000ms", timeout)
	}
}

// Round-trip still works with the new DSN (WAL conversion of a fresh file).
func TestSaveAndGetCharacter(t *testing.T) {
	repo, err := NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	if err := repo.SaveCharacter("arya", "Arya", "Guerrier", "Taverne", 42, 50, "[]", "{}", "{}", "[]", 10, 100); err != nil {
		t.Fatal(err)
	}
	nom, classe, lieu, pv, _, or, xp, _, _, _, _, err := repo.GetCharacter("arya")
	if err != nil {
		t.Fatal(err)
	}
	if nom != "Arya" || classe != "Guerrier" || lieu != "Taverne" || pv != 42 || or != 10 || xp != 100 {
		t.Errorf("round trip mismatch: %s %s %s pv=%d or=%v xp=%v", nom, classe, lieu, pv, or, xp)
	}
}
