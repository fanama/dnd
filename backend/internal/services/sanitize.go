package services

import "dnd-backend/internal/domain"

// clean normalizes user-provided text (trim, strip control chars, cap length).
func clean(s string, max int) string {
	return domain.SanitizeString(s, max)
}

// sanitizeAction normalizes every free-text field of an incoming action before
// the handler touches it, so display names and keys stay bounded and clean.
func sanitizeAction(a *Action) {
	a.Cible = clean(a.Cible, 40)
	a.Destination = clean(a.Destination, 40)
	a.Sort = clean(a.Sort, 40)
	a.ItemName = clean(a.ItemName, 40)
	a.LootName = clean(a.LootName, 40)
	a.TargetPlayer = clean(a.TargetPlayer, 24)
	a.Alignement = clean(a.Alignement, 40)
	a.NewName = clean(a.NewName, 40)
	a.LocationBg = clean(a.LocationBg, 200)
	a.QuestName = clean(a.QuestName, 40)
	a.Slot = clean(a.Slot, 12)
	a.MobType = clean(a.MobType, 12)
	a.Message = clean(a.Message, 400)

	a.Item.Nom = clean(a.Item.Nom, 40)
	a.Item.DesDégâts = clean(a.Item.DesDégâts, 16)

	a.Spell.Nom = clean(a.Spell.Nom, 40)
	a.Spell.EcoleMagie = clean(a.Spell.EcoleMagie, 40)
	a.Spell.Portee = clean(a.Spell.Portee, 40)
	a.Spell.Duree = clean(a.Spell.Duree, 40)
	a.Spell.DesDégâts = clean(a.Spell.DesDégâts, 16)

	a.Quest.Nom = clean(a.Quest.Nom, 40)
	a.Quest.Objectif = clean(a.Quest.Objectif, 200)
	a.Quest.Obstacle = clean(a.Quest.Obstacle, 200)
	a.Quest.Information = clean(a.Quest.Information, 400)

	// Numeric bounds.
	a.Stats.Force = clampStat(a.Stats.Force)
	a.Stats.Constitution = clampStat(a.Stats.Constitution)
	a.Stats.Vitesse = clampStat(a.Stats.Vitesse)
	a.Stats.Charisme = clampStat(a.Stats.Charisme)
	a.Stats.Savoir = clampStat(a.Stats.Savoir)
	a.Stats.Instinct = clampStat(a.Stats.Instinct)
	if a.PV < 0 || a.PV > 1_000_000 {
		a.PV = 0
	}
	if a.Or < 0 || a.Or > 1_000_000_000 {
		a.Or = 0
	}
	if a.Count < 0 || a.Count > 20 {
		a.Count = 1
	}
}

func clampStat(v float64) float64 {
	if v < 1 {
		return 1
	}
	if v > 30 {
		return 30
	}
	return v
}