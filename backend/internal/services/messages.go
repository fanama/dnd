package services

import (
	"encoding/json"
	"fmt"

	"dnd-backend/internal/domain"
	"github.com/gorilla/websocket"
)

// ChatMessage is the typed payload for human-readable game events.
type ChatMessage struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// PlayerEntry is the typed representation of a player character in a sync.
type PlayerEntry struct {
	Nom        string           `json:"nom"`
	PV         float64          `json:"pv"`
	MaxPV      float64          `json:"max_pv"`
	Classe     string           `json:"classe"`
	Lieu       string           `json:"lieu"`
	Alignement string           `json:"alignement"`
	Sorts      []domain.Sort    `json:"sorts"`
	Inventaire []domain.Item    `json:"inventaire"`
	Equipement domain.Equipment `json:"equipement"`
	Quests     []domain.Quest   `json:"quests"`
	Stats      domain.Stats     `json:"stats"`
	Role       bool             `json:"role"`
}

// NPCEntry is the typed representation of an NPC in a sync.
type NPCEntry struct {
	Nom        string           `json:"nom"`
	PV         float64          `json:"pv"`
	MaxPV      float64          `json:"max_pv"`
	Classe     string           `json:"classe"`
	Lieu       string           `json:"lieu"`
	Alignement string           `json:"alignement"`
	Sorts      []domain.Sort    `json:"sorts"`
	Inventaire []domain.Item    `json:"inventaire"`
	Equipement domain.Equipment `json:"equipement"`
	Quests     []domain.Quest   `json:"quests"`
	Stats      domain.Stats     `json:"stats"`
	IsNPC      bool             `json:"is_npc"`
}

// SyncMessage is the full game state broadcast after every mutation.
type SyncMessage struct {
	Type      string                 `json:"type"`
	Liste     map[string]PlayerEntry `json:"liste"`
	NPCs      map[string]NPCEntry    `json:"npcs"`
	Locations []domain.Location      `json:"locations"`
}

func (gm *GameManager) broadcast(data interface{}) {
	msg, _ := json.Marshal(data)
	for _, conn := range gm.Connections {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (gm *GameManager) chat(format string, args ...interface{}) {
	gm.broadcast(ChatMessage{Type: "chat", Msg: fmt.Sprintf(format, args...)})
}
