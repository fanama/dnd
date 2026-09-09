package services

import (
	"dnd-backend/internal/domain"
	"github.com/google/uuid"
)

// --- Player management (DM only) ---

func (gm *GameManager) dmEditStats(targetPseudo string, stats domain.Stats) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil {
		return
	}
	char.Stats.Force = stats.Force
	char.Stats.Constitution = stats.Constitution
	char.Stats.Vitesse = stats.Vitesse
	char.Stats.Charisme = stats.Charisme
	char.Stats.Savoir = stats.Savoir
	char.Stats.Instinct = stats.Instinct
	if stats.Nom != "" {
		char.Stats.Nom = stats.Nom
	}
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("📜 Le Maître du Donjon a modifié les stats de %s.", char.Stats.Nom)
	gm.NotifyChange()
}

func (gm *GameManager) dmSetPV(targetPseudo string, pv float64) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil {
		return
	}
	char.CurrentPV = pv
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("❤️ Le Maître du Donjon a mis les PV de %s à %.0f.", char.Stats.Nom, pv)
	gm.NotifyChange()
}

func (gm *GameManager) dmAddItem(targetPseudo string, item domain.Item) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil {
		return
	}
	char.Inventaire = append(char.Inventaire, item)
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("🎁 Le MDJ a ajouté \"%s\" à l'inventaire de %s.", item.Nom, char.Stats.Nom)
	gm.NotifyChange()
}

func (gm *GameManager) dmRemoveItem(targetPseudo string, index int) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil || index < 0 || index >= len(char.Inventaire) {
		return
	}
	removed := char.Inventaire[index]
	char.Inventaire = append(char.Inventaire[:index], char.Inventaire[index+1:]...)
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("🗑️ Le MDJ a retiré \"%s\" de l'inventaire de %s.", removed.Nom, char.Stats.Nom)
	gm.NotifyChange()
}

func (gm *GameManager) dmAddSpell(targetPseudo string, spell domain.Sort) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil {
		return
	}
	char.Sorts = append(char.Sorts, spell)
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("✨ Le MDJ a ajouté le sort \"%s\" à %s.", spell.Nom, char.Stats.Nom)
	gm.NotifyChange()
}

func (gm *GameManager) dmRemoveSpell(targetPseudo string, index int) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil || index < 0 || index >= len(char.Sorts) {
		return
	}
	removed := char.Sorts[index]
	char.Sorts = append(char.Sorts[:index], char.Sorts[index+1:]...)
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("🚫 Le MDJ a retiré le sort \"%s\" de %s.", removed.Nom, char.Stats.Nom)
	gm.NotifyChange()
}

func (gm *GameManager) dmMovePlayer(targetPseudo string, destination string) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil {
		return
	}
	char.Lieu = destination
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("🌀 Le MDJ a téléporté %s vers %s.", char.Stats.Nom, destination)
	gm.NotifyChange()
}

func (gm *GameManager) dmDeletePlayer(targetPseudo string) {
	char := gm.getLatestCharacter(targetPseudo)
	name := targetPseudo
	if char != nil {
		name = char.Stats.Nom
	}
	delete(gm.World.Players, targetPseudo)
	delete(gm.Connections, targetPseudo)
	delete(gm.DMs, targetPseudo)
	gm.Repo.DeleteCharacter(targetPseudo)
	gm.chat("💀 Le MDJ a supprimé %s du monde.", name)
	gm.NotifyChange()
}

func (gm *GameManager) dmEditAlign(targetPseudo string, align string) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil {
		return
	}
	char.Alignement = align
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("⚖️ Le MDJ a changé l'alignement de %s à %s.", char.Stats.Nom, align)
	gm.NotifyChange()
}

// --- Location management (DM only) ---

func (gm *GameManager) dmAddLocationItem(locationName string, item domain.Item) {
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == locationName {
			gm.World.Locations[i].Objects = append(gm.World.Locations[i].Objects, item)
			gm.chat("📦 Le MDJ a ajouté \"%s\" à %s.", item.Nom, locationName)
			gm.persistWorld()
			gm.NotifyChange()
			return
		}
	}
}

