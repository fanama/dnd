package services

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"dnd-backend/internal/domain"
)

var lootTable = []domain.Item{
	// Consumables
	{Nom: "Potion de Soin", IsConsumable: true, Prix: 25},
	{Nom: "Potion de Force", IsConsumable: true, Prix: 30},
	{Nom: "Potion d'Invisibilité", IsConsumable: true, Prix: 50},
	{Nom: "Eau Bénite", IsConsumable: true, Prix: 20},
	{Nom: "Antidote", IsConsumable: true, Prix: 15},
	// Weapons
	{Nom: "Dague d'Argent", IsConsumable: false, BonusDégâts: 3, Prix: 60},
	{Nom: "Hache de Guerre", IsConsumable: false, BonusDégâts: 6, Prix: 120},
	{Nom: "Arc Court", IsConsumable: false, BonusDégâts: 4, Prix: 80},
	{Nom: "Masse d'Arme", IsConsumable: false, BonusDégâts: 5, Prix: 90},
	{Nom: "Bâton runique", IsConsumable: false, BonusDégâts: 4, Prix: 75},
	{Nom: "Lance Percutante", IsConsumable: false, BonusDégâts: 5, Prix: 95},
	// Armor
	{Nom: "Bouclier en Bois", IsConsumable: false, BonusArmure: 2, Prix: 40},
	{Nom: "Plastron de Fer", IsConsumable: false, BonusArmure: 4, Prix: 150},
	{Nom: "Cape de Cuir Renforcé", IsConsumable: false, BonusArmure: 3, Prix: 100},
	{Nom: "Casque à Crête", IsConsumable: false, BonusArmure: 2, Prix: 60},
	// Misc
	{Nom: "Vieille Carte", IsConsumable: false, Prix: 10},
	{Nom: "Clé Rouillée", IsConsumable: false, Prix: 5},
	{Nom: "Joyau Éclatant", IsConsumable: false, Prix: 80},
	{Nom: "Parchemin Ancien", IsConsumable: false, Prix: 35},
	{Nom: "Corne d'Abondance", IsConsumable: false, Prix: 45},
}

var rollD20Fn = func() int {
	return rand.Intn(20) + 1
}

var rollDiceFn = func(count, sides int) int {
	total := 0
	if count < 1 {
		count = 1
	}
	if sides < 1 {
		sides = 1
	}
	for i := 0; i < count; i++ {
		total += rand.Intn(sides) + 1
	}
	return total
}

func normalizeName(s string) string {
	replacer := strings.NewReplacer(
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"à", "a", "â", "a", "ä", "a",
		"ç", "c",
		"î", "i", "ï", "i",
		"ô", "o", "ö", "o",
		"ù", "u", "û", "u", "ü", "u",
		"É", "e", "È", "e", "Ê", "e", "À", "a", "Â", "a", "Ç", "c", "Î", "i", "Ï", "i", "Ô", "o", "Û", "u",
	)
	return strings.ToLower(replacer.Replace(s))
}

// parseDiceSides extracts the number of faces from a dice string like "d8" or "1d8".
func parseDiceSides(s string) int {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, 'd'); i >= 0 {
		s = s[i+1:]
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 2 {
		return 0
	}
	return n
}

// weaponInfo returns the weapon damage dice (number of faces) and whether it is ranged (uses Vitesse).
// An explicit dice value on the item overrides the dice guessed from its name.
func weaponInfo(w *domain.Item) (sides int, ranged bool) {
	if w == nil {
		return 2, false // Mains nues (1d2)
	}
	norm := normalizeName(w.Nom)
	ranged = strings.Contains(norm, "arc") || strings.Contains(norm, "arbalete") || strings.Contains(norm, "carquois")
	switch {
	case strings.Contains(norm, "dague"):
		sides = 4
	case strings.Contains(norm, "arbalete") || strings.Contains(norm, "carquois"):
		sides = 8
	case strings.Contains(norm, "arc"):
		sides = 8
	case strings.Contains(norm, "epee"):
		sides = 6
	default:
		sides = 6
	}
	if w.DesDégâts != "" {
		if faces := parseDiceSides(w.DesDégâts); faces > 0 {
			sides = faces
		}
	}
	return sides, ranged
}

