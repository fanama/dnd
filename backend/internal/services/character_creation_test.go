package services

import (
	"testing"

	"dnd-backend/internal/domain"
)

func TestBuildCharacterPerClass(t *testing.T) {
	for _, classe := range []string{"Guerrier", "Magicien", "Voleur", "Clerc", "Barde", "Ranger"} {
		c := buildCharacter(classe, "Héro")
		if c.Stats.Nom != "Héro" || c.Stats.Background != classe {
			t.Fatalf("class/name not set: %+v", c.Stats)
		}
		if c.Lieu != "Taverne" {
			t.Fatalf("should start in Taverne, got %s", c.Lieu)
		}
		if c.CurrentPV != c.Stats.CalculateLifePoints() {
			t.Fatalf("life points mismatch for %s", classe)
		}
		if len(c.Inventaire) < 2 {
			t.Fatalf("starting inventory too small for %s", classe)
		}
		if classe == "Magicien" && len(c.Sorts) == 0 {
			t.Fatalf("Magicien should start with a spell")
		}
	}
	if c := buildCharacter("", "X"); c.Stats.Background != "Guerrier" {
		t.Fatalf("empty class should default to Guerrier")
	}
}

func TestCreateCharacterFromWizard(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["hero"] = &domain.Player{
		Pseudo:     "hero",
		Characters: []*domain.Character{buildCharacter("Guerrier", "placeholder")},
	}

	gm.HandleAction("hero", Action{
		Type: "create_character",
		Stats: domain.Stats{
			Nom: "Lancelot", Background: "Guerrier",
			Force: 15, Constitution: 14, Vitesse: 12,
			Charisme: 11, Instinct: 10, Savoir: 9,
		},
	})

	hero := gm.getLatestCharacter("hero")
	if hero.Stats.Nom != "Lancelot" {
		t.Fatalf("name not applied: %q", hero.Stats.Nom)
	}
	if hero.Stats.Background != "Guerrier" {
		t.Fatalf("class not applied: %q", hero.Stats.Background)
	}
	if hero.Stats.Constitution != 14 {
		t.Fatalf("allocated stats not applied: %+v", hero.Stats)
	}
	if len(hero.Inventaire) != len(defaultStartingItems("Guerrier")) {
		t.Fatalf("inventory not re-rolled for the class")
	}
	if hero.CurrentPV != hero.Stats.CalculateLifePoints() {
		t.Fatalf("life points not recomputed")
	}
}

func TestCreateCharacterClampsStats(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["hero"] = &domain.Player{
		Pseudo:     "hero",
		Characters: []*domain.Character{buildCharacter("Barde", "placeholder")},
	}

	gm.HandleAction("hero", Action{
		Type: "create_character",
		Stats: domain.Stats{
			Nom: "Bardou", Background: "Barde",
			Force: 1, Constitution: 30, Vitesse: 12,
			Charisme: 16, Instinct: 12, Savoir: 10,
		},
	})

	hero := gm.getLatestCharacter("hero")
	if hero.Stats.Force != 3 || hero.Stats.Constitution != 20 {
		t.Fatalf("stats should be clamped to 3..20: %+v", hero.Stats)
	}
}