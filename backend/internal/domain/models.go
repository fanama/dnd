package domain

import (
	"github.com/google/uuid"
)

type Stats struct {
	Nom          string  `json:"nom"`
	Background   string  `json:"background"`
	Force        float64 `json:"force"`
	Constitution float64 `json:"constitution"`
	Vitesse      float64 `json:"vitesse"`
	Charisme     float64 `json:"charisme"`
	Savoir      float64 `json:"savoir"`
	Instinct     float64 `json:"instinct"`
}

func (s *Stats) CalculateLifePoints() float64 {
	return s.Constitution * 10.0
}

func (s *Stats) CalculateArmor() float64 {
	return s.Vitesse * 1.5
}

func (s *Stats) CalculateDamage() float64 {
	return s.Force * 2.0
}

type Sort struct {
	Nom         string  `json:"nom"`
	NiveauSort  float64 `json:"niveauSort"`
	EcoleMagie  string  `json:"ecoleMagie"`
	Portee      string  `json:"portee"`
	Duree       string  `json:"duree"`
}

type Item struct {
	Nom          string  `json:"nom"`
	Prix         float64 `json:"prix"`
	Encombrement float64 `json:"encombrement"`
	IsConsumable bool    `json:"isConsumable"`
	BonusDégâts  float64 `json:"bonusDegats"`
	BonusArmure  float64 `json:"bonusArmure"`
}

type Character struct {
	ID           uuid.UUID `json:"id"`
	Alignement   string    `json:"alignement"`
	Stats        Stats     `json:"stats"`
	CurrentPV    float64   `json:"pv"`
	Inventaire   []Item    `json:"inventaire"`
	Sorts        []Sort    `json:"sorts"`
	Lieu         string    `json:"lieu"`
}

type Player struct {
	Pseudo     string     `json:"pseudo"`
	Characters []*Character `json:"characters"`
}

type Location struct {
	Nom        string `json:"nom"`
	Background string `json:"background"`
	Objects    []Item `json:"objects"`
	Position   struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
}

type World struct {
	ID        uuid.UUID   `json:"id"`
	Players   map[string]*Player `json:"players"`
	Locations []Location  `json:"locations"`
}
