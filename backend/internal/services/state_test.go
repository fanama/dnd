package services

import (
	"encoding/json"
	"testing"

	"dnd-backend/internal/domain"
)

func TestStateExportImportRoundTrip(t *testing.T) {
	src := newTestGameManager(t)
	defer src.Close()

	src.DMs["dm_save"] = true
	src.HandleAction("dm_save", Action{Type: "dm_add_quest", Destination: "Taverne", Quest: questFixture()})
	src.World.Players["hero"] = &domain.Player{
		Pseudo:     "hero",
		Characters: []*domain.Character{newTestCharacter("Héro", domain.Stats{Nom: "Héro"}, 100)},
	}
	hero := src.getLatestCharacter("hero")
	hero.Lieu = "Taverne"
	hero.Quests = append(hero.Quests, questFixture())
	src.World.NPCs["Bandit"] = &domain.Character{
		ID:         domain.Character{}.ID,
		Stats:      domain.Stats{Nom: "Bandit"},
		CurrentPV:  30,
		Alignement: "Chaotique Mauvais",
	}
	hero.Inventaire = append(hero.Inventaire, domain.Item{Nom: "Épée Rouillée", BonusDégâts: 2})
	src.saveCharacterState("hero", hero)

	// Export as JSON, like the DM download.
	snap := src.snapshotState()
	data, _ := json.Marshal(snap)

	// Load it into a fresh world.
	dst := newTestGameManager(t)
	defer dst.Close()
	dst.DMs["dm_save"] = true
	dst.HandleAction("dm_save", Action{Type: "dm_load_state", Payload: string(data)})

	// NPCs restored.
	if _, ok := dst.World.NPCs["Bandit"]; !ok {
		t.Fatalf("NPC should be restored")
	}
	if len(dst.World.NPCs) != len(src.World.NPCs) {
		t.Fatalf("NPC count mismatch: %d vs %d", len(dst.World.NPCs), len(src.World.NPCs))
	}
	// Locations restored (default world replaced by imported one).
	quest := dst.findLocationQuest("Taverne", "Chasse au Rat")
	if quest == nil {
		t.Fatalf("location quest should be restored")
	}
	if len(dst.World.Locations) != len(src.World.Locations) {
		t.Fatalf("locations count mismatch: %d vs %d", len(dst.World.Locations), len(src.World.Locations))
	}
	// Players restored with their character state.
	h := dst.getLatestCharacter("hero")
	if h == nil {
		t.Fatalf("player character should be restored")
	}
	if h.Lieu != "Taverne" || len(h.Quests) != 1 || len(h.Inventaire) != 1 {
		t.Fatalf("character state not restored properly: %+v", h)
	}
}

func TestStateLoadRejectsInvalidPayload(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()
	gm.DMs["dm_save"] = true

	gm.HandleAction("dm_save", Action{Type: "dm_load_state", Payload: "not json"})
	if len(gm.World.NPCs) == 0 {
		t.Fatalf("world should be untouched after invalid load")
	}
}