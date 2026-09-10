package services

import (
	"dnd-backend/internal/domain"
	"github.com/google/uuid"
)

// defaultStatsForClass returns the base attribute values for a class.
// The Expression/Background and Name are set by the caller.
func defaultStatsForClass(classe string) domain.Stats {
	s := domain.Stats{Background: classe}
	switch classe {
	case "Magicien":
		s.Force, s.Constitution, s.Vitesse = 8, 9, 11
		s.Charisme, s.Instinct, s.Savoir = 12, 14, 16
	case "Voleur":
		s.Force, s.Constitution, s.Vitesse = 10, 8, 16
		s.Charisme, s.Instinct, s.Savoir = 10, 15, 11
	case "Clerc":
		s.Force, s.Constitution, s.Vitesse = 12, 14, 8
		s.Charisme, s.Instinct, s.Savoir = 15, 9, 12
	case "Barde":
		s.Force, s.Constitution, s.Vitesse = 9, 10, 13
		s.Charisme, s.Instinct, s.Savoir = 16, 12, 10
	case "Ranger":
		s.Force, s.Constitution, s.Vitesse = 13, 11, 14
		s.Charisme, s.Instinct, s.Savoir = 8, 15, 9
	default: // Guerrier
		s.Force, s.Constitution, s.Vitesse = 15, 12, 10
		s.Charisme, s.Instinct, s.Savoir = 10, 10, 10
	}
	return s
}

// defaultStartingSpells returns the spells granted to a starting class.
func defaultStartingSpells(classe string) []domain.Sort {
	switch classe {
	case "Magicien":
		return []domain.Sort{{Nom: "Boule de Feu", EcoleMagie: "Évocations"}}
	case "Clerc":
		return []domain.Sort{{Nom: "Soin Divin", EcoleMagie: "Guérison"}}
	case "Barde":
		return []domain.Sort{{Nom: "Mélodie Envoûtante", EcoleMagie: "Enchantement"}}
	default:
		return []domain.Sort{}
	}
}

// defaultStartingItems returns the equipment granted at first level.
func defaultStartingItems(classe string) []domain.Item {
	items := []domain.Item{}
	switch classe {
	case "Guerrier":
		items = append(items, domain.Item{Nom: "Épée Longue", IsConsumable: false, BonusDégâts: 5, Prix: 100})
	case "Magicien":
		items = append(items, domain.Item{Nom: "Bâton Mystique", IsConsumable: false, BonusDégâts: 3, Prix: 80})
	case "Voleur":
		items = append(items, domain.Item{Nom: "Dague Empoisonnée", IsConsumable: false, BonusDégâts: 4, Prix: 90})
	case "Clerc":
		items = append(items, domain.Item{Nom: "Marteau Sacré", IsConsumable: false, BonusDégâts: 4, Prix: 95})
	case "Barde":
		items = append(items, domain.Item{Nom: "Luth Enchanté", IsConsumable: false, BonusDégâts: 2, Prix: 70})
	case "Ranger":
		items = append(items, domain.Item{Nom: "Arc Long", IsConsumable: false, BonusDégâts: 5, Prix: 100})
	}
	items = append(items, domain.Item{Nom: "Potion de Soin", IsConsumable: true, Prix: 25})
	return items
}

// buildCharacter assembles a brand-new level-1 character for the given class.
func buildCharacter(classe, name string) *domain.Character {
	if classe == "" {
		classe = "Guerrier"
	}
	stats := defaultStatsForClass(classe)
	stats.Nom = name

	char := &domain.Character{
		ID:         uuid.New(),
		Alignement: "Neutre",
		Stats:      stats,
		CurrentPV:  stats.CalculateLifePoints(),
		Lieu:       "Taverne",
		Sorts:      defaultStartingSpells(classe),
		Inventaire: defaultStartingItems(classe),
	}
	for _, it := range char.Inventaire {
		if !it.IsConsumable && it.BonusDégâts > 0 {
			equipped := it
			char.Equipement.Arme = &equipped
			break
		}
	}
	return char
}

// applyCustomCreation rebuilds the character from the wizard choices
// (class, allocated stats, name) shared by the onboarding flow.
func (gm *GameManager) applyCustomCreation(char *domain.Character, action Action) {
	classe := action.Stats.Background
	if classe == "" {
		classe = "Guerrier"
	}

	stats := action.Stats
	if stats.Nom == "" {
		stats.Nom = char.Stats.Nom
	}
	clamp := func(v *float64) {
		if *v < 3 {
			*v = 3
		}
		if *v > 20 {
			*v = 20
		}
	}
	clamp(&stats.Force)
	clamp(&stats.Constitution)
	clamp(&stats.Vitesse)
	clamp(&stats.Charisme)
	clamp(&stats.Instinct)
	clamp(&stats.Savoir)
	stats.Background = classe

	char.Stats = stats
	char.Sorts = defaultStartingSpells(classe)
	char.Inventaire = defaultStartingItems(classe)
	char.Equipement.Arme = nil
	for _, it := range char.Inventaire {
		if !it.IsConsumable && it.BonusDégâts > 0 {
			equipped := it
			char.Equipement.Arme = &equipped
			break
		}
	}
	char.CurrentPV = stats.CalculateLifePoints()
}

// createCharacter handles the final step of the onboarding wizard.
func (gm *GameManager) createCharacter(char *domain.Character, action Action) {
	gm.applyCustomCreation(char, action)
	gm.chat("⚔️ %s a incarné %s (%s) !", action.Pseudo, char.Stats.Nom, char.Stats.Background)
}