package services

import (
	"path/filepath"
	"testing"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
)

func newTestGameManager(t *testing.T) *GameManager {
	t.Helper()
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	return NewGameManager(repo)
}

func TestWorldPersistsAcrossRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "game.db")

	repo1, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	gm1 := NewGameManager(repo1)
	if len(gm1.World.NPCs) != 7 {
		t.Fatalf("expected 7 default NPCs, got %d", len(gm1.World.NPCs))
	}

	gm1.DMs["dm_test"] = true
	gm1.HandleAction("dm_test", Action{
		Type:        "dm_add_npc",
		ItemName:    "Bob le Barde",
		Destination: "Taverne",
		PV:          50,
		Alignement:  "Neutre Bon",
	})
	gm1.Close()

	repo2, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo2.Close()

	gm2 := NewGameManager(repo2)
	defer gm2.Close()

	if npc, ok := gm2.World.NPCs["Bob le Barde"]; !ok || npc.CurrentPV != 50 {
		t.Fatalf("NPC Bob was not restored from DB: %+v", gm2.World.NPCs["Bob le Barde"])
	}
	if len(gm2.World.NPCs) != 8 {
		t.Fatalf("expected 8 NPCs (7 defaults + Bob), got %d", len(gm2.World.NPCs))
	}
	if gm2.World.Locations[0].Nom != "Taverne" {
		t.Fatalf("locations were not restored, got %q", gm2.World.Locations[0].Nom)
	}
}

func TestPlayerActionRouter(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya"}, 100)},
	}

	gm.HandleAction("arya", Action{Type: "move", Destination: "Donjon"})

	char := gm.getLatestCharacter("arya")
	if char == nil || char.Lieu != "Donjon" {
		t.Fatalf("move action was not routed, lieu = %v", char.Lieu)
	}
}

func TestDMUnknownActionIsIgnored(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.DMs["dm_test"] = true
	gm.HandleAction("dm_test", Action{Type: "dm_nope"})

	if len(gm.World.NPCs) != 7 {
		t.Fatalf("unknown DM action mutated the world: %d NPCs", len(gm.World.NPCs))
	}
}
