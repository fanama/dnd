package services

import (
	"encoding/json"
	"time"

	"dnd-backend/internal/domain"
	"github.com/gorilla/websocket"
)

// StateSnapshot is a full, portable dump of the game state (players, NPCs and
// locations) used by the DM to download and restore a campaign.
type StateSnapshot struct {
	Version    int                         `json:"version"`
	ExportedAt string                      `json:"exported_at"`
	Players    map[string]*domain.Player   `json:"players"`
	NPCs       map[string]*domain.Character `json:"npcs"`
	Locations  []domain.Location           `json:"locations"`
}

// sendTo delivers a message to a single connected client.
func (gm *GameManager) sendTo(pseudo string, data interface{}) {
	conn := gm.Connections[pseudo]
	if conn == nil {
		return
	}
	msg, _ := json.Marshal(data)
	conn.WriteMessage(websocket.TextMessage, msg)
}

// snapshotState captures the current world as a portable snapshot.
func (gm *GameManager) snapshotState() *StateSnapshot {
	players := make(map[string]*domain.Player, len(gm.World.Players))
	for pseudo, pl := range gm.World.Players {
		players[pseudo] = pl
	}
	npcs := make(map[string]*domain.Character, len(gm.World.NPCs))
	for name, npc := range gm.World.NPCs {
		npcs[name] = npc
	}
	return &StateSnapshot{
		Version:    1,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Players:    players,
		NPCs:       npcs,
		Locations:  gm.World.Locations,
	}
}

// dmExportState streams the full game state back to the requesting DM.
func (gm *GameManager) dmExportState(pseudo string) {
	gm.sendTo(pseudo, map[string]interface{}{
		"type":    "state_export",
		"payload": gm.snapshotState(),
	})
	gm.chat("📦 Le MDJ a téléchargé l'état du jeu.")
}

// dmLoadState replaces the current game state with a previously exported one.
func (gm *GameManager) dmLoadState(payload string) {
	var snap StateSnapshot
	if err := json.Unmarshal([]byte(payload), &snap); err != nil {
		gm.chat("❌ État de partie invalide : %v", err)
		return
	}
	if snap.Version != 1 || snap.Players == nil && snap.NPCs == nil && snap.Locations == nil {
		gm.chat("❌ État de partie invalide ou incompatible (version %d).", snap.Version)
		return
	}

	gm.World.Players = make(map[string]*domain.Player)
	for pseudo, pl := range snap.Players {
		if pl != nil {
			gm.World.Players[pseudo] = pl
		}
	}
	gm.World.NPCs = make(map[string]*domain.Character)
	for name, npc := range snap.NPCs {
		if npc != nil {
			gm.World.NPCs[name] = npc
		}
	}
	gm.World.Locations = snap.Locations

	gm.persistWorld()
	for pseudo, pl := range gm.World.Players {
		if len(pl.Characters) > 0 {
			gm.saveCharacterState(pseudo, pl.Characters[len(pl.Characters)-1])
		}
	}
	gm.chat("💾 Le MDJ a restauré un état de partie sauvegardé.")
	gm.NotifyChange()
}