func (gm *GameManager) dmRemoveLocationItem(locationName string, index int) {
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == locationName {
			loc := &gm.World.Locations[i]
			if index < 0 || index >= len(loc.Objects) {
				return
			}
			removed := loc.Objects[index]
			loc.Objects = append(loc.Objects[:index], loc.Objects[index+1:]...)
			gm.chat("🗑️ Le MDJ a retiré \"%s\" de %s.", removed.Nom, locationName)
			gm.persistWorld()
			gm.NotifyChange()
			return
		}
	}
}

func (gm *GameManager) dmEditLocation(oldName, newName, newBg string) {
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == oldName {
			if newName != "" {
				gm.World.Locations[i].Nom = newName
				// Update all players at this location
				for _, player := range gm.World.Players {
					for _, char := range player.Characters {
						if char.Lieu == oldName {
							char.Lieu = newName
						}
					}
				}
				// Update all NPCs at this location
				for _, npc := range gm.World.NPCs {
					if npc.Lieu == oldName {
						npc.Lieu = newName
					}
				}
			}
			if newBg != "" {
				gm.World.Locations[i].Background = newBg
			}
			gm.chat("🗺️ Le MDJ a modifié le lieu \"%s\".", oldName)
			gm.persistWorld()
			gm.NotifyChange()
			return
		}
	}
}

// --- NPC management (DM only) ---

