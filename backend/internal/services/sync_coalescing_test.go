package services

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
)

// newProbeClient wraps a fake connection that records every broadcast
// (writeLoop intentionally not started — tests read from the channel).
func newProbeClient() *client {
	return &client{
		send: make(chan []byte, 64),
		done: make(chan struct{}),
	}
}

// newSyncTestManager builds a manager with one player and one DM probe
// attached. The sync loop is NOT started: tests drive flushSync directly so
// assertions stay deterministic.
func newSyncTestManager(t *testing.T) (*GameManager, *client, *client) {
	t.Helper()
	repo, err := repository.NewSQLiteRepository(filepath.Join(t.TempDir(), "game.db"))
	if err != nil {
		t.Fatal(err)
	}
	gm := NewGameManager(repo)
	hero := newTestCharacter("Arya", domain.Stats{Nom: "Arya", Background: "Guerrier"}, 100)
	hero.Lieu = "Taverne"
	hero.Inventaire = []domain.Item{{Nom: "Épée Longue", BonusDégâts: 2}}
	gm.World.Players["arya"] = &domain.Player{Pseudo: "arya", Characters: []*domain.Character{hero}}

	playerProbe := newProbeClient()
	dmProbe := newProbeClient()
	gm.Connections["arya"] = playerProbe
	gm.DMs["arya"] = false
	gm.Connections["dm"] = dmProbe
	gm.DMs["dm"] = true
	return gm, playerProbe, dmProbe
}

// recvRaw reads the next broadcast from a probe.
func recvRaw(t *testing.T, c *client) []byte {
	t.Helper()
	select {
	case raw := <-c.send:
		return raw
	case <-time.After(time.Second):
		t.Fatal("expected a message")
		return nil
	}
}

// recv reads and decodes the next broadcast.
func recv(t *testing.T, c *client) map[string]interface{} {
	t.Helper()
	var msg map[string]interface{}
	if err := json.Unmarshal(recvRaw(t, c), &msg); err != nil {
		t.Fatalf("invalid broadcast: %v", err)
	}
	return msg
}

// expect asserts the type of the next broadcast.
func expect(t *testing.T, c *client, want string) map[string]interface{} {
	t.Helper()
	msg := recv(t, c)
	if got := msg["type"]; got != want {
		t.Fatalf("expected message type %q, got %q", want, got)
	}
	return msg
}

func assertDrained(t *testing.T, c *client, label string) {
	t.Helper()
	select {
	case raw := <-c.send:
		t.Fatalf("%s: unexpected extra message: %s", label, raw)
	default:
	}
}

// A chat is broadcast immediately but must not mark the state dirty: no
// save, no sync payload.
func TestChatDoesNotTriggerSync(t *testing.T) {
	gm, player, _ := newSyncTestManager(t)
	defer gm.Close()

	gm.HandleAction("arya", Action{Type: "chat_msg", Message: "salut tout le monde"})

	expect(t, player, "chat")
	gm.flushSync()
	assertDrained(t, player, "chat must not trigger a sync")
}

// Three mutations before a flush produce ONE coalesced sync: the receiver's
// own full character ("moi") followed by the slim roster — and nothing on a
// second flush. The DM gets a single complete sync.
func TestMutationsAreCoalescedIntoOneSync(t *testing.T) {
	gm, player, dm := newSyncTestManager(t)
	defer gm.Close()

	gm.HandleAction("arya", Action{Type: "equip_item", ItemName: "Épée Longue"})
	gm.HandleAction("arya", Action{Type: "unequip_item", Slot: "weapon"})
	gm.HandleAction("arya", Action{Type: "equip_item", ItemName: "Épée Longue"})

	// Chat lines are broadcast immediately — but no state sync yet.
	expect(t, player, "chat")
	expect(t, player, "chat")
	expect(t, player, "chat")
	assertDrained(t, player, "before flush (sync must be coalesced)")

	gm.flushSync()

	// 1. own full character
	raw := recvRaw(t, player)
	var me MeMessage
	if err := json.Unmarshal(raw, &me); err != nil || me.Type != "moi" {
		t.Fatalf("first flush message should be moi: %s (%v)", raw, err)
	}
	if me.Moi.Nom != "Arya" || me.Moi.Lieu != "Taverne" || len(me.Moi.Inventaire) != 1 || me.Moi.Equipement.Arme == nil {
		t.Fatalf("moi should carry the full sheet: %+v", me.Moi)
	}

	// 2. slim roster (decoded both ways to prove it carries no full sheet)
	raw = recvRaw(t, player)
	var roster PlayerSyncMessage
	if err := json.Unmarshal(raw, &roster); err != nil || roster.Type != "sync" {
		t.Fatalf("second flush message should be sync: %s (%v)", raw, err)
	}
	if roster.Liste["arya"].Lieu != "Taverne" || roster.Liste["arya"].Nom != "Arya" {
		t.Fatalf("roster entry wrong: %+v", roster.Liste["arya"])
	}
	var generic map[string]interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	entry, _ := generic["liste"].(map[string]interface{})["arya"].(map[string]interface{})
	if _, hasFull := entry["inventaire"]; hasFull {
		t.Fatal("roster must stay slim (no inventaire)")
	}
	assertDrained(t, player, "after flush")

	// The DM received exactly one full sync with complete entries.
	expect(t, dm, "chat")
	expect(t, dm, "chat")
	expect(t, dm, "chat")
	var dmSync SyncMessage
	if err := json.Unmarshal(recvRaw(t, dm), &dmSync); err != nil || dmSync.Type != "sync" {
		t.Fatalf("dm should receive a sync: %v", err)
	}
	dmEntry := dmSync.Liste["arya"]
	if dmEntry.Stats.Nom != "Arya" || len(dmEntry.Inventaire) != 1 || dmEntry.Equipement.Arme == nil {
		t.Fatalf("dm sync must be complete: %+v", dmEntry)
	}
	assertDrained(t, dm, "after flush")

	// A flush with nothing dirty sends nothing.
	gm.flushSync()
	assertDrained(t, player, "idle flush (player)")
	assertDrained(t, dm, "idle flush (dm)")
}

