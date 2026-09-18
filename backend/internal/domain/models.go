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
	Or         float64   `json:"or,omitempty"`
	Xp         float64   `json:"xp,omitempty"`
	Lieu       string    `json:"lieu"`
	MobType    string    `json:"mobType,omitempty"`
}

// Capacity is the maximum encumbrance value a character can carry.
func (c *Character) Capacity() float64 {
	if c.Stats.Force <= 0 {
		return 150
	}
	return c.Stats.Force * 15
}

// ItemWeight returns the effective weight of an item, falling back to a
// sensible default when the DM never set an explicit encumbrance value.
func ItemWeight(it Item) float64 {
	if it.Encombrement > 0 {
		return it.Encombrement
	}
	if it.IsConsumable {
		return 0.5
	}
	if it.BonusArmure > 0 {
		return 8
	}
	if it.BonusDégâts > 0 || it.DesDégâts != "" {
		return 4
	}
	return 1
}

// CarriedWeight totals the encumbrance of equipped and carried items.
func (c *Character) CarriedWeight() float64 {
	var total float64
	for _, it := range c.Inventaire {
		total += ItemWeight(it)
	}
	if c.Equipement.Arme != nil {
		total += ItemWeight(*c.Equipement.Arme)
	}
	if c.Equipement.Armure != nil {
		total += ItemWeight(*c.Equipement.Armure)
	}
	return total
}

// Level derives the character level from cumulative XP (100 XP per level).
func (c *Character) Level() int {
	return 1 + int(c.Xp)/100
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
	Commerce   []Item  `json:"commerce,omitempty"`
	Links      []string `json:"liens,omitempty"`
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