func signed(v float64) string {
	return fmt.Sprintf("%+.0f", v)
}

func resolvePhysicalAttack(attacker, target *domain.Character) string {
	weapon := attacker.Equipement.Arme
	sides, ranged := weaponInfo(weapon)

	var magicBonus float64
	weaponName := "Mains nues"
	if weapon != nil {
		magicBonus = weapon.BonusDégâts
		weaponName = weapon.Nom
	}

	attackStat := attacker.Stats.Force
	if ranged {
		attackStat = attacker.Stats.Vitesse
	}
	attackMod := domain.AbilityModifier(attackStat)
	dmgMod := domain.AbilityModifier(attackStat)

	armorBonus := 0.0
	if target.Equipement.Armure != nil {
		armorBonus = target.Equipement.Armure.BonusArmure
	}
	ac := target.Stats.BaseAC() + armorBonus

	roll := rollD20Fn()
	if roll == 1 {
		return fmt.Sprintf("💨 %s attaque %s avec %s ! [1d20 = 1] ❌ Raté (fumble) !", attacker.Stats.Nom, target.Stats.Nom, weaponName)
	}

	total := float64(roll) + attackMod + magicBonus
	rollLabel := fmt.Sprintf("1d20%+s%+s = %.0f", signed(attackMod), signed(magicBonus), total)

	if roll == 20 || total >= ac {
		isCrit := roll == 20
		dieCount := 1
		dieDesc := fmt.Sprintf("1d%d", sides)
		critMark := ""
		if isCrit {
			dieCount = 2
			dieDesc = fmt.Sprintf("2d%d", sides)
			critMark = " 💥 CRITIQUE !"
		}
		dmg := float64(rollDiceFn(dieCount, sides)) + dmgMod + magicBonus
		if dmg < 1 {
			dmg = 1
		}
		target.CurrentPV -= dmg
		return fmt.Sprintf("🎯%s %s attaque %s avec %s ! Jet %s vs CA %.0f → Touché ! Dégâts : [%s%+s%+s = %.0f]. (PV : %.0f)",
			critMark, attacker.Stats.Nom, target.Stats.Nom, weaponName, rollLabel, ac, dieDesc, signed(dmgMod), signed(magicBonus), dmg, target.CurrentPV)
	}

	return fmt.Sprintf("❌ %s attaque %s avec %s ! Jet %s < CA %.0f → Raté !", attacker.Stats.Nom, target.Stats.Nom, weaponName, rollLabel, ac)
}

func resolveSpellAttack(attacker, target *domain.Character, spell domain.Sort) string {
	spellMod := domain.AbilityModifier(attacker.Stats.Savoir)
	bonus := spell.Bonus

	armorBonus := 0.0
	if target.Equipement.Armure != nil {
		armorBonus = target.Equipement.Armure.BonusArmure
	}
	ac := target.Stats.BaseAC() + armorBonus

	roll := rollD20Fn()
	if roll == 1 {
		return fmt.Sprintf("🕯️ %s lance %s sur %s ! [1d20 = 1] ❌ Raté !", attacker.Stats.Nom, spell.Nom, target.Stats.Nom)
	}

	total := float64(roll) + spellMod + bonus
	rollLabel := fmt.Sprintf("1d20%+s%+s = %.0f", signed(spellMod), signed(bonus), total)

	if roll == 20 || total >= ac {
		isCrit := roll == 20
		critMark := ""
		if isCrit {
			critMark = " 💥 CRITIQUE !"
		}
		spellName := spell.Nom
		if spell.DesDégâts == "" {
			dmg := attacker.Stats.Savoir * 1.5
			if isCrit {
				dmg *= 2
			}
			target.CurrentPV -= dmg
			return fmt.Sprintf("🔥%s %s lance %s sur %s ! Jet %s vs CA %.0f → Touché ! Dégâts magiques : %.0f. (PV : %.0f)",
				critMark, attacker.Stats.Nom, spellName, target.Stats.Nom, rollLabel, ac, dmg, target.CurrentPV)
		}
		sides := parseDiceSides(spell.DesDégâts)
		if sides <= 0 {
			sides = 6
		}
		dieCount := 1
		dieDesc := fmt.Sprintf("1d%d", sides)
		if isCrit {
			dieCount = 2
			dieDesc = fmt.Sprintf("2d%d", sides)
		}
		dmg := float64(rollDiceFn(dieCount, sides))
		if dmg < 1 {
			dmg = 1
		}
		target.CurrentPV -= dmg
		return fmt.Sprintf("🔥%s %s lance %s sur %s ! Jet %s vs CA %.0f → Touché ! Dégâts : [%s = %.0f]. (PV : %.0f)",
			critMark, attacker.Stats.Nom, spellName, target.Stats.Nom, rollLabel, ac, dieDesc, dmg, target.CurrentPV)
	}

	return fmt.Sprintf("❌ %s lance %s sur %s ! Jet %s < CA %.0f → Raté !", attacker.Stats.Nom, spell.Nom, target.Stats.Nom, rollLabel, ac)
}

func applyStatsBuff(stats *domain.Stats, buff domain.SortBuff) (bool, string) {
	switch strings.ToLower(strings.TrimSpace(buff.Stat)) {
	case "force":
		stats.Force += buff.Valeur
		return true, "Force"
	case "constitution":
		stats.Constitution += buff.Valeur
		return true, "Constitution"
	case "vitesse":
		stats.Vitesse += buff.Valeur
		return true, "Vitesse"
	case "charisme":
		stats.Charisme += buff.Valeur
		return true, "Charisme"
	case "savoir":
		stats.Savoir += buff.Valeur
		return true, "Savoir"
	case "instinct":
		stats.Instinct += buff.Valeur
		return true, "Instinct"
	}
	return false, buff.Stat
}

func (gm *GameManager) getTargetCharacter(targetPseudo string) *domain.Character {
	targetPlayer, ok := gm.World.Players[targetPseudo]
	if !ok || len(targetPlayer.Characters) == 0 {
		return nil
	}
	return targetPlayer.Characters[len(targetPlayer.Characters)-1]
}

func (gm *GameManager) actionEquip(char *domain.Character, itemName string) {
	for i := range char.Inventaire {
		it := char.Inventaire[i]
		if it.Nom != itemName {
			continue
		}
		equipped := it
		switch {
		case it.BonusDégâts > 0 || it.DesDégâts != "":
			char.Equipement.Arme = &equipped
			gm.chat("🗡️ %s équipe %s !", char.Stats.Nom, it.Nom)
		case it.BonusArmure > 0:
			char.Equipement.Armure = &equipped
			gm.chat("🛡️ %s équipe %s !", char.Stats.Nom, it.Nom)
		default:
			gm.chat("❌ %s ne peut pas équiper %s (ni arme ni armure)", char.Stats.Nom, it.Nom)
		}
		return
	}
	gm.chat("❌ %s ne possède pas « %s »", char.Stats.Nom, itemName)
}

func (gm *GameManager) actionUnequip(char *domain.Character, slot string) {
	switch slot {
	case "weapon":
		if char.Equipement.Arme != nil {
			gm.chat("🔄 %s retire %s.", char.Stats.Nom, char.Equipement.Arme.Nom)
			char.Equipement.Arme = nil
		}
	case "armor":
		if char.Equipement.Armure != nil {
			gm.chat("🔄 %s retire %s.", char.Stats.Nom, char.Equipement.Armure.Nom)
			char.Equipement.Armure = nil
		}
	}
}

func (gm *GameManager) actionAttack(attacker *domain.Character, targetPseudo string) {
	if target := gm.getTargetCharacter(targetPseudo); target != nil && attacker.Lieu == target.Lieu {
		gm.chat("%s", resolvePhysicalAttack(attacker, target))
		gm.checkDeath(target)
		return
	}

	npc, ok := gm.World.NPCs[targetPseudo]
	if !ok || attacker.Lieu != npc.Lieu {
		return
	}
	gm.chat("%s", resolvePhysicalAttack(attacker, npc))
	gm.npcAfterDamage(targetPseudo, npc)
	gm.persistWorld()
}

