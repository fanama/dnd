package services

import (
	"path/filepath"
	"testing"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
)

func TestGoldEconomy(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya", Background: "Guerrier"}, 100)},
	}
	hero := gm.getLatestCharacter("arya")
	hero.Lieu = "Taverne"
	hero.Or = 100
	hero.Inventaire = []domain.Item{
		{Nom: "Potion de Soin", IsConsumable: true, Prix: 25},
		{Nom: "Épée Rouillée", BonusDégâts: 2, Prix: 40},
	}

	// Selling gives the item's price in gold.
	gm.HandleAction("arya", Action{Type: "sell_item", ItemIndex: 0})
	hero = gm.getLatestCharacter("arya")
	if hero.Or != 125 || len(hero.Inventaire) != 1 {
		t.Fatalf("after sell: or=%v inv=%d, want 125 and 1 item", hero.Or, len(hero.Inventaire))
	}

	// Buying pulls from the current location's shop stock.
	taverne := gm.findLocation("Taverne")
	if len(taverne.Commerce) == 0 {
		t.Fatalf("Taverne should have a seeded shop stock")
	}
	first := taverne.Commerce[0]
	stock := len(taverne.Commerce)
	gm.HandleAction("arya", Action{Type: "buy_item", ItemIndex: 0})
	hero = gm.getLatestCharacter("arya")
	if hero.Or != 125-first.Prix {
		t.Fatalf("after buy: or=%v, want %v", hero.Or, 125-first.Prix)
	}
	if len(taverne.Commerce) != stock-1 {
		t.Fatalf("shop stock have shrunk; got %d, want %d", len(taverne.Commerce), stock-1)
	}
	found := false
	for _, it := range hero.Inventaire {
		if it.Nom == first.Nom {
			found = true
		}
	}
	if !found {
		t.Fatalf("bought item missing from inventory")
	}

	// Cannot buy without enough gold.
	poorGold := gm.findLocation("Taverne").Commerce[0].Prix
	hero.Or = 0
	before := len(hero.Inventaire)
	gm.HandleAction("arya", Action{Type: "buy_item", ItemIndex: 0})
	if len(gm.getLatestCharacter("arya").Inventaire) != before {
		t.Fatalf("buy should fail without funds (item priced %v)", poorGold)
	}
}

func TestGoldAndShopPersistAcrossRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "game.db")

	repo1, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	gm1 := NewGameManager(repo1)
	gm1.DMs["dm_save"] = true

	// Seed the shop and a player's gold, then persist via Close().
	gm1.HandleAction("dm_save", Action{Type: "dm_shop_add", Destination: "Taverne", Item: domain.Item{Nom: "Griffe de Dragon", Prix: 99}})
	gm1.World.Players["ary"] = &domain.Player{
		Pseudo:     "ary",
		Characters: []*domain.Character{newTestCharacter("Ary", domain.Stats{Nom: "Ary"}, 100)},
	}
	ary := gm1.getLatestCharacter("ary")
	ary.Lieu = "Taverne"
	ary.Or = 42
	gm1.saveCharacterState("ary", ary)
	gm1.Close()

	repo2, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo2.Close()
	gm2 := NewGameManager(repo2)
	defer gm2.Close()

	loc := gm2.findLocation("Taverne")
	found := false
	for _, it := range loc.Commerce {
		if it.Nom == "Griffe de Dragon" && it.Prix == 99 {
			found = true
		}
	}
	if !found {
		t.Fatalf("shop item did not survive restart: %+v", loc.Commerce)
	}
	// Players rehydrate from the characters table on connect; check the row directly.
	_, _, _, _, _, savedOr, _, _, _, _, _, err := repo2.GetCharacter("ary")
	if err != nil {
		t.Fatalf("player row missing: %v", err)
	}
	if savedOr != 42 {
		t.Fatalf("gold did not survive restart: or=%v, want 42", savedOr)
	}
}

func TestDMSetOr(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.DMs["dm_save"] = true
	gm.World.Players["bob"] = &domain.Player{
		Pseudo:     "bob",
		Characters: []*domain.Character{newTestCharacter("Bob", domain.Stats{Nom: "Bob"}, 50)},
	}
	gm.HandleAction("dm_save", Action{Type: "dm_edit_or", TargetPlayer: "bob", Or: 500})
	if got := gm.getLatestCharacter("bob").Or; got != 500 {
		t.Fatalf("dm_edit_or: or=%v, want 500", got)
	}
}