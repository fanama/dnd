package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

type Action struct {
	Type         string       `json:"type"`
	Cible        string       `json:"cible,omitempty"`
	Destination  string       `json:"destination,omitempty"`
	Sort         string       `json:"sort,omitempty"`
	ItemIndex    int          `json:"item_index,omitempty"`
	ItemName     string       `json:"item_name,omitempty"`
	LootName     string       `json:"loot_name,omitempty"`
	TargetPlayer string       `json:"target_player,omitempty"`
	Stats        domain.Stats `json:"stats,omitempty"`
	Item         domain.Item  `json:"item,omitempty"`
	Spell        domain.Sort  `json:"spell,omitempty"`
	PV           float64      `json:"pv,omitempty"`
	Alignement   string       `json:"alignement,omitempty"`
	Overwrite    bool         `json:"overwrite,omitempty"`
	NewName      string       `json:"new_name,omitempty"`
	LocationBg   string       `json:"location_bg,omitempty"`
	Slot         string       `json:"slot,omitempty"`
}

type GameManager struct {
	Connections map[string]*websocket.Conn
	World       *domain.World
	Repo        *repository.SQLiteRepository
	DMs         map[string]bool
	mu          sync.Mutex
}

func NewGameManager(repo *repository.SQLiteRepository) *GameManager {
	gm := &GameManager{
		Connections: make(map[string]*websocket.Conn),
		DMs:         make(map[string]bool),
		Repo:        repo,
		World: &domain.World{
			ID:      uuid.New(),
			Players: make(map[string]*domain.Player),
			NPCs:    make(map[string]*domain.Character),
			Locations: []domain.Location{
				{Nom: "Taverne", Background: "Ambiance chaleureuse, odeur de biere", Objects: []domain.Item{
					{Nom: "Vieille Carte", IsConsumable: false, Prix: 10},
					{Nom: "Chope de Biere", IsConsumable: true, Prix: 5},
				}},
				{Nom: "Donjon", Background: "Sombre et humide, murs couverts de mousse", Objects: []domain.Item{
					{Nom: "Épée Rouillée", IsConsumable: false, BonusDégâts: 2, Prix: 40},
					{Nom: "Potion de Soin", IsConsumable: true, Prix: 25},
				}},
				{Nom: "Foret Enchantee", Background: "Arbres millenaires, lumiere filtreee", Objects: []domain.Item{
					{Nom: "Herbes Medecinales", IsConsumable: true, Prix: 20},
					{Nom: "Arc Elfe", IsConsumable: false, BonusDégâts: 4, Prix: 85},
				}},
				{Nom: "Montagne Rocheuse", Background: "Pics aceres, vent glacial", Objects: []domain.Item{
					{Nom: "Haches de Guerre", IsConsumable: false, BonusDégâts: 6, Prix: 130},
					{Nom: "Gantelets de Fer", IsConsumable: false, BonusArmure: 3, Prix: 90},
				}},
				{Nom: "Marais Hante", Background: "Brume epaisse, craquements suspects", Objects: []domain.Item{
					{Nom: "Potion d'Invisibilite", IsConsumable: true, Prix: 50},
					{Nom: "Fiole de Venom", IsConsumable: true, Prix: 35},
				}},
				{Nom: "Plaine des Conflits", Background: "Champ de bataille, drapeaux dechu", Objects: []domain.Item{
					{Nom: "Bouclier en Bois", IsConsumable: false, BonusArmure: 2, Prix: 40},
					{Nom: "Lance Percutante", IsConsumable: false, BonusDégâts: 5, Prix: 95},
				}},
				{Nom: "Temple Abandonne", Background: "Piliers brises, ombres dansantes", Objects: []domain.Item{
					{Nom: "Sceptre Sacre", IsConsumable: false, BonusDégâts: 7, Prix: 160},
					{Nom: "Parchemin Ancien", IsConsumable: false, Prix: 35},
				}},
			},
		},
	}
	gm.seedDefaultNPCs()
	return gm
}