func (gm *GameManager) dmAddNPC(name, location string, pv float64, align string) {
	if name == "" {
		return
	}
	// Check existing
	if _, exists := gm.World.NPCs[name]; exists {
		gm.chat("⚠️ Un PNJ nommé \"%s\" existe déjà.", name)
		return
	}
	stats := domain.Stats{
		Nom:        name,
		Background: "PNJ",
		Force:      10, Constitution: 10, Vitesse: 10, Charisme: 10, Savoir: 10, Instinct: 10,
	}
	if pv <= 0 {
		pv = stats.CalculateLifePoints()
	}
	if align == "" {
		align = "Neutre"
	}
	npc := &domain.Character{
		ID:         uuid.New(),
		Alignement: align,
		Stats:      stats,
		CurrentPV:  pv,
		Lieu:       location,
		Inventaire: []domain.Item{},
		Sorts:      []domain.Sort{},
	}
	gm.World.NPCs[name] = npc
	gm.chat("🤝 Le MDJ a ajouté le PNJ \"%s\" à %s.", name, location)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmRemoveNPC(name string) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	delete(gm.World.NPCs, name)
	gm.chat("🗑️ Le MDJ a supprimé le PNJ \"%s\" du monde.", npc.Stats.Nom)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmEditNPC(name string, stats domain.Stats, pv float64, align string, overwrite bool) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	if stats.Nom != "" {
		npc.Stats.Nom = stats.Nom
	}
	if stats.Background != "" {
		npc.Stats.Background = stats.Background
	}
	if overwrite {
		npc.Stats.Force = stats.Force
		npc.Stats.Constitution = stats.Constitution
		npc.Stats.Vitesse = stats.Vitesse
		npc.Stats.Charisme = stats.Charisme
		npc.Stats.Savoir = stats.Savoir
		npc.Stats.Instinct = stats.Instinct
		npc.CurrentPV = pv
	} else {
		if stats.Force != 0 {
			npc.Stats.Force = stats.Force
		}
		if stats.Constitution != 0 {
			npc.Stats.Constitution = stats.Constitution
		}
		if stats.Vitesse != 0 {
			npc.Stats.Vitesse = stats.Vitesse
		}
		if stats.Charisme != 0 {
			npc.Stats.Charisme = stats.Charisme
		}
		if stats.Savoir != 0 {
			npc.Stats.Savoir = stats.Savoir
		}
		if stats.Instinct != 0 {
			npc.Stats.Instinct = stats.Instinct
		}
		if pv > 0 {
			npc.CurrentPV = pv
		}
	}
	if align != "" {
		npc.Alignement = align
	}
	gm.chat("📜 Le MDJ a modifié le PNJ \"%s\".", npc.Stats.Nom)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmMoveNPC(name, destination string) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	npc.Lieu = destination
	gm.chat("🌀 Le MDJ a déplacé le PNJ \"%s\" vers %s.", npc.Stats.Nom, destination)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCAddItem(name string, item domain.Item) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	npc.Inventaire = append(npc.Inventaire, item)
	gm.chat("🎁 Le MDJ a donné \"%s\" au PNJ \"%s\".", item.Nom, npc.Stats.Nom)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCRemoveItem(name string, index int) {
	npc, ok := gm.World.NPCs[name]
	if !ok || index < 0 || index >= len(npc.Inventaire) {
		return
	}
	removed := npc.Inventaire[index]
	npc.Inventaire = append(npc.Inventaire[:index], npc.Inventaire[index+1:]...)
	gm.chat("🗑️ Le MDJ a retiré \"%s\" du PNJ \"%s\".", removed.Nom, npc.Stats.Nom)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCAddSpell(name string, spell domain.Sort) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	npc.Sorts = append(npc.Sorts, spell)
	gm.chat("✨ Le MDJ a ajouté le sort \"%s\" au PNJ \"%s\".", spell.Nom, npc.Stats.Nom)
	gm.persistWorld()
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCRemoveSpell(name string, index int) {
	npc, ok := gm.World.NPCs[name]
	if !ok || index < 0 || index >= len(npc.Sorts) {
		return
	}
	removed := npc.Sorts[index]
	npc.Sorts = append(npc.Sorts[:index], npc.Sorts[index+1:]...)
	gm.chat("🚫 Le MDJ a retiré le sort \"%s\" du PNJ \"%s\".", removed.Nom, npc.Stats.Nom)
	gm.persistWorld()
	gm.NotifyChange()
}

// --- Quest management (DM only) ---

func (gm *GameManager) dmAddQuest(locationName string, quest domain.Quest) {
	if quest.Nom == "" {
		return
	}
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == locationName {
			gm.World.Locations[i].Quests = append(gm.World.Locations[i].Quests, quest)
			gm.chat("📌 Le MDJ a ajouté la quête « %s » à %s.", quest.Nom, locationName)
			gm.persistWorld()
			gm.NotifyChange()
			return
		}
	}
}

func (gm *GameManager) dmRemoveQuest(locationName string, index int) {
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == locationName {
			loc := &gm.World.Locations[i]
			if index < 0 || index >= len(loc.Quests) {
				return
			}
			removed := loc.Quests[index]
			loc.Quests = append(loc.Quests[:index], loc.Quests[index+1:]...)
			gm.chat("🗑️ Le MDJ a retiré la quête « %s » de %s.", removed.Nom, locationName)
			gm.persistWorld()
			gm.NotifyChange()
			return
		}
	}
}

func (gm *GameManager) dmAddPlayerQuest(targetPseudo string, quest domain.Quest) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil || quest.Nom == "" {
		return
	}
	for _, q := range char.Quests {
		if q.Nom == quest.Nom {
			gm.chat("⚠️ %s a déjà la quête « %s ».", char.Stats.Nom, quest.Nom)
			return
		}
	}
	char.Quests = append(char.Quests, quest)
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("📜 Le MDJ a confié la quête « %s » à %s.", quest.Nom, char.Stats.Nom)
	gm.NotifyChange()
}

func (gm *GameManager) dmRemovePlayerQuest(targetPseudo string, index int) {
	char := gm.getLatestCharacter(targetPseudo)
	if char == nil || index < 0 || index >= len(char.Quests) {
		return
	}
	removed := char.Quests[index]
	char.Quests = append(char.Quests[:index], char.Quests[index+1:]...)
	gm.saveCharacterState(targetPseudo, char)
	gm.chat("🚫 Le MDJ a retiré la quête « %s » de %s.", removed.Nom, char.Stats.Nom)
	gm.NotifyChange()
}
