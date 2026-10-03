package services

import (
	"encoding/json"
	"sync"
	"time"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
	"github.com/google/uuid"
)

// syncInterval bounds how often state changes are rebuilt and broadcast:
// actions only mark the sync dirty (NotifyChange) and the sync loop flushes
// at most once per interval, coalescing a burst of actions into a single
// payload instead of one full-state message per action.
const syncInterval = 60 * time.Millisecond

// Action is a single player or DM command sent from a client.
type Action struct {
	Type         string       `json:"type"`
	Cible        string       `json:"cible,omitempty"`
	Destination  string       `json:"destination,omitempty"`
	Sort         string       `json:"sort,omitempty"`
	ItemIndex    int          `json:"item_index,omitempty"`
	ItemName     string       `json:"item_name,omitempty"`
	LootName     string       `json:"loot_name,omitempty"`
	TargetPlayer string       `json:"target_player,omitempty"`
	Stats        domain.Stats `json:"stats,omitempty"`
	Item         domain.Item  `json:"item,omitempty"`
	Spell        domain.Sort  `json:"spell,omitempty"`
	PV           float64      `json:"pv,omitempty"`
	Alignement   string       `json:"alignement,omitempty"`
	Overwrite    bool         `json:"overwrite,omitempty"`
	NewName      string       `json:"new_name,omitempty"`
	LocationBg   string       `json:"location_bg,omitempty"`
	Slot         string       `json:"slot,omitempty"`
	MobType      string       `json:"mob_type,omitempty"`
	Count        int          `json:"count,omitempty"`
	Names        []string     `json:"names,omitempty"`
	Quest        domain.Quest `json:"quest,omitempty"`
	QuestName    string       `json:"quest_name,omitempty"`
	QuestIndex   int          `json:"quest_index,omitempty"`
	Payload      string       `json:"payload,omitempty"`
	Message      string       `json:"message,omitempty"`
	Or           float64      `json:"or,omitempty"`
	Pseudo       string       `json:"-"`
}

type GameManager struct {
	Connections   map[string]*client
	World         *domain.World
	Repo          *repository.SQLiteRepository
	DMs           map[string]bool
	persister     *Persister
	playerActions map[string]playerActionFn
	dmActions     map[string]dmActionFn
	mu            sync.Mutex
	syncDirty     bool          // state changed, a sync broadcast is pending
	locsDirty     bool          // locations changed, a locations broadcast is pending
	syncStop      chan struct{}
	syncDone      chan struct{}
	syncStarted   bool
	closeOnce     sync.Once
}

func NewGameManager(repo *repository.SQLiteRepository) *GameManager {
	gm := &GameManager{
		Connections: make(map[string]*client),
		DMs:         make(map[string]bool),
		Repo:        repo,
		persister:   NewPersister(repo),
		syncStop:    make(chan struct{}),
		syncDone:    make(chan struct{}),
		World: &domain.World{
			ID:        uuid.New(),
			Players:   make(map[string]*domain.Player),
			NPCs:      make(map[string]*domain.Character),
			Locations: defaultLocations(),
		},
	}
	gm.registerActions()
	gm.loadOrSeedWorld()
	return gm
}

// StartSyncLoop starts the coalescing broadcast loop. It is separate from
// NewGameManager so unit tests stay deterministic (they call flushSync
// directly); main() is expected to call it right after construction.
func (gm *GameManager) StartSyncLoop() {
	gm.syncStarted = true
	go gm.syncLoop()
}

func (gm *GameManager) syncLoop() {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()
	defer close(gm.syncDone)
	for {
		select {
		case <-gm.syncStop:
			gm.flushSync() // deliver the final state before shutting down
			return
		case <-ticker.C:
			gm.flushSync()
		}
	}
}

// Close stops the sync loop and flushes pending database writes. Idempotent.
func (gm *GameManager) Close() {
	gm.closeOnce.Do(func() {
		close(gm.syncStop)
		if gm.syncStarted {
			<-gm.syncDone
		}
		gm.persister.Close()
	})
}

// DisconnectAll drops every active connection (used during shutdown).
func (gm *GameManager) DisconnectAll() {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for pseudo, c := range gm.Connections {
		c.close()
		delete(gm.Connections, pseudo)
	}
}