func (gm *GameManager) seedDefaultNPCs() {
	defaults := []struct {
		name, lieu, classe, align           string
		pv, force, con, vit, cha, sav, inst float64
		items                               []domain.Item
	}{
		{"Arnold le Tavernier", "Taverne", "Aubergiste", "Neutre Bon", 60, 12, 12, 10, 14, 10, 10,
			[]domain.Item{{Nom: "Chope de Biere", IsConsumable: true, Prix: 5}}},
		{"Mira la Voyante", "Taverne", "Devineresse", "Chaotique Neutre", 45, 8, 9, 12, 9, 16, 15,
			[]domain.Item{{Nom: "Boule de Cristal", Prix: 200}}},
		{"Grum le Garde", "Donjon", "Gardien", "Loyal Neutre", 80, 15, 14, 11, 8, 9, 10,
			[]domain.Item{{Nom: "Hallebarde", BonusDégâts: 7, Prix: 120}}},
		{"Elara la Dryade", "Foret Enchantee", "Gardienne de la Foret", "Neutre Bon", 55, 9, 10, 15, 12, 11, 16,
			[]domain.Item{{Nom: "Herbes Medecinales", IsConsumable: true, Prix: 20}}},
		{"Boris le Forgeron", "Montagne Rocheuse", "Forgeron", "Loyal Neutre", 70, 16, 14, 9, 10, 9, 10,
			[]domain.Item{{Nom: "Marteau de Forgeron", BonusDégâts: 6, Prix: 110}}},
		{"Zorra la Sorciere", "Marais Hante", "Sorciere", "Chaotique Mauvais", 40, 7, 10, 11, 10, 17, 14,
			[]domain.Item{{Nom: "Fiole de Venom", IsConsumable: true, Prix: 35}}},
		{"Sir Aldric", "Temple Abandonne", "Paladin", "Loyal Bon", 90, 17, 16, 10, 12, 11, 11,
			[]domain.Item{{Nom: "Epée Sacrée", BonusDégâts: 8, Prix: 250}}},
	}
	for _, d := range defaults {
		stats := domain.Stats{
			Nom: d.name, Background: d.classe,
			Force: d.force, Constitution: d.con, Vitesse: d.vit,
			Charisme: d.cha, Savoir: d.sav, Instinct: d.inst,
		}
		gm.World.NPCs[d.name] = &domain.Character{
			ID:         uuid.New(),
			Alignement: d.align,
			Stats:      stats,
			CurrentPV:  d.pv,
			Lieu:       d.lieu,
			Inventaire: d.items,
			Sorts:      []domain.Sort{},
		}
	}
}

