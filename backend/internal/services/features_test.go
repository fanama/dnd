package services

import (
	"encoding/json"
	"testing"
	"time"

	"dnd-backend/internal/domain"
)

// captureClient registers a fake local client whose outgoing queue we can
// inspect directly without a real websocket connection.
func captureClient(gm *GameManager, pseudo string) *client {
	c := &client{send: make(chan []byte, sendQueueSize), done: make(chan struct{})}
	gm.Connections[pseudo] = c
	return c
}

func TestChatMsgBroadcastWithSanitizedLength(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo: "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya", Background: "Guerrier"}, 100)},
	}
	c := captureClient(gm, "arya")

	long := ""
	for i := 0; i < 100; i++ {
		long += "héllo très long "
	}
	gm.HandleAction("arya", Action{Type: "chat_msg", Message: long})

	select {
	case msg := <-c.send:
		var out map[string]interface{}
		if err := json.Unmarshal(msg, &out); err != nil {
			t.Fatalf("chat broadcast is not valid JSON: %v", err)
		}
		text, _ := out["msg"].(string)
		if runes := len([]rune(text)); runes > 420 {
			t.Errorf("chat message not clamped by sanitizer, runes=%d", runes)
		}
	default:
		t.Fatal("no broadcast received for chat_msg")
	}
	// chat() also pushes a sync snapshot; drain the queue before the next step.
	for {
		select {
		case <-c.send:
		default:
			goto drained
		}
	}
drained:
	gm.HandleAction("arya", Action{Type: "chat_msg", Message: "   "})
	deadline := time.After(200 * time.Millisecond)
	for {
		select {
		case msg := <-c.send:
			var out map[string]interface{}
			if err := json.Unmarshal(msg, &out); err == nil && out["type"] == "chat" {
				t.Fatal("empty chat message should not be broadcast as a chat line")
			}
		case <-deadline:
			return
		}
	}
}

func TestMoveRestrictedToAdjacentLocations(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya"}, 100)},
	}
	char := gm.getLatestCharacter("arya")
	char.Lieu = "Taverne"

	// Taverne links to Donjon, Foret Enchantee, Marais Hante.
	gm.HandleAction("arya", Action{Type: "move", Destination: "Donjon"})
	if char.Lieu != "Donjon" {
		t.Fatalf("adjacent move blocked, lieu=%q", char.Lieu)
	}

	// Temples de l'ombre : Montagne Rocheuse n'est PAS liée à la Taverne mais
	// l'est depuis le Donjon. Depuis le Donjon, Foret Enchantee est hors limites.
	gm.HandleAction("arya", Action{Type: "move", Destination: "Montagne Rocheuse"})
	if char.Lieu != "Montagne Rocheuse" {
		t.Fatalf("adjacent move from Donjon blocked, lieu=%q", char.Lieu)
	}
	gm.HandleAction("arya", Action{Type: "move", Destination: "Foret Enchantee"})
	if char.Lieu == "Foret Enchantee" {
		t.Fatalf("non-adjacent move should be blocked from Montagne Rocheuse, lieu=%q", char.Lieu)
	}

	// Moving to a valid adjacent link from Montagne Rocheuse.
	gm.HandleAction("arya", Action{Type: "move", Destination: "Plaine des Conflits"})
	if char.Lieu != "Plaine des Conflits" {
		t.Fatalf("adjacent move from Montagne Rocheuse blocked, lieu=%q", char.Lieu)
	}
}

func TestXpAndLevelAwardedOnKill(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 16 }
	rollDiceFn = func(count, sides int) int { return sides * count }
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Force: 10}, 100)},
	}
	attacker := gm.World.Players["arya"].Characters[0]
	attacker.Stats.Nom = "Arya"
	attacker.Lieu = "Donjon"
	attacker.Xp = 190 // niveau 2 (1 + 190/100)
	attacker.Equipement.Arme = &domain.Item{Nom: "Marteau", DesDégâts: "d12", BonusDégâts: 5}

	gm.dmAddNPC("Rat Sinistre", "Donjon", 3, "Chaotique Mauvais", domain.Stats{}, "minion", 1)

	gm.actionAttack(attacker, "Rat Sinistre")

	if attacker.Xp != 210 {
		t.Fatalf("xp = %v, want 210 (+20 minion)", attacker.Xp)
	}
	if attacker.Level() != 3 {
		t.Fatalf("level = %d, want 3", attacker.Level())
	}
	if _, still := gm.World.NPCs["Rat Sinistre"]; still {
		t.Fatalf("killed mob should be removed from the world")
	}
}

