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

// DerivedCombat holds the combat statistics precomputed by the server so that
// clients never re-implement the game rules (single source of truth).
type DerivedCombat struct {
	AC         float64 `json:"ac"`
	AttackMod  float64 `json:"attack_mod"`
	DamageMod  float64 `json:"damage_mod"`
	DamageDice string  `json:"damage_dice"`
	HitDice    int     `json:"hit_dice"`
	WeaponName string  `json:"weapon_name"`
	Ranged     bool    `json:"ranged"`
}

// computeDerivedCombat derives the combat statistics of a character from its
// raw stats and equipment, mirroring the shake resolution rules exactly.
func computeDerivedCombat(char *domain.Character) DerivedCombat {
	weapon := char.Equipement.Arme
	sides, ranged := weaponInfo(weapon)
	weaponName := "Mains nues"
	var magicBonus float64
	if weapon != nil {
		weaponName = weapon.Nom
		magicBonus = weapon.BonusDégâts
	}
	attackStat := char.Stats.Force
	if ranged {
		attackStat = char.Stats.Vitesse
	}
	attackMod := domain.AbilityModifier(attackStat)
	var armorBonus float64
	if char.Equipement.Armure != nil {
		armorBonus = char.Equipement.Armure.BonusArmure
	}
	return DerivedCombat{
		AC:         char.Stats.BaseAC() + armorBonus,
		AttackMod:  attackMod + magicBonus,
		DamageMod:  attackMod,
		DamageDice: fmt.Sprintf("1d%d%s", sides, signed(attackMod)),
		HitDice:    char.Stats.HitDiceSides(),
		WeaponName: weaponName,
		Ranged:     ranged,
	}
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
	Combat     DerivedCombat    `json:"combat"`
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
	Combat     DerivedCombat    `json:"combat"`
	MobType    string           `json:"mobType,omitempty"`
	IsNPC      bool             `json:"is_npc"`
}

// npcMaxPV returns the normalized hit point ceiling of an NPC or MOB: the
// DM-set maximum when present, otherwise the maximum of the current pool and
// the D&D-derived one (graceful fallback for legacy NPCs).
func npcMaxPV(npc *domain.Character) float64 {
	if npc.MaxPV > 0 {
		return npc.MaxPV
	}
	derived := npc.Stats.CalculateLifePoints()
	if npc.CurrentPV > derived {
		return npc.CurrentPV
	}
	return derived
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