func (gm *GameManager) Connect(pseudo string, ws *websocket.Conn, charInfo map[string]string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.Connections[pseudo] = ws

	// Detect DM role
	isDM := pseudo == "dm" || len(pseudo) > 3 && pseudo[:3] == "dm_"
	gm.DMs[pseudo] = isDM

	player, ok := gm.World.Players[pseudo]
	if !ok {
		player = &domain.Player{Pseudo: pseudo}
		gm.World.Players[pseudo] = player
	}

	// 1. Attempt to restore from DB
	_, _, lieu, pv, _, invStr, statsStr, equipStr, err := gm.Repo.GetCharacter(pseudo)
	if err == nil {
		var stats domain.Stats
		json.Unmarshal([]byte(statsStr), &stats)
		var inventory []domain.Item
		json.Unmarshal([]byte(invStr), &inventory)
		var equip domain.Equipment
		json.Unmarshal([]byte(equipStr), &equip)

		char := &domain.Character{
			ID:         uuid.New(),
			Alignement: "Neutre",
			Stats:      stats,
			CurrentPV:  float64(pv),
			Lieu:       lieu,
			Inventaire: inventory,
			Equipement: equip,
		}
		player.Characters = append(player.Characters, char)
		gm.broadcast(map[string]interface{}{
			"type": "chat",
			"msg":  fmt.Sprintf("👋 %s est revenu dans le monde !", pseudo),
		})
	} else {
		// 2. New character creation
		charName := charInfo["nom_personnage"]
		charClass := charInfo["classe"]

		stats := domain.Stats{Nom: charName, Background: charClass}
		switch charClass {
		case "Magicien":
			stats.Force = 8
			stats.Constitution = 9
			stats.Vitesse = 11
			stats.Charisme = 12
			stats.Instinct = 14
			stats.Savoir = 16
		case "Voleur":
			stats.Force = 10
			stats.Constitution = 8
			stats.Vitesse = 16
			stats.Charisme = 10
			stats.Instinct = 15
			stats.Savoir = 11
		case "Clerc":
			stats.Force = 12
			stats.Constitution = 14
			stats.Vitesse = 8
			stats.Charisme = 15
			stats.Instinct = 9
			stats.Savoir = 12
		case "Barde":
			stats.Force = 9
			stats.Constitution = 10
			stats.Vitesse = 13
			stats.Charisme = 16
			stats.Instinct = 12
			stats.Savoir = 10
		case "Ranger":
			stats.Force = 13
			stats.Constitution = 11
			stats.Vitesse = 14
			stats.Charisme = 8
			stats.Instinct = 15
			stats.Savoir = 9
		default: // Guerrier
			stats.Force = 15
			stats.Constitution = 12
			stats.Vitesse = 10
			stats.Charisme = 10
			stats.Instinct = 10
			stats.Savoir = 10
		}

		char := &domain.Character{
			ID:         uuid.New(),
			Alignement: "Neutre",
			Stats:      stats,
			CurrentPV:  stats.CalculateLifePoints(),
			Lieu:       "Taverne",
		}

		switch charClass {
		case "Magicien":
			char.Sorts = append(char.Sorts, domain.Sort{Nom: "Boule de Feu", EcoleMagie: "Évocations"})
		case "Clerc":
			char.Sorts = append(char.Sorts, domain.Sort{Nom: "Soin Divin", EcoleMagie: "Guérison"})
		case "Barde":
			char.Sorts = append(char.Sorts, domain.Sort{Nom: "Mélodie Envoûtante", EcoleMagie: "Enchantement"})
		}

		switch charClass {
		case "Guerrier":
			char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Épée Longue", IsConsumable: false, BonusDégâts: 5, Prix: 100})
		case "Magicien":
			char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Bâton Mystique", IsConsumable: false, BonusDégâts: 3, Prix: 80})
		case "Voleur":
			char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Dague Empoisonnée", IsConsumable: false, BonusDégâts: 4, Prix: 90})
		case "Clerc":
			char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Marteau Sacré", IsConsumable: false, BonusDégâts: 4, Prix: 95})
		case "Barde":
			char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Luth Enchanté", IsConsumable: false, BonusDégâts: 2, Prix: 70})
		case "Ranger":
			char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Arc Long", IsConsumable: false, BonusDégâts: 5, Prix: 100})
		}
		char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Potion de Soin", IsConsumable: true, Prix: 25})

		for _, it := range char.Inventaire {
			if !it.IsConsumable && it.BonusDégâts > 0 {
				equipped := it
				char.Equipement.Arme = &equipped
				break
			}
		}

		player.Characters = append(player.Characters, char)
		gm.saveCharacterState(pseudo, char)
		gm.broadcast(map[string]interface{}{
			"type": "chat",
			"msg":  fmt.Sprintf("⚔️ %s a incarné %s (%s) !", pseudo, charName, charClass),
		})
	}
	gm.NotifyChange()
}

func (gm *GameManager) HandleAction(pseudo string, action Action) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// DM actions
	if gm.DMs[pseudo] {
		switch action.Type {
		case "dm_edit_stats":
			gm.dmEditStats(action.TargetPlayer, action.Stats)
			return
		case "dm_set_pv":
			gm.dmSetPV(action.TargetPlayer, action.PV)
			return
		case "dm_add_item":
			gm.dmAddItem(action.TargetPlayer, action.Item)
			return
		case "dm_remove_item":
			gm.dmRemoveItem(action.TargetPlayer, action.ItemIndex)
			return
		case "dm_add_spell":
			gm.dmAddSpell(action.TargetPlayer, action.Spell)
			return
		case "dm_remove_spell":
			gm.dmRemoveSpell(action.TargetPlayer, action.ItemIndex)
			return
		case "dm_move_player":
			gm.dmMovePlayer(action.TargetPlayer, action.Destination)
			return
		case "dm_teleport_item_add":
			gm.dmAddLocationItem(action.Destination, action.Item)
			return
		case "dm_teleport_item_remove":
			gm.dmRemoveLocationItem(action.Destination, action.ItemIndex)
			return
		case "dm_delete_player":
			gm.dmDeletePlayer(action.TargetPlayer)
			return
		case "dm_edit_align":
			gm.dmEditAlign(action.TargetPlayer, action.Alignement)
			return
		case "dm_edit_location":
			gm.dmEditLocation(action.Destination, action.NewName, action.LocationBg)
			return
		case "dm_add_npc":
			gm.dmAddNPC(action.ItemName, action.Destination, action.PV, action.Alignement)
			return
		case "dm_remove_npc":
			gm.dmRemoveNPC(action.ItemName)
			return
		case "dm_edit_npc":
			gm.dmEditNPC(action.ItemName, action.Stats, action.PV, action.Alignement, action.Overwrite)
			return
		case "dm_move_npc":
			gm.dmMoveNPC(action.ItemName, action.Destination)
			return
		case "dm_npc_add_item":
			gm.dmNPCAddItem(action.ItemName, action.Item)
			return
		case "dm_npc_remove_item":
			gm.dmNPCRemoveItem(action.ItemName, action.ItemIndex)
			return
		case "dm_npc_add_spell":
			gm.dmNPCAddSpell(action.ItemName, action.Spell)
			return
		case "dm_npc_remove_spell":
			gm.dmNPCRemoveSpell(action.ItemName, action.ItemIndex)
			return
		}
	}

	char := gm.getLatestCharacter(pseudo)
	if char == nil {
		return
	}

	switch action.Type {
	case "attack":
		gm.actionAttack(char, action.Cible)
	case "move":
		gm.actionMove(char, action.Destination)
	case "cast_spell":
		gm.actionCastSpell(char, action.Sort, action.Cible)
	case "use_consumable":
		gm.actionConsume(char, action.ItemName)
	case "loot":
		gm.actionLoot(char, action.LootName)
	case "equip_item":
		gm.actionEquip(char, action.ItemName)
	case "unequip_item":
		gm.actionUnequip(char, action.Slot)
	}

	gm.saveCharacterState(pseudo, char)
	gm.NotifyChange()
}

