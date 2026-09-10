package services

import (
	"encoding/json"
	"sync"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

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
	Pseudo       string       `json:"-"`
}

type GameManager struct {
	Connections   map[string]*websocket.Conn
	World         *domain.World
	Repo          *repository.SQLiteRepository
	DMs           map[string]bool
	persister     *Persister
	playerActions map[string]playerActionFn
	dmActions     map[string]dmActionFn
	mu            sync.Mutex
}

func NewGameManager(repo *repository.SQLiteRepository) *GameManager {
	gm := &GameManager{
		Connections: make(map[string]*websocket.Conn),
		DMs:         make(map[string]bool),
		Repo:        repo,
		persister:   NewPersister(repo),
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

// Close flushes pending database writes.
func (gm *GameManager) Close() {
	gm.persister.Close()
}

// DisconnectAll drops every active connection (used during shutdown).
func (gm *GameManager) DisconnectAll() {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	for pseudo := range gm.Connections {
		delete(gm.Connections, pseudo)
	}
}

func (gm *GameManager) Connect(pseudo string, ws *websocket.Conn, charInfo map[string]string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.Connections[pseudo] = ws

	// Detect DM role
	isDM := pseudo == "dm" || len(pseudo) > 3 && pseudo[:3] == "dm_"
	gm.DMs[pseudo] = isDM

	player, ok := gm.World.Players[pseudo]
	if !ok {
		player = &domain.Player{Pseudo: pseudo}
		gm.World.Players[pseudo] = player
	}

	// 1. Attempt to restore from DB
	_, _, lieu, pv, _, invStr, statsStr, equipStr, questsStr, err := gm.Repo.GetCharacter(pseudo)
	if err == nil {
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
			Lieu:       lieu,
			Inventaire: inventory,
			Equipement: equip,
			Quests:     quests,
		}
		player.Characters = append(player.Characters, char)
		gm.chat("👋 %s est revenu dans le monde !", pseudo)
	} else {
		// 2. New character creation: build a temporary adventurer, then let
		// the client run the onboarding wizard to finalize class and stats.
		char := buildCharacter(charInfo["classe"], charInfo["nom_personnage"])
		player.Characters = append(player.Characters, char)
		gm.saveCharacterState(pseudo, char)
		gm.sendTo(pseudo, map[string]interface{}{"type": "init_new_char"})
		gm.chat("🆕 %s arrive en quête d'un destin...", pseudo)
	}
	gm.NotifyChange()
}

func (gm *GameManager) HandleAction(pseudo string, action Action) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

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
	if handler, ok := gm.playerActions[action.Type]; ok {
		handler(gm, char, action)
	}
	gm.saveCharacterState(pseudo, char)
	gm.NotifyChange()
}

// saveCharacterState snapshots the character and queues an async database write.
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
	gm.persister.Enqueue(func() {
		gm.Repo.SaveCharacter(pseudo, nom, classe, lieu, pv, maxPV, string(invJSON), string(statsJSON), string(equipJSON), string(questsJSON))
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

func (gm *GameManager) NotifyChange() {
	syncData := make(map[string]PlayerEntry)
	for pseudo, player := range gm.World.Players {
		if len(player.Characters) > 0 {
			char := player.Characters[len(player.Characters)-1]
			syncData[pseudo] = PlayerEntry{
				Nom:        char.Stats.Nom,
				PV:         char.CurrentPV,
				MaxPV:      char.Stats.CalculateLifePoints(),
				Classe:     char.Stats.Background,
				Lieu:       char.Lieu,
				Alignement: char.Alignement,
				Sorts:      char.Sorts,
				Inventaire: char.Inventaire,
				Equipement: char.Equipement,
				Quests:     char.Quests,
				Stats:      char.Stats,
				Role:       gm.DMs[pseudo],
			}
		}
	}
	npcData := make(map[string]NPCEntry)
	for name, npc := range gm.World.NPCs {
		npcData[name] = NPCEntry{
			Nom:        npc.Stats.Nom,
			PV:         npc.CurrentPV,
			MaxPV:      npc.Stats.CalculateLifePoints(),
			Classe:     npc.Stats.Background,
			Lieu:       npc.Lieu,
			Alignement: npc.Alignement,
			Sorts:      npc.Sorts,
			Inventaire: npc.Inventaire,
			Equipement: npc.Equipement,
			Quests:     npc.Quests,
			Stats:      npc.Stats,
			IsNPC:      true,
		}
	}
	gm.broadcast(SyncMessage{
		Type:      "sync",
		Liste:     syncData,
		NPCs:      npcData,
		Locations: gm.World.Locations,
	})
}

func (gm *GameManager) Disconnect(pseudo string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.Connections, pseudo)
	delete(gm.DMs, pseudo)
	gm.chat("🏃 %s a quitté le jeu.", pseudo)
	gm.NotifyChange()
}