// restoreCharacter loads a character from the database. It performs file I/O
// and must be called WITHOUT the game lock held, so a login waiting behind
// pending SQLite writes cannot stall every other session.
func (gm *GameManager) restoreCharacter(pseudo string) (*domain.Character, error) {
	_, _, lieu, pv, _, or, xp, invStr, statsStr, equipStr, questsStr, err := gm.Repo.GetCharacter(pseudo)
	if err != nil {
		return nil, err
	}
	var stats domain.Stats
	json.Unmarshal([]byte(statsStr), &stats)
	var inventory []domain.Item
	json.Unmarshal([]byte(invStr), &inventory)
	var equip domain.Equipment
	json.Unmarshal([]byte(equipStr), &equip)
	var quests []domain.Quest
	json.Unmarshal([]byte(questsStr), &quests)

	char := &domain.Character{
		ID:         uuid.New(),
		Alignement: "Neutre",
		Stats:      stats,
		CurrentPV:  float64(pv),
		Or:         or,
		Xp:         xp,
		Lieu:       lieu,
		Inventaire: inventory,
		Equipement: equip,
		Quests:     quests,
	}
	if char.CurrentPV <= 0 {
		char.CurrentPV = stats.CalculateLifePoints()
	}
	return char, nil
}

func (gm *GameManager) Connect(pseudo string, c *client, charInfo map[string]string) {
	// Database restore happens outside the lock (see restoreCharacter).
	restored, restoreErr := gm.restoreCharacter(pseudo)

	gm.mu.Lock()
	defer gm.mu.Unlock()

	if old, ok := gm.Connections[pseudo]; ok && old != nil {
		old.close()
		gm.chat("⚠️ %s vient de se reconnecter, l'ancienne session est fermée.", pseudo)
	}
	gm.Connections[pseudo] = c

	// Detect DM role
	isDM := pseudo == "dm" || len(pseudo) > 3 && pseudo[:3] == "dm_"
	gm.DMs[pseudo] = isDM

	player, ok := gm.World.Players[pseudo]
	if !ok {
		player = &domain.Player{Pseudo: pseudo}
		gm.World.Players[pseudo] = player
	}

	if restoreErr == nil {
		player.Characters = append(player.Characters, restored)
		gm.chat("👋 %s est revenu dans le monde !", pseudo)
	} else {
		// New character creation: build a temporary adventurer, then let
		// the client run the onboarding wizard to finalize class and stats.
		char := buildCharacter(charInfo["classe"], charInfo["nom_personnage"])
		player.Characters = append(player.Characters, char)
		gm.saveCharacterState(pseudo, char)
		gm.sendTo(pseudo, map[string]interface{}{"type": "init_new_char"})
		gm.chat("🆕 %s arrive en quête d'un destin...", pseudo)
	}

	// Location data is otherwise only broadcast when it changes, so the
	// joiner receives the full list right away. The coalesced sync
	// (moi + roster) follows on the next flush.
	gm.sendTo(pseudo, LocationsMessage{Type: "locations", Locations: gm.World.Locations})
	gm.NotifyChange()
}

func (gm *GameManager) HandleAction(pseudo string, action Action) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	sanitizeAction(&action)
	action.Pseudo = pseudo

	// DM actions are routed to the DM handler registry.
	if gm.DMs[pseudo] {
		if handler, ok := gm.dmActions[action.Type]; ok {
			handler(gm, action)
		}
		return
	}

	char := gm.getLatestCharacter(pseudo)
	if char == nil {
		return
	}
	// Only persist and broadcast when the action actually mutated state:
	// chats, refused moves or failed attacks then cost nothing.
	if handler, ok := gm.playerActions[action.Type]; ok && handler(gm, char, action) {
		gm.saveCharacterState(pseudo, char)
		gm.NotifyChange()
	}
}

// saveCharacterState snapshots the character and queues an async database
// write. Everything the writer goroutine reads is copied here — it must not
// touch the character, which the game loop keeps mutating concurrently.
// Writes are keyed per pseudo, so bursts collapse into the latest snapshot.
func (gm *GameManager) saveCharacterState(pseudo string, char *domain.Character) {
	statsJSON, _ := json.Marshal(char.Stats)
	invJSON, _ := json.Marshal(char.Inventaire)
	equipJSON, _ := json.Marshal(char.Equipement)
	questsJSON, _ := json.Marshal(char.Quests)
	nom := char.Stats.Nom
	classe := char.Stats.Background
	lieu := char.Lieu
	pv := int(char.CurrentPV)
	maxPV := int(char.Stats.CalculateLifePoints())
	or, xp := char.Or, char.Xp
	gm.persister.Enqueue("char:"+pseudo, func() {
		gm.Repo.SaveCharacter(pseudo, nom, classe, lieu, pv, maxPV, string(invJSON), string(statsJSON), string(equipJSON), string(questsJSON), or, xp)
	})
}

func (gm *GameManager) getLatestCharacter(pseudo string) *domain.Character {
	player, ok := gm.World.Players[pseudo]
	if !ok || len(player.Characters) == 0 {
		return nil
	}
	return player.Characters[len(player.Characters)-1]
}

// getQuestTarget resolves the character who owns quests, either a player character or a world NPC.
func (gm *GameManager) getQuestTarget(pseudo string) (*domain.Character, bool) {
	if char := gm.getLatestCharacter(pseudo); char != nil {
		return char, false
	}
	if npc, ok := gm.World.NPCs[pseudo]; ok {
		return npc, true
	}
	return nil, false
}

// NotifyChange marks the game state as needing a client sync. The broadcast
// is coalesced: the sync loop flushes at most once per syncInterval, so a
// burst of actions produces a single payload instead of one per action.
// Must be called with gm.mu held.
func (gm *GameManager) NotifyChange() {
	gm.syncDirty = true
}

// markLocationsDirty flags location data (ground loot, shops, quests) as
// changed so the next flush re-broadcasts it; unchanged locations are not
// repeated in every sync. Must be called with gm.mu held.
func (gm *GameManager) markLocationsDirty() {
	gm.locsDirty = true
}

// flushSync rebuilds the pending broadcasts. State is snapshotted under the
// lock and serialised outside of it, so running actions never wait on JSON
// encoding. Player clients receive a slim roster plus their own full
// character (separate messages so one roster payload serves every player);
// the DM panel receives the complete state; locations travel separately and
// only when they changed.
func (gm *GameManager) flushSync() {
	gm.mu.Lock()
	if !gm.syncDirty && !gm.locsDirty {
		gm.mu.Unlock()
		return
	}
	syncDirty, locsDirty := gm.syncDirty, gm.locsDirty
	gm.syncDirty = false
	gm.locsDirty = false

	type recipient struct {
		pseudo string
		client *client
	}
	var playerRecipients []recipient
	var dmConns []*client
	for pseudo, c := range gm.Connections {
		if gm.DMs[pseudo] {
			dmConns = append(dmConns, c)
		} else {
			playerRecipients = append(playerRecipients, recipient{pseudo: pseudo, client: c})
		}
	}

	var roster map[string]PlayerSummary
	var npcSlim map[string]NPCSummary
	own := make(map[string]PlayerEntry, len(playerRecipients))
	var dmSync *SyncMessage
	var locsMsg *LocationsMessage

	if syncDirty {
		if len(playerRecipients) > 0 {
			roster = gm.buildRoster()
			npcSlim = gm.buildNPCSummaries()
			for _, r := range playerRecipients {
				own[r.pseudo] = gm.buildOwnEntry(r.pseudo)
			}
		}
		if len(dmConns) > 0 {
			dmSync = gm.buildDMSync()
		}
	}
	if locsDirty && (len(playerRecipients) > 0 || len(dmConns) > 0) {
		locsMsg = &LocationsMessage{Type: "locations", Locations: cloneLocations(gm.World.Locations)}
	}
	gm.mu.Unlock()

	// --- serialisation happens with the lock released ---

	// Locations first: clients recompute their current-location view from
	// them and a change may land in the same flush as a sync.
	if locsMsg != nil {
		if msg, err := json.Marshal(locsMsg); err == nil {
			for _, r := range playerRecipients {
				r.client.enqueue(msg)
			}
			for _, c := range dmConns {
				c.enqueue(msg)
			}
		}
	}

	if syncDirty {
		// Own full character first, so the roster sync that follows cannot
		// overwrite it with the slim summary (the client keeps its latest
		// full entry either way).
		for _, r := range playerRecipients {
			me, ok := own[r.pseudo]
			if !ok {
				continue
			}
			if msg, err := json.Marshal(&MeMessage{Type: "moi", Moi: me}); err == nil {
				r.client.enqueue(msg)
			}
		}
		if roster != nil {
			if msg, err := json.Marshal(&PlayerSyncMessage{Type: "sync", Liste: roster, NPCs: npcSlim}); err == nil {
				for _, r := range playerRecipients {
					r.client.enqueue(msg)
				}
			}
		}
		if dmSync != nil {
			if msg, err := json.Marshal(dmSync); err == nil {
				for _, c := range dmConns {
					c.enqueue(msg)
				}
			}
		}
	}
}

// buildRoster returns the slim summary of every character. Caller holds gm.mu.
func (gm *GameManager) buildRoster() map[string]PlayerSummary {
	roster := make(map[string]PlayerSummary, len(gm.World.Players))
	for pseudo, player := range gm.World.Players {
		if len(player.Characters) == 0 {
			continue
		}
		char := player.Characters[len(player.Characters)-1]
		roster[pseudo] = PlayerSummary{
			Nom:        char.Stats.Nom,
			PV:         char.CurrentPV,
			MaxPV:      char.Stats.CalculateLifePoints(),
			Classe:     char.Stats.Background,
			Lieu:       char.Lieu,
			Alignement: char.Alignement,
			Niveau:     char.Level(),
			Role:       gm.DMs[pseudo],
		}
	}
	return roster
}

// buildNPCSummaries returns the slim NPC/MOB view sent to player clients.
// Caller holds gm.mu.
func (gm *GameManager) buildNPCSummaries() map[string]NPCSummary {
	summaries := make(map[string]NPCSummary, len(gm.World.NPCs))
	for name, npc := range gm.World.NPCs {
		summaries[name] = NPCSummary{
			Nom:        npc.Stats.Nom,
			PV:         npc.CurrentPV,
			MaxPV:      npcMaxPV(npc),
			Classe:     npc.Stats.Background,
			Lieu:       npc.Lieu,
			Alignement: npc.Alignement,
			MobType:    npc.MobType,
		}
	}
	return summaries
}

// buildOwnEntry returns the full sync entry of one character with private
// copies of every slice, so it can be marshalled after the lock is released.
// Caller holds gm.mu.
func (gm *GameManager) buildOwnEntry(pseudo string) PlayerEntry {
	player, ok := gm.World.Players[pseudo]
	if !ok || len(player.Characters) == 0 {
		return PlayerEntry{}
	}
	return playerEntryFor(player.Characters[len(player.Characters)-1], gm.DMs[pseudo])
}

// buildDMSync returns the complete state for the DM panel. Caller holds gm.mu.
func (gm *GameManager) buildDMSync() *SyncMessage {
	liste := make(map[string]PlayerEntry, len(gm.World.Players))
	for pseudo, player := range gm.World.Players {
		if len(player.Characters) == 0 {
			continue
		}
		liste[pseudo] = playerEntryFor(player.Characters[len(player.Characters)-1], gm.DMs[pseudo])
	}
	npcs := make(map[string]NPCEntry, len(gm.World.NPCs))
	for name, npc := range gm.World.NPCs {
		stats := npc.Stats
		equip := npc.Equipement
		npcs[name] = NPCEntry{
			Nom:        stats.Nom,
			PV:         npc.CurrentPV,
			MaxPV:      npcMaxPV(npc),
			Classe:     stats.Background,
			Lieu:       npc.Lieu,
			Alignement: npc.Alignement,
			Sorts:      cloneSorts(npc.Sorts),
			Inventaire: cloneItems(npc.Inventaire),
			Equipement: equip,
			Quests:     cloneQuests(npc.Quests),
			Stats:      stats,
			Combat:     computeDerivedCombat(npc),
			MobType:    npc.MobType,
			IsNPC:      true,
		}
	}
	return &SyncMessage{Type: "sync", Liste: liste, NPCs: npcs}
}

// playerEntryFor builds the full PlayerEntry, deep-copying every slice that
// will be marshalled outside the lock. Caller holds gm.mu.
func playerEntryFor(char *domain.Character, isDM bool) PlayerEntry {
	stats := char.Stats
	equip := char.Equipement
	return PlayerEntry{
		Nom:          stats.Nom,
		PV:           char.CurrentPV,
		MaxPV:        stats.CalculateLifePoints(),
		Classe:       stats.Background,
		Lieu:         char.Lieu,
		Alignement:   char.Alignement,
		Sorts:        cloneSorts(char.Sorts),
		Inventaire:   cloneItems(char.Inventaire),
		Equipement:   equip,
		Quests:       cloneQuests(char.Quests),
		Stats:        stats,
		Combat:       computeDerivedCombat(char),
		Role:         isDM,
		Or:           char.Or,
		Xp:           char.Xp,
		Niveau:       char.Level(),
		Encombrement: char.CarriedWeight(),
		Capacite:     char.Capacity(),
	}
}

func (gm *GameManager) Disconnect(pseudo string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	if c, ok := gm.Connections[pseudo]; ok {
		c.close()
	}
	delete(gm.Connections, pseudo)
	delete(gm.DMs, pseudo)
	gm.chat("🏃 %s a quitté le jeu.", pseudo)
	gm.NotifyChange()
}