// findLocation returns a pointer to the location with the given name.
func (gm *GameManager) findLocation(name string) *domain.Location {
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == name {
			return &gm.World.Locations[i]
		}
	}
	return nil
}

// npcAfterDamage handles an NPC or MOB that just took damage: clamps PV to 0
// and, on death, drops its inventory as loot on the ground and removes the
// dead NPC from the world.
func (gm *GameManager) npcAfterDamage(name string, npc *domain.Character) {
	if npc.CurrentPV < 0 {
		npc.CurrentPV = 0
	}
	if npc.CurrentPV != 0 {
		return
	}
	loc := gm.findLocation(npc.Lieu)
	if loc != nil && len(npc.Inventaire) > 0 {
		loc.Objects = append(loc.Objects, npc.Inventaire...)
		npc.Inventaire = nil
		gm.chat("☠️ %s est tombé ! Son butin tombe au sol.", npc.Stats.Nom)
	} else {
		gm.chat("☠️ %s est tombé !", npc.Stats.Nom)
	}
	delete(gm.World.NPCs, name)
}

func (gm *GameManager) actionMove(char *domain.Character, dest string) {
	char.Lieu = dest
	gm.chat("🧳 %s s'est déplacé vers : %s.", char.Stats.Nom, dest)
	gm.spawnLoot(dest)
}

func (gm *GameManager) spawnLoot(locationName string) {
	var loc *domain.Location
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == locationName {
			loc = &gm.World.Locations[i]
			break
		}
	}
	if loc == nil {
		return
	}
	if len(loc.Objects) >= 5 {
		return
	}

	numSpawns := rand.Intn(3) // 0, 1, or 2 items
	for i := 0; i < numSpawns && len(loc.Objects) < 5; i++ {
		item := lootTable[rand.Intn(len(lootTable))]
		loc.Objects = append(loc.Objects, item)
		gm.chat("✨ Un objet est apparu dans %s : %s !", locationName, item.Nom)
	}
	gm.persistWorld()
}

func (gm *GameManager) actionCastSpell(char *domain.Character, spellName string, targetPseudo string) {
	var spell domain.Sort
	found := false
	for _, s := range char.Sorts {
		if s.Nom == spellName {
			spell = s
			found = true
			break
		}
	}
	if !found {
		return
	}

	if targetChar := gm.getTargetCharacter(targetPseudo); targetChar != nil && char.Lieu == targetChar.Lieu {
		if spell.Buff != nil {
			if ok, label := applyStatsBuff(&targetChar.Stats, *spell.Buff); ok {
				gm.chat("✨ %s lance %s sur %s ! +%.0f %s (permanent).",
					char.Stats.Nom, spell.Nom, targetChar.Stats.Nom, spell.Buff.Valeur, label)
			}
			gm.saveCharacterState(targetPseudo, targetChar)
			return
		}
		msg := resolveSpellAttack(char, targetChar, spell)
		gm.chat("%s", msg)
		gm.checkDeath(targetChar)
		gm.saveCharacterState(targetPseudo, targetChar)
		return
	}

	npc, ok := gm.World.NPCs[targetPseudo]
	if !ok || char.Lieu != npc.Lieu {
		return
	}
	if spell.Buff != nil {
		if ok2, label := applyStatsBuff(&npc.Stats, *spell.Buff); ok2 {
			gm.chat("✨ %s lance %s sur %s ! +%.0f %s (permanent).",
				char.Stats.Nom, spell.Nom, npc.Stats.Nom, spell.Buff.Valeur, label)
		}
		gm.persistWorld()
		return
	}
	msg := resolveSpellAttack(char, npc, spell)
	gm.chat("%s", msg)
	gm.npcAfterDamage(targetPseudo, npc)
	gm.persistWorld()
}

