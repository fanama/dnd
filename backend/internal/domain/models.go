package domain

import (
	"math"

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

func (s *Stats) CalculateLifePoints() float64 {
	return s.Constitution * 10.0
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
	Nom        string `json:"nom"`
	Objectif   string `json:"objectif"`
	Obstacle   []Item `json:"obstacle,omitempty"`
	Recompense []Item `json:"recompense,omitempty"`
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
	Inventaire []Item    `json:"inventaire"`
	Sorts      []Sort    `json:"sorts"`
	Equipement Equipment `json:"equipement"`
	Quests     []Quest   `json:"quests"`
	Lieu       string    `json:"lieu"`
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
