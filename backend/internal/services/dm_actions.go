package services

import (
	"fmt"
	"strings"

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

// mobPresetStats returns the "à la volée" default stats for a MOB type.
func mobPresetStats(mobType string) domain.Stats {
	switch mobType {
	case "boss":
		return domain.Stats{Force: 18, Constitution: 16, Vitesse: 10, Charisme: 8, Savoir: 12, Instinct: 12}
	case "minion":
		return domain.Stats{Force: 12, Constitution: 8, Vitesse: 12, Charisme: 6, Savoir: 4, Instinct: 8}
	default:
		return domain.Stats{Force: 10, Constitution: 10, Vitesse: 10, Charisme: 10, Savoir: 10, Instinct: 10}
	}
}

// mobPresetPV returns the default hit points for a MOB type when none is given.
func mobPresetPV(mobType string) float64 {
	switch mobType {
	case "boss":
		return 300
	case "minion":
		return 50
	default:
		return 0 // 0 → derive from Constitution
	}
}

func (gm *GameManager) uniqueNPCName(base string) string {
	name := base
	for i := 1; ; i++ {
		if _, exists := gm.World.NPCs[name]; !exists {
			return name
		}
		name = fmt.Sprintf("%s (%d)", base, i)
	}
}

// dmAddNPC creates one or more NPCs/MOBs "à la volée". `mobType` is empty
// (PNJ), "boss" or "minion"; `count` spawns several copies for minions.
func (gm *GameManager) dmAddNPC(name, location string, pv float64, align string, stats domain.Stats, mobType string, count int) {
	mobType = strings.ToLower(strings.TrimSpace(mobType))
	if mobType != "boss" && mobType != "minion" {
		mobType = ""
	}
	if name == "" {
		return
	}
	if count < 1 {
		count = 1
	}
	if mobType == "boss" {
		count = 1
	}
	if count > 20 {
		count = 20
	}

	label := "PNJ"
	if mobType == "boss" {
		label = "Boss"
	} else if mobType == "minion" {
		label = "Minion"
	}

	statsProvided := stats.Nom != "" || stats.Background != "" ||
		stats.Force != 0 || stats.Constitution != 0 || stats.Vitesse != 0 ||
		stats.Charisme != 0 || stats.Savoir != 0 || stats.Instinct != 0

	spawned := 0
	for i := 0; i < count; i++ {
		displayName := name
		if count > 1 {
			displayName = fmt.Sprintf("%s #%d", name, i+1)
		}
		displayName = gm.uniqueNPCName(displayName)

		base := stats
		if statsProvided {
			base.Background = stats.Background
			if base.Background == "" {
				base.Background = label
			}
		} else {
			base = mobPresetStats(mobType)
			base.Background = label
		}
		base.Nom = displayName

		if pv <= 0 {
			pv = mobPresetPV(mobType)
			if pv <= 0 {
				pv = base.CalculateLifePoints()
			}
		}
		if align == "" {
			align = "Neutre"
		}

		gm.World.NPCs[displayName] = &domain.Character{
			ID:         uuid.New(),
			Alignement: align,
			Stats:      base,
			CurrentPV:  pv,
			Lieu:       location,
			Inventaire: []domain.Item{},
			Sorts:      []domain.Sort{},
			MobType:    mobType,
		}
		spawned++
	}

	switch {
	case mobType == "boss":
		gm.chat("🐲 Le MDJ a ajouté le BOSS \"%s\" à %s.", name, location)
	case mobType == "minion" && spawned > 1:
		gm.chat("👹 Le MDJ a ajouté %d minions \"%s\" à %s.", spawned, name, location)
	case mobType == "minion":
		gm.chat("👹 Le MDJ a ajouté le minion \"%s\" à %s.", name, location)
	default:
		gm.chat("🤝 Le MDJ a ajouté le PNJ \"%s\" à %s.", name, location)
	}
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

// dmRemoveNPCs removes several NPCs/MOBs (boss, minions...) by name in one shot.
func (gm *GameManager) dmRemoveNPCs(names []string) {
	if len(names) == 0 {
		return
	}
	removed := 0
	for _, n := range names {
		if _, ok := gm.World.NPCs[n]; ok {
			delete(gm.World.NPCs, n)
			removed++
		}
	}
	if removed == 0 {
		return
	}
	gm.chat("🗑️ Le MDJ a supprimé %d PNJ/MOB du monde.", removed)
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
