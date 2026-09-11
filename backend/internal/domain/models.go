package domain

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/google/uuid"
)

type Stats struct {
	Nom          string  `json:"nom"`
	Background   string  `json:"background"`
	Force        float64 `json:"force"`
	Constitution float64 `json:"constitution"`
	Vitesse      float64 `json:"vitesse"`
	Charisme     float64 `json:"charisme"`
	Savoir       float64 `json:"savoir"`
	Instinct     float64 `json:"instinct"`
}

func (s *Stats) HitDiceSides() int {
	switch s.Background {
	case "Magicien":
		return 6
	case "Voleur", "Clerc", "Barde", "Ranger":
		return 8
	default: // Guerrier
		return 10
	}
}

func (s *Stats) CalculateLifePoints() float64 {
	hd := float64(s.HitDiceSides())
	conMod := AbilityModifier(s.Constitution)
	maxRoll := hd
 pv := maxRoll + conMod
	if pv < 1 {
		pv = 1
	}
	return pv
}

func AbilityModifier(stat float64) float64 {
	return math.Floor((stat - 10.0) / 2.0)
}

func (s *Stats) BaseAC() float64 {
	return 10.0 + AbilityModifier(s.Vitesse)
}

type Sort struct {
	Nom        string    `json:"nom"`
	NiveauSort float64   `json:"niveauSort"`
	EcoleMagie string    `json:"ecoleMagie"`
	Portee     string    `json:"portee"`
	Duree      string    `json:"duree"`
	Bonus      float64   `json:"bonus,omitempty"`
	DesDégâts  string    `json:"desDegats,omitempty"`
	Buff       *SortBuff `json:"buff,omitempty"`
}

type SortBuff struct {
	Stat   string  `json:"stat"`
	Valeur float64 `json:"valeur"`
}

type Quest struct {
	Nom         string `json:"nom"`
	Objectif    string `json:"objectif"`
	Obstacle    string `json:"obstacle,omitempty"`
	Recompense  []Item `json:"recompense,omitempty"`
	Information string `json:"information,omitempty"`
}

func (q *Quest) UnmarshalJSON(data []byte) error {
	type alias Quest
	var plain alias
	if err := json.Unmarshal(data, &plain); err == nil {
		*q = Quest(plain)
		return nil
	}
	var old struct {
		Nom        string `json:"nom"`
		Objectif   string `json:"objectif"`
		Obstacle   []Item `json:"obstacle"`
		Recompense []Item `json:"recompense"`
	}
	if err2 := json.Unmarshal(data, &old); err2 != nil {
		return err2
	}
	desc := make([]string, 0, len(old.Obstacle))
	for _, it := range old.Obstacle {
		desc = append(desc, it.Nom)
	}
	*q = Quest{Nom: old.Nom, Objectif: old.Objectif, Obstacle: strings.Join(desc, ", "), Recompense: old.Recompense}
	return nil
}

type Item struct {
	Nom          string  `json:"nom"`
	Prix         float64 `json:"prix"`
	Encombrement float64 `json:"encombrement"`
	IsConsumable bool    `json:"isConsumable"`
	BonusDégâts  float64 `json:"bonusDegats"`
	BonusArmure  float64 `json:"bonusArmure"`
	DesDégâts    string  `json:"desDegats,omitempty"`
}

type Equipment struct {
	Arme   *Item `json:"arme,omitempty"`
	Armure *Item `json:"armure,omitempty"`
}

type Character struct {
	ID         uuid.UUID `json:"id"`
	Alignement string    `json:"alignement"`
	Stats      Stats     `json:"stats"`
	CurrentPV  float64   `json:"pv"`
	MaxPV      float64   `json:"max_pv,omitempty"`
	Inventaire []Item    `json:"inventaire"`
	Sorts      []Sort    `json:"sorts"`
	Equipement Equipment `json:"equipement"`
	Quests     []Quest   `json:"quests"`
	Lieu       string    `json:"lieu"`
	MobType    string    `json:"mobType,omitempty"`
}

type Player struct {
	Pseudo     string       `json:"pseudo"`
	Characters []*Character `json:"characters"`
}

type Location struct {
	Nom        string  `json:"nom"`
	Background string  `json:"background"`
	Objects    []Item  `json:"objects"`
	Quests     []Quest `json:"quests,omitempty"`
	Position   struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
}

type World struct {
	ID        uuid.UUID             `json:"id"`
	Players   map[string]*Player    `json:"players"`
	NPCs      map[string]*Character `json:"npcs"`
	Locations []Location            `json:"locations"`
}