func (gm *GameManager) saveCharacterState(pseudo string, char *domain.Character) {
	statsJson, _ := json.Marshal(char.Stats)
	invJson, _ := json.Marshal(char.Inventaire)
	equipJson, _ := json.Marshal(char.Equipement)
	gm.Repo.SaveCharacter(pseudo, char.Stats.Nom, char.Stats.Background, char.Lieu, int(char.CurrentPV), int(char.Stats.CalculateLifePoints()), string(invJson), string(statsJson), string(equipJson))
}

func (gm *GameManager) getTargetCharacter(targetPseudo string) *domain.Character {
	targetPlayer, ok := gm.World.Players[targetPseudo]
	if !ok || len(targetPlayer.Characters) == 0 {
		return nil
	}
	return targetPlayer.Characters[len(targetPlayer.Characters)-1]
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
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("🗡️ %s équipe %s !", char.Stats.Nom, it.Nom),
			})
		case it.BonusArmure > 0:
			char.Equipement.Armure = &equipped
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("🛡️ %s équipe %s !", char.Stats.Nom, it.Nom),
			})
		default:
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("❌ %s ne peut pas équiper %s (ni arme ni armure)", char.Stats.Nom, it.Nom),
			})
		}
		return
	}
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("❌ %s ne possède pas « %s »", char.Stats.Nom, itemName),
	})
}

func (gm *GameManager) actionUnequip(char *domain.Character, slot string) {
	switch slot {
	case "weapon":
		if char.Equipement.Arme != nil {
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("🔄 %s retire %s.", char.Stats.Nom, char.Equipement.Arme.Nom),
			})
			char.Equipement.Arme = nil
		}
	case "armor":
		if char.Equipement.Armure != nil {
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("🔄 %s retire %s.", char.Stats.Nom, char.Equipement.Armure.Nom),
			})
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
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  msg,
	})
	gm.checkDeath(target)
}

func (gm *GameManager) actionMove(char *domain.Character, dest string) {
	char.Lieu = dest
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🧳 %s s'est déplacé vers : %s.", char.Stats.Nom, dest),
	})
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
		gm.broadcast(map[string]interface{}{
			"type": "chat",
			"msg":  fmt.Sprintf("✨ Un objet est apparu dans %s : %s !", locationName, item.Nom),
		})
	}
}

func (gm *GameManager) actionCastSpell(char *domain.Character, spellName string, targetPseudo string) {
	target := gm.getTargetCharacter(targetPseudo)
	if target == nil || char.Lieu != target.Lieu {
		return
	}

	msg := resolveSpellAttack(char, target, spellName)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  msg,
	})
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
		gm.broadcast(map[string]interface{}{
			"type": "chat",
			"msg":  fmt.Sprintf("🧪 %s boit une %s et récupère %.0f PV !", char.Stats.Nom, item.Nom, heal),
		})
	}
}

