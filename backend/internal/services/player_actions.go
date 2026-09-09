package services

import (
	"fmt"
	"math/rand"
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

// weaponInfo returns the weapon damage dice (number of sides) and whether it is ranged (uses Vitesse).
func weaponInfo(w *domain.Item) (sides int, ranged bool) {
	if w == nil {
		return 2, false // Mains nues (1d2)
	}
	switch {
	case strings.Contains(normalizeName(w.Nom), "dague"):
		return 4, false
	case strings.Contains(normalizeName(w.Nom), "arbalete") || strings.Contains(normalizeName(w.Nom), "carquois"):
		return 8, true
	case strings.Contains(normalizeName(w.Nom), "arc"):
		return 8, true
	case strings.Contains(normalizeName(w.Nom), "epee"):
		return 6, false
	default:
		return 6, false
	}
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

func resolveSpellAttack(attacker, target *domain.Character, spellName string) string {
	spellMod := domain.AbilityModifier(attacker.Stats.Savoir)

	armorBonus := 0.0
	if target.Equipement.Armure != nil {
		armorBonus = target.Equipement.Armure.BonusArmure
	}
	ac := target.Stats.BaseAC() + armorBonus

	roll := rollD20Fn()
	if roll == 1 {
		return fmt.Sprintf("🕯️ %s lance %s sur %s ! [1d20 = 1] ❌ Raté !", attacker.Stats.Nom, spellName, target.Stats.Nom)
	}

	total := float64(roll) + spellMod
	rollLabel := fmt.Sprintf("1d20%+s = %.0f", signed(spellMod), total)

	if roll == 20 || total >= ac {
		isCrit := roll == 20
		critMark := ""
		if isCrit {
			critMark = " 💥 CRITIQUE !"
		}
		dmg := attacker.Stats.Savoir * 1.5
		if isCrit {
			dmg *= 2
		}
		target.CurrentPV -= dmg
		return fmt.Sprintf("🔥%s %s lance %s sur %s ! Jet %s vs CA %.0f → Touché ! Dégâts magiques : %.0f. (PV : %.0f)",
			critMark, attacker.Stats.Nom, spellName, target.Stats.Nom, rollLabel, ac, dmg, target.CurrentPV)
	}

	return fmt.Sprintf("❌ %s lance %s sur %s ! Jet %s < CA %.0f → Raté !", attacker.Stats.Nom, spellName, target.Stats.Nom, rollLabel, ac)
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
		case it.BonusDégâts > 0:
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
	target := gm.getTargetCharacter(targetPseudo)
	if target == nil || attacker.Lieu != target.Lieu {
		return
	}

	msg := resolvePhysicalAttack(attacker, target)
	gm.chat("%s", msg)
	gm.checkDeath(target)
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
	target := gm.getTargetCharacter(targetPseudo)
	if target == nil || char.Lieu != target.Lieu {
		return
	}

	msg := resolveSpellAttack(char, target, spellName)
	gm.chat("%s", msg)
	gm.checkDeath(target)
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
