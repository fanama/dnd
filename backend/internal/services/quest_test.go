package services

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
)

func questFixture() domain.Quest {
	return domain.Quest{
		Nom:        "Chasse au Rat",
		Objectif:   "Rapporte la queue du Rat Géant.",
		Obstacle:   []domain.Item{{Nom: "Queue de Rat"}},
		Recompense: []domain.Item{{Nom: "Potion de Soin", IsConsumable: true, Prix: 25}},
	}
}

func TestQuestAcceptAndComplete(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.DMs["dm_quest"] = true
	gm.HandleAction("dm_quest", Action{Type: "dm_add_quest", Destination: "Taverne", Quest: questFixture()})

	gm.World.Players["hero"] = &domain.Player{
		Pseudo:     "hero",
		Characters: []*domain.Character{newTestCharacter("Héro", domain.Stats{Nom: "Héro"}, 100)},
	}
	hero := gm.getLatestCharacter("hero")
	hero.Lieu = "Taverne"

	gm.HandleAction("hero", Action{Type: "accept_quest", QuestName: "Chasse au Rat"})
	if len(hero.Quests) != 1 {
		t.Fatalf("quest should be accepted, got %d", len(hero.Quests))
	}

	// Completing without the obstacle must fail.
	gm.HandleAction("hero", Action{Type: "complete_quest", QuestName: "Chasse au Rat"})
	if len(hero.Quests) != 1 {
		t.Fatalf("quest must not complete without the obstacle item")
	}

	hero.Inventaire = append(hero.Inventaire, domain.Item{Nom: "Queue de Rat"})
	gm.HandleAction("hero", Action{Type: "complete_quest", QuestName: "Chasse au Rat"})
	if len(hero.Quests) != 0 {
		t.Fatalf("quest should be completed, got %d active", len(hero.Quests))
	}
	for _, it := range hero.Inventaire {
		if it.Nom == "Queue de Rat" {
			t.Fatalf("obstacle item should be consumed")
		}
	}
	foundReward := false
	for _, it := range hero.Inventaire {
		if it.Nom == "Potion de Soin" {
			foundReward = true
		}
	}
	if !foundReward {
		t.Fatalf("reward should be granted")
	}
}

func TestQuestCannotBeAcceptedTwice(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.DMs["dm_quest"] = true
	gm.HandleAction("dm_quest", Action{Type: "dm_add_quest", Destination: "Taverne", Quest: questFixture()})

	gm.World.Players["hero"] = &domain.Player{
		Pseudo:     "hero",
		Characters: []*domain.Character{newTestCharacter("Héro", domain.Stats{Nom: "Héro"}, 100)},
	}
	hero := gm.getLatestCharacter("hero")
	hero.Lieu = "Taverne"

	gm.HandleAction("hero", Action{Type: "accept_quest", QuestName: "Chasse au Rat"})
	gm.HandleAction("hero", Action{Type: "accept_quest", QuestName: "Chasse au Rat"})
	if len(hero.Quests) != 1 {
		t.Fatalf("quest must not be accepted twice, got %d", len(hero.Quests))
	}
}

func TestQuestPersistsInWorldAcrossRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "game.db")

	repo1, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	gm1 := NewGameManager(repo1)
	gm1.DMs["dm_quest"] = true
	gm1.HandleAction("dm_quest", Action{Type: "dm_add_quest", Destination: "Taverne", Quest: questFixture()})
	gm1.Close()

	repo2, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo2.Close()
	gm2 := NewGameManager(repo2)
	defer gm2.Close()

	quest := gm2.findLocationQuest("Taverne", "Chasse au Rat")
	if quest == nil || quest.Objectif == "" {
		t.Fatalf("quest should be restored with the location")
	}
}

func TestQuestRepoRoundTrip(t *testing.T) {
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()

	questsJSON, _ := json.Marshal([]domain.Quest{questFixture()})
	if err := repo.SaveCharacter("hero", "Héro", "Guerrier", "Taverne", 100, 100, "[]", "{}", "{}", string(questsJSON)); err != nil {
		t.Fatal(err)
	}
	_, _, _, _, _, _, _, _, questsOut, err := repo.GetCharacter("hero")
	if err != nil {
		t.Fatal(err)
	}
	var quests []domain.Quest
	if err := json.Unmarshal([]byte(questsOut), &quests); err != nil || len(quests) != 1 || quests[0].Nom != "Chasse au Rat" {
		t.Fatalf("quests roundtrip failed: %+v, err=%v", quests, err)
	}
}