func (gm *GameManager) checkDeath(char *domain.Character) {
	if char.CurrentPV <= 0 {
		char.CurrentPV = char.Stats.CalculateLifePoints()
		char.Lieu = "Taverne"
		gm.broadcast(map[string]interface{}{
			"type": "chat",
			"msg":  fmt.Sprintf("😇 %s est tombé au combat mais ressuscite à la Taverne !", char.Stats.Nom),
		})
	}
}

func (gm *GameManager) getLatestCharacter(pseudo string) *domain.Character {
	player, ok := gm.World.Players[pseudo]
	if !ok || len(player.Characters) == 0 {
		return nil
	}
	return player.Characters[len(player.Characters)-1]
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
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🎒 %s a ramassé %s dans %s !", char.Stats.Nom, item.Nom, char.Lieu),
	})
}

// --- DM Actions ---

func (gm *GameManager) getCharacterByPseudo(targetPseudo string) *domain.Character {
	player, ok := gm.World.Players[targetPseudo]
	if !ok || len(player.Characters) == 0 {
		return nil
	}
	return player.Characters[len(player.Characters)-1]
}

func (gm *GameManager) dmEditStats(targetPseudo string, stats domain.Stats) {
	char := gm.getCharacterByPseudo(targetPseudo)
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
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("📜 Le Maître du Donjon a modifié les stats de %s.", char.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmSetPV(targetPseudo string, pv float64) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil {
		return
	}
	char.CurrentPV = pv
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("❤️ Le Maître du Donjon a mis les PV de %s à %.0f.", char.Stats.Nom, pv),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmAddItem(targetPseudo string, item domain.Item) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil {
		return
	}
	char.Inventaire = append(char.Inventaire, item)
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🎁 Le MDJ a ajouté \"%s\" à l'inventaire de %s.", item.Nom, char.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmRemoveItem(targetPseudo string, index int) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil || index < 0 || index >= len(char.Inventaire) {
		return
	}
	removed := char.Inventaire[index]
	char.Inventaire = append(char.Inventaire[:index], char.Inventaire[index+1:]...)
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🗑️ Le MDJ a retiré \"%s\" de l'inventaire de %s.", removed.Nom, char.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmAddSpell(targetPseudo string, spell domain.Sort) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil {
		return
	}
	char.Sorts = append(char.Sorts, spell)
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("✨ Le MDJ a ajouté le sort \"%s\" à %s.", spell.Nom, char.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmRemoveSpell(targetPseudo string, index int) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil || index < 0 || index >= len(char.Sorts) {
		return
	}
	removed := char.Sorts[index]
	char.Sorts = append(char.Sorts[:index], char.Sorts[index+1:]...)
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🚫 Le MDJ a retiré le sort \"%s\" de %s.", removed.Nom, char.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmMovePlayer(targetPseudo string, destination string) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil {
		return
	}
	char.Lieu = destination
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🌀 Le MDJ a téléporté %s vers %s.", char.Stats.Nom, destination),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmAddLocationItem(locationName string, item domain.Item) {
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == locationName {
			gm.World.Locations[i].Objects = append(gm.World.Locations[i].Objects, item)
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("📦 Le MDJ a ajouté \"%s\" à %s.", item.Nom, locationName),
			})
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
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("🗑️ Le MDJ a retiré \"%s\" de %s.", removed.Nom, locationName),
			})
			gm.NotifyChange()
			return
		}
	}
}

func (gm *GameManager) dmDeletePlayer(targetPseudo string) {
	char := gm.getCharacterByPseudo(targetPseudo)
	name := targetPseudo
	if char != nil {
		name = char.Stats.Nom
	}
	delete(gm.World.Players, targetPseudo)
	delete(gm.Connections, targetPseudo)
	delete(gm.DMs, targetPseudo)
	gm.Repo.DeleteCharacter(targetPseudo)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("💀 Le MDJ a supprimé %s du monde.", name),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmEditAlign(targetPseudo string, align string) {
	char := gm.getCharacterByPseudo(targetPseudo)
	if char == nil {
		return
	}
	char.Alignement = align
	gm.saveCharacterState(targetPseudo, char)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("⚖️ Le MDJ a changé l'alignement de %s à %s.", char.Stats.Nom, align),
	})
	gm.NotifyChange()
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
			gm.broadcast(map[string]interface{}{
				"type": "chat",
				"msg":  fmt.Sprintf("🗺️ Le MDJ a modifié le lieu \"%s\".", oldName),
			})
			gm.NotifyChange()
			return
		}
	}
}

// --- NPC Actions ---

func (gm *GameManager) dmAddNPC(name, location string, pv float64, align string) {
	if name == "" {
		return
	}
	// Check existing
	if _, exists := gm.World.NPCs[name]; exists {
		gm.broadcast(map[string]interface{}{
			"type": "chat",
			"msg":  fmt.Sprintf("⚠️ Un PNJ nommé \"%s\" existe déjà.", name),
		})
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
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🤝 Le MDJ a ajouté le PNJ \"%s\" à %s.", name, location),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmRemoveNPC(name string) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	delete(gm.World.NPCs, name)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🗑️ Le MDJ a supprimé le PNJ \"%s\" du monde.", npc.Stats.Nom),
	})
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
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("📜 Le MDJ a modifié le PNJ \"%s\".", npc.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmMoveNPC(name, destination string) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	npc.Lieu = destination
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🌀 Le MDJ a déplacé le PNJ \"%s\" vers %s.", npc.Stats.Nom, destination),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCAddItem(name string, item domain.Item) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	npc.Inventaire = append(npc.Inventaire, item)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🎁 Le MDJ a donné \"%s\" au PNJ \"%s\".", item.Nom, npc.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCRemoveItem(name string, index int) {
	npc, ok := gm.World.NPCs[name]
	if !ok || index < 0 || index >= len(npc.Inventaire) {
		return
	}
	removed := npc.Inventaire[index]
	npc.Inventaire = append(npc.Inventaire[:index], npc.Inventaire[index+1:]...)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🗑️ Le MDJ a retiré \"%s\" du PNJ \"%s\".", removed.Nom, npc.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCAddSpell(name string, spell domain.Sort) {
	npc, ok := gm.World.NPCs[name]
	if !ok {
		return
	}
	npc.Sorts = append(npc.Sorts, spell)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("✨ Le MDJ a ajouté le sort \"%s\" au PNJ \"%s\".", spell.Nom, npc.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) dmNPCRemoveSpell(name string, index int) {
	npc, ok := gm.World.NPCs[name]
	if !ok || index < 0 || index >= len(npc.Sorts) {
		return
	}
	removed := npc.Sorts[index]
	npc.Sorts = append(npc.Sorts[:index], npc.Sorts[index+1:]...)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🚫 Le MDJ a retiré le sort \"%s\" du PNJ \"%s\".", removed.Nom, npc.Stats.Nom),
	})
	gm.NotifyChange()
}

func (gm *GameManager) broadcast(data interface{}) {
	msg, _ := json.Marshal(data)
	for _, conn := range gm.Connections {
		conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (gm *GameManager) NotifyChange() {
	syncData := make(map[string]interface{})
	for pseudo, player := range gm.World.Players {
		if len(player.Characters) > 0 {
			char := player.Characters[len(player.Characters)-1]
			syncData[pseudo] = map[string]interface{}{
				"nom":        char.Stats.Nom,
				"pv":         char.CurrentPV,
				"max_pv":     char.Stats.CalculateLifePoints(),
				"classe":     char.Stats.Background,
				"lieu":       char.Lieu,
				"alignement": char.Alignement,
				"sorts":      char.Sorts,
				"inventaire": char.Inventaire,
				"equipement": char.Equipement,
				"stats":      char.Stats,
				"role":       gm.DMs[pseudo],
			}
		}
	}
	npcData := make(map[string]interface{})
	for name, npc := range gm.World.NPCs {
		npcData[name] = map[string]interface{}{
			"nom":        npc.Stats.Nom,
			"pv":         npc.CurrentPV,
			"max_pv":     npc.Stats.CalculateLifePoints(),
			"classe":     npc.Stats.Background,
			"lieu":       npc.Lieu,
			"alignement": npc.Alignement,
			"sorts":      npc.Sorts,
			"inventaire": npc.Inventaire,
			"equipement": npc.Equipement,
			"stats":      npc.Stats,
			"is_npc":     true,
		}
	}
	gm.broadcast(map[string]interface{}{
		"type":      "sync",
		"liste":     syncData,
		"npcs":      npcData,
		"locations": gm.World.Locations,
	})
}

func (gm *GameManager) Disconnect(pseudo string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.Connections, pseudo)
	delete(gm.DMs, pseudo)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🏃 %s a quitté le jeu.", pseudo),
	})
	gm.NotifyChange()
}