func (gm *GameManager) actionConsume(char *domain.Character, itemName string) {
	itemIdx := -1
	for i, item := range char.Inventaire {
		if item.Nom == itemName {
			itemIdx = i
			break
		}
	}
	if itemIdx == -1 {
		return
	}
	item := char.Inventaire[itemIdx]
	if item.IsConsumable {
		heal := char.Stats.Constitution * 5
		char.CurrentPV += heal
		if char.CurrentPV > char.Stats.CalculateLifePoints() {
			char.CurrentPV = char.Stats.CalculateLifePoints()
		}
		char.Inventaire = append(char.Inventaire[:itemIdx], char.Inventaire[itemIdx+1:]...)
		gm.chat("🧪 %s boit une %s et récupère %.0f PV !", char.Stats.Nom, item.Nom, heal)
	}
}

func (gm *GameManager) checkDeath(char *domain.Character) {
	if char.CurrentPV <= 0 {
		char.CurrentPV = char.Stats.CalculateLifePoints()
		char.Lieu = "Taverne"
		gm.chat("😇 %s est tombé au combat mais ressuscite à la Taverne !", char.Stats.Nom)
	}
}

func (gm *GameManager) actionLoot(char *domain.Character, itemName string) {
	var currentLocation *domain.Location
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == char.Lieu {
			currentLocation = &gm.World.Locations[i]
			break
		}
	}
	if currentLocation == nil || itemName == "" {
		return
	}
	itemIdx := -1
	for i, obj := range currentLocation.Objects {
		if obj.Nom == itemName {
			itemIdx = i
			break
		}
	}
	if itemIdx == -1 {
		return
	}
	item := currentLocation.Objects[itemIdx]
	char.Inventaire = append(char.Inventaire, item)
	currentLocation.Objects = append(currentLocation.Objects[:itemIdx], currentLocation.Objects[itemIdx+1:]...)
	gm.chat("🎒 %s a ramassé %s dans %s !", char.Stats.Nom, item.Nom, char.Lieu)
	gm.persistWorld()
}

func (gm *GameManager) actionAcceptQuest(char *domain.Character, questName string) {
	quest := gm.findLocationQuest(char.Lieu, questName)
	if quest == nil {
		return
	}
	for _, q := range char.Quests {
		if q.Nom == quest.Nom {
			gm.chat("⚠️ %s a déjà accepté la quête « %s ».", char.Stats.Nom, quest.Nom)
			return
		}
	}
	char.Quests = append(char.Quests, *quest)
	gm.chat("📜 %s a accepté la quête « %s » : %s", char.Stats.Nom, quest.Nom, quest.Objectif)
}

func (gm *GameManager) actionCompleteQuest(char *domain.Character, questName string) {
	idx := -1
	for i, q := range char.Quests {
		if q.Nom == questName {
			idx = i
			break
		}
	}
	if idx == -1 {
		gm.chat("❌ %s n'a pas accepté la quête « %s ».", char.Stats.Nom, questName)
		return
	}
	quest := char.Quests[idx]

	for _, reward := range quest.Recompense {
		char.Inventaire = append(char.Inventaire, reward)
	}
	char.Quests = append(char.Quests[:idx], char.Quests[idx+1:]...)
	if strings.TrimSpace(quest.Information) != "" {
		gm.chat("📜 %s obtient une information : %s", char.Stats.Nom, strings.TrimSpace(quest.Information))
	}
	gm.chat("🏆 %s a terminé la quête « %s » !", char.Stats.Nom, quest.Nom)
}

// findLocationQuest looks up a quest by name in a location's available quests.
func (gm *GameManager) findLocationQuest(locationName, questName string) *domain.Quest {
	for i := range gm.World.Locations {
		loc := &gm.World.Locations[i]
		if loc.Nom != locationName {
			continue
		}
		for j := range loc.Quests {
			if loc.Quests[j].Nom == questName {
				return &loc.Quests[j]
			}
		}
	}
	return nil
}

func (gm *GameManager) hasItem(items []domain.Item, name string) bool {
	for _, it := range items {
		if it.Nom == name {
			return true
		}
	}
	return false
}

func (gm *GameManager) removeItem(items []domain.Item, name string) []domain.Item {
	for i, it := range items {
		if it.Nom == name {
			return append(items[:i], items[i+1:]...)
		}
	}
	return items
}
