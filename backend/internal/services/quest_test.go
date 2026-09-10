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
		Nom:         "Chasse au Rat",
		Objectif:    "Rapporte la queue du Rat Géant.",
		Obstacle:    "Un Rat Géant garde les égouts.",
		Recompense:  []domain.Item{{Nom: "Potion de Soin", IsConsumable: true, Prix: 25}},
		Information: "Les Rats Géants redoutent le feu.",
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

	gm.HandleAction("hero", Action{Type: "complete_quest", QuestName: "Chasse au Rat"})
	if len(hero.Quests) != 0 {
		t.Fatalf("quest should be completed, got %d active", len(hero.Quests))
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
	if quests[0].Obstacle != "Un Rat Géant garde les égouts." || quests[0].Information != "Les Rats Géants redoutent le feu." {
		t.Fatalf("quest fields roundtrip failed: %+v", quests[0])
	}
}

func TestQuestMigratesFromItemList(t *testing.T) {
	oldJSON := `{"nom":"Vieux","objectif":"obj","obstacle":[{"nom":"Pierre","prix":0}],"recompense":[{"nom":"Or","prix":10}]}`
	var q domain.Quest
	if err := json.Unmarshal([]byte(oldJSON), &q); err != nil {
		t.Fatalf("old-format quest should unmarshal: %v", err)
	}
	if q.Obstacle != "Pierre" {
		t.Fatalf("obstacle should become a description, got %q", q.Obstacle)
	}
	if len(q.Recompense) != 1 || q.Recompense[0].Nom != "Or" {
		t.Fatalf("rewards should be preserved")
	}
}

func TestDmEditQuestReplacesAtIndex(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.DMs["dm_quest"] = true
	gm.HandleAction("dm_quest", Action{Type: "dm_add_quest", Destination: "Taverne", Quest: questFixture()})

	edited := questFixture()
	edited.Obstacle = "Deux Rats Géants gardent les égouts."
	edited.Information = "Ils craignent aussi l'argent."
	gm.HandleAction("dm_quest", Action{Type: "dm_edit_quest", Destination: "Taverne", QuestIndex: 0, Quest: edited})

	quest := gm.findLocationQuest("Taverne", "Chasse au Rat")
	if quest == nil {
		t.Fatalf("quest should still exist")
	}
	if quest.Obstacle != "Deux Rats Géants gardent les égouts." || quest.Information != "Ils craignent aussi l'argent." {
		t.Fatalf("quest should be edited, got %+v", quest)
	}
}
