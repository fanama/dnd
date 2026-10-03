package services

import (
	"encoding/json"
	"fmt"

	"dnd-backend/internal/domain"
)

// ChatMessage is the typed payload for human-readable game events.
type ChatMessage struct {
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// GameEvent is a structured combat/log event consumed by the frontends to
// drive toasts, floating damage numbers, flashes and the death overlay.
type GameEvent struct {
	Type     string  `json:"type"`
	Event    string  `json:"event"` // damage | heal | death | loot | quest | buff | xp | level
	Source   string  `json:"source,omitempty"`
	Target   string  `json:"target,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
	Crit     bool    `json:"crit,omitempty"`
	Level    int     `json:"level,omitempty"`
	LootCount int    `json:"lootCount,omitempty"`
	Text     string  `json:"text,omitempty"`
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
	Or         float64          `json:"or"`
	Xp         float64          `json:"xp"`
	Niveau     int              `json:"niveau"`
	Encombrement float64        `json:"encombrement"`
	Capacite     float64        `json:"capacite"`
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

// SyncMessage is the full-state broadcast the DM panel receives: complete
// roster and complete NPCs. Locations are not part of it — they are sent
// once on connect and re-broadcast only when they change (LocationsMessage).
type SyncMessage struct {
	Type  string                 `json:"type"`
	Liste map[string]PlayerEntry `json:"liste"`
	NPCs  map[string]NPCEntry    `json:"npcs"`
}

// PlayerSummary is the slim roster entry player clients receive about
// *other* heroes: exactly the fields the player UI displays (HP bar, zone
// list, target picking). Full character sheets are sent only to their owner.
type PlayerSummary struct {
	Nom        string  `json:"nom"`
	PV         float64 `json:"pv"`
	MaxPV      float64 `json:"max_pv"`
	Classe     string  `json:"classe"`
	Lieu       string  `json:"lieu"`
	Alignement string  `json:"alignement"`
	Niveau     int     `json:"niveau"`
	Role       bool    `json:"role"`
}

// NPCSummary is the slim NPC/MOB entry player clients receive: HP bar,
// badges and targeting are all the player UI needs; combat resolution stays
// server-side. The DM panel keeps receiving full NPCEntry values.
type NPCSummary struct {
	Nom        string  `json:"nom"`
	PV         float64 `json:"pv"`
	MaxPV      float64 `json:"max_pv"`
	Classe     string  `json:"classe"`
	Lieu       string  `json:"lieu"`
	Alignement string  `json:"alignement"`
	MobType    string  `json:"mobType,omitempty"`
}

// PlayerSyncMessage is the coalesced roster broadcast for player clients:
// slim entries for everyone. The receiver's own full character arrives in a
// separate "moi" message so one shared payload can serve every player.
type PlayerSyncMessage struct {
	Type  string                  `json:"type"`
	Liste map[string]PlayerSummary `json:"liste"`
	NPCs  map[string]NPCSummary   `json:"npcs"`
}

// MeMessage carries the receiving player's own full character, so their
// sheet stays complete while the roster stays slim.
type MeMessage struct {
	Type string      `json:"type"`
	Moi  PlayerEntry `json:"moi"`
}

// LocationsMessage carries the full location data (ground loot, shops,
// quests). It is sent once on connect and re-broadcast only when locations
// actually change, instead of being repeated in every sync.
type LocationsMessage struct {
	Type      string            `json:"type"`
	Locations []domain.Location `json:"locations"`
}

// --- deep-copy helpers ------------------------------------------------------
//
// Sync payloads are snapshotted under the game lock but marshalled outside of
// it, so every slice that crosses that boundary must be a private copy: the
// game loop keeps mutating the originals concurrently.

func cloneItems(src []domain.Item) []domain.Item {
	if src == nil {
		return nil
	}
	return append([]domain.Item(nil), src...)
}

func cloneSorts(src []domain.Sort) []domain.Sort {
	if src == nil {
		return nil
	}
	return append([]domain.Sort(nil), src...)
}

func cloneQuests(src []domain.Quest) []domain.Quest {
	if src == nil {
		return nil
	}
	out := make([]domain.Quest, len(src))
	for i, q := range src {
		out[i] = q
		out[i].Recompense = cloneItems(q.Recompense)
	}
	return out
}

func cloneLocations(src []domain.Location) []domain.Location {
	out := make([]domain.Location, len(src))
	for i, l := range src {
		out[i] = l
		out[i].Links = append([]string(nil), l.Links...)
		out[i].Objects = cloneItems(l.Objects)
		out[i].Commerce = cloneItems(l.Commerce)
		out[i].Quests = cloneQuests(l.Quests)
	}
	return out
}

func (gm *GameManager) broadcast(data interface{}) {
	msg, _ := json.Marshal(data)
	gm.broadcastBytes(msg)
}

// broadcastBytes delivers raw bytes to every connected client through their
// per-connection queue. Messages are dropped (never blocking) when a client's
// queue is full; the next full-state sync self-heals it.
func (gm *GameManager) broadcastBytes(msg []byte) {
	for _, c := range gm.Connections {
		c.enqueue(msg)
	}
}

func (gm *GameManager) emit(e GameEvent) {
	gm.broadcast(e)
}

func (gm *GameManager) chat(format string, args ...interface{}) {
	gm.broadcast(ChatMessage{Type: "chat", Msg: fmt.Sprintf(format, args...)})
}