// A missed attack changes no state: chat and event fire immediately, but no
// save and no sync are queued.
func TestMissDoesNotTriggerSync(t *testing.T) {
	oldRoll := rollD20Fn
	rollD20Fn = func() int { return 5 }
	defer func() { rollD20Fn = oldRoll }()

	gm, player, _ := newSyncTestManager(t)
	defer gm.Close()

	// Second hero at the same location, AC 10 vs roll 5 + mod(Force 10)=0 → miss.
	foe := newTestCharacter("Brutus", domain.Stats{Nom: "Brutus", Vitesse: 10}, 50)
	foe.Lieu = "Taverne"
	gm.World.Players["brutus"] = &domain.Player{Pseudo: "brutus", Characters: []*domain.Character{foe}}

	gm.HandleAction("arya", Action{Type: "attack", Cible: "brutus"})

	expect(t, player, "chat")
	expect(t, player, "event")
	gm.flushSync()
	assertDrained(t, player, "miss must not trigger a sync")
}

// Locations travel in their own message, only when they actually changed.
func TestLocationsAreSentOnlyWhenChanged(t *testing.T) {
	gm, player, dm := newSyncTestManager(t)
	defer gm.Close()

	// A state sync with no location change: no locations message at all.
	gm.HandleAction("arya", Action{Type: "equip_item", ItemName: "Épée Longue"})
	expect(t, player, "chat")
	gm.flushSync()
	expect(t, player, "moi")
	expect(t, player, "sync")
	assertDrained(t, player, "no locations expected")

	// The DM changes location data → next flush broadcasts it to everyone.
	gm.HandleAction("dm", Action{
		Type:        "dm_teleport_item_add",
		Destination: "Taverne",
		Item:        domain.Item{Nom: "Joyau Éclatant", Prix: 80},
	})
	expect(t, player, "chat")

	gm.flushSync()

	var locs LocationsMessage
	if err := json.Unmarshal(recvRaw(t, player), &locs); err != nil || locs.Type != "locations" {
		t.Fatalf("expected locations message: %v", err)
	}
	found := false
	for _, l := range locs.Locations {
		if l.Nom != "Taverne" {
			continue
		}
		for _, obj := range l.Objects {
			if obj.Nom == "Joyau Éclatant" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("new ground loot missing from the locations broadcast: %+v", locs.Locations)
	}
	// The DM action also marked the state dirty → moi + sync follow.
	expect(t, player, "moi")
	expect(t, player, "sync")
	assertDrained(t, player, "player done")

	// The DM probe saw: equip chat + first flush sync, then the DM-action
	// chat, then the second flush (locations + sync).
	expect(t, dm, "chat")
	expect(t, dm, "sync")
	expect(t, dm, "chat")
	expect(t, dm, "locations")
	expect(t, dm, "sync")
	assertDrained(t, dm, "dm done")
}

// A joining client receives the location data right away (it is otherwise
// only broadcast on change), then the coalesced sync on the next flush.
func TestConnectSendsLocationsThenSync(t *testing.T) {
	gm, _, _ := newSyncTestManager(t)
	defer gm.Close()

	probe := newProbeClient()
	gm.Connect("newbie", probe, map[string]string{"classe": "Guerrier", "nom_personnage": "Nouveau"})

	expect(t, probe, "init_new_char")
	expect(t, probe, "chat") // « Nouveau arrive en quête d'un destin... »
	var locs LocationsMessage
	if err := json.Unmarshal(recvRaw(t, probe), &locs); err != nil || locs.Type != "locations" {
		t.Fatalf("expected locations on connect: %v", err)
	}
	if len(locs.Locations) == 0 {
		t.Fatal("connect locations payload empty")
	}

	gm.flushSync()
	expect(t, probe, "moi")
	expect(t, probe, "sync")
}

// The background sync loop delivers dirty state on its own ticker, and Close
// (idempotent) stops it without deadlocking.
func TestStartSyncLoopDeliversAndCloses(t *testing.T) {
	gm, _, _ := newSyncTestManager(t)

	probe := newProbeClient()
	gm.Connect("newbie", probe, map[string]string{"classe": "Guerrier", "nom_personnage": "Nouveau"})
	// Drain the connect messages, then start the loop.
	expect(t, probe, "init_new_char")
	expect(t, probe, "chat")
	expect(t, probe, "locations")

	gm.StartSyncLoop()
	// The connect NotifyChange is still pending → the loop must flush it.
	expect(t, probe, "moi")
	expect(t, probe, "sync")

	gm.Close()
	gm.Close() // idempotent
}

// A refused action must not queue a database write.
func TestRefusedActionDoesNotPersist(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "game.db")
	repo, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	gm := NewGameManager(repo)
	hero := newTestCharacter("Arya", domain.Stats{Nom: "Arya"}, 100)
	hero.Lieu = "Taverne"
	gm.World.Players["arya"] = &domain.Player{Pseudo: "arya", Characters: []*domain.Character{hero}}

	// Destination unreachable from the Taverne → refused, state unchanged.
	gm.HandleAction("arya", Action{Type: "move", Destination: "Quelque Part"})
	gm.Close()

	if _, _, _, _, _, _, _, _, _, _, _, err := repo.GetCharacter("arya"); err == nil {
		t.Fatal("a refused action must not write the character to the database")
	}
	repo.Close()
}