func TestEncumbranceBlocksLoot(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya", Force: 1}, 100)},
	}
	char := gm.getLatestCharacter("arya")
	char.Lieu = "Donjon"
	char.Inventaire = []domain.Item{{Nom: "Rocher Géant", Encombrement: 14}}

	loc := gm.findLocation("Donjon")
	loc.Objects = append(loc.Objects, domain.Item{Nom: "Trésor Lourd", Encombrement: 4})

	if char.Capacity() != 15 {
		t.Fatalf("Force 1 → Capacity 15, got %v", char.Capacity())
	}
	gm.HandleAction("arya", Action{Type: "loot", LootName: "Trésor Lourd"})
	char = gm.getLatestCharacter("arya")
	if len(char.Inventaire) != 1 {
		t.Fatalf("loot should be blocked by encumbrance, inv=%v", char.Inventaire)
	}
	for _, obj := range gm.findLocation("Donjon").Objects {
		if obj.Nom == "Trésor Lourd" {
			return // object stays on the ground
		}
	}
	t.Fatalf("blocked loot was removed from the ground anyway")
}

func TestEncumbranceBlocksBuy(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya", Force: 1}, 100)},
	}
	char := gm.getLatestCharacter("arya")
	char.Lieu = "Taverne"
	char.Or = 500
	char.Inventaire = []domain.Item{{Nom: "Rocher Géant", Encombrement: 14}}

	// Taverne shop index 1 = "Dague d'Argent" (BonusDégâts=3 → default weight 4 → 14+4>15).
	heavyIdx := -1
	for i, item := range gm.findLocation("Taverne").Commerce {
		if item.BonusDégâts > 0 {
			heavyIdx = i
			break
		}
	}
	if heavyIdx < 0 {
		t.Fatal("could not locate a heavy item in the Taverne shop")
	}
	before := len(char.Inventaire)
	gm.HandleAction("arya", Action{Type: "buy_item", ItemIndex: heavyIdx})
	char = gm.getLatestCharacter("arya")
	if char.Or != 500 || len(char.Inventaire) != before {
		t.Fatalf("buy should be blocked by encumbrance: or=%v inv=%v", char.Or, char.Inventaire)
	}
}

func TestItemWeightAndCapacityHelpers(t *testing.T) {
	if got := domain.ItemWeight(domain.Item{IsConsumable: true}); got != 0.5 {
		t.Errorf("consumable weight = %v, want 0.5", got)
	}
	if got := domain.ItemWeight(domain.Item{BonusArmure: 2}); got != 8 {
		t.Errorf("armor weight = %v, want 8", got)
	}
	if got := domain.ItemWeight(domain.Item{BonusDégâts: 3}); got != 4 {
		t.Errorf("weapon weight = %v, want 4", got)
	}
	if got := domain.ItemWeight(domain.Item{Nom: "Clé"}); got != 1 {
		t.Errorf("misc weight = %v, want 1", got)
	}
	if got := domain.ItemWeight(domain.Item{Encombrement: 12}); got != 12 {
		t.Errorf("explicit encumbrance = %v, want 12", got)
	}
	if got := (&domain.Character{}).Capacity(); got != 150 {
		t.Errorf("default capacity = %v, want 150", got)
	}
}

func TestNonBlockingBroadcastCapped(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	full := &client{send: make(chan []byte, sendQueueSize), done: make(chan struct{})}
	gm.Connections["full"] = full
	for i := 0; i < sendQueueSize; i++ {
		full.send <- []byte("x")
	}

	gm.chat("test de saturation de la file")
	gm.chat("second message")

	// The enqueue must not block the caller even with a full queue.
	select {
	case <-time.After(200 * time.Millisecond):
		t.Fatal("broadcast blocked on a full queue")
	default:
	}

	// Draining must still deliver queued messages first.
	drained := 0
	for {
		select {
		case <-full.send:
			drained++
		default:
			if drained < sendQueueSize {
				t.Fatalf("dropped queued messages during saturation, drained=%d", drained)
			}
			return
		}
	}
}