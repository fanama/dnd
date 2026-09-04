package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
	Type        string `json:"type"`
	Cible       string `json:"cible,omitempty"`
	Destination string `json:"destination,omitempty"`
	Sort        string `json:"sort,omitempty"`
	ItemIndex   int    `json:"item_index,omitempty"`
	ItemName    string `json:"item_name,omitempty"`
	LootName    string `json:"loot_name,omitempty"`
}

type GameManager struct {
	Connections map[string]*websocket.Conn
	World       *domain.World
	Repo        *repository.SQLiteRepository
	mu          sync.Mutex
}

func NewGameManager(repo *repository.SQLiteRepository) *GameManager {
	gm := &GameManager{
		Connections: make(map[string]*websocket.Conn),
		Repo:        repo,
		World: &domain.World{
			ID:      uuid.New(),
			Players: make(map[string]*domain.Player),
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
	return gm
}

func (gm *GameManager) Connect(pseudo string, ws *websocket.Conn, charInfo map[string]string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gm.Connections[pseudo] = ws

	player, ok := gm.World.Players[pseudo]
	if !ok {
		player = &domain.Player{Pseudo: pseudo}
		gm.World.Players[pseudo] = player
	}

	// 1. Attempt to restore from DB
	_, _, lieu, pv, _, invStr, statsStr, err := gm.Repo.GetCharacter(pseudo)
	if err == nil {
		var stats domain.Stats
		json.Unmarshal([]byte(statsStr), &stats)
		var inventory []domain.Item
		json.Unmarshal([]byte(invStr), &inventory)

		char := &domain.Character{
			ID:         uuid.New(),
			Alignement: "Neutre",
			Stats:      stats,
			CurrentPV:  float64(pv),
			Lieu:       lieu,
			Inventaire: inventory,
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
			stats.Force = 8; stats.Constitution = 9; stats.Vitesse = 11; stats.Charisme = 12; stats.Instinct = 14; stats.Savoir = 16
		case "Voleur":
			stats.Force = 10; stats.Constitution = 8; stats.Vitesse = 16; stats.Charisme = 10; stats.Instinct = 15; stats.Savoir = 11
		case "Clerc":
			stats.Force = 12; stats.Constitution = 14; stats.Vitesse = 8; stats.Charisme = 15; stats.Instinct = 9; stats.Savoir = 12
		case "Barde":
			stats.Force = 9; stats.Constitution = 10; stats.Vitesse = 13; stats.Charisme = 16; stats.Instinct = 12; stats.Savoir = 10
		case "Ranger":
			stats.Force = 13; stats.Constitution = 11; stats.Vitesse = 14; stats.Charisme = 8; stats.Instinct = 15; stats.Savoir = 9
		default: // Guerrier
			stats.Force = 15; stats.Constitution = 12; stats.Vitesse = 10; stats.Charisme = 10; stats.Instinct = 10; stats.Savoir = 10
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

	char := gm.getLatestCharacter(pseudo)
	if char == nil { return }

	switch action.Type {
	case "attack": gm.actionAttack(char, action.Cible)
	case "move": gm.actionMove(char, action.Destination)
	case "cast_spell": gm.actionCastSpell(char, action.Sort, action.Cible)
	case "use_consumable": gm.actionConsume(char, action.ItemName)
	case "loot": gm.actionLoot(char, action.LootName)
	}

	gm.saveCharacterState(pseudo, char)
	gm.NotifyChange()
}

func (gm *GameManager) saveCharacterState(pseudo string, char *domain.Character) {
	statsJson, _ := json.Marshal(char.Stats)
	invJson, _ := json.Marshal(char.Inventaire)
	gm.Repo.SaveCharacter(pseudo, char.Stats.Nom, char.Stats.Background, char.Lieu, int(char.CurrentPV), int(char.Stats.CalculateLifePoints()), string(invJson), string(statsJson))
}

func (gm *GameManager) actionAttack(attacker *domain.Character, targetPseudo string) {
	targetPlayer, ok := gm.World.Players[targetPseudo]
	if !ok || len(targetPlayer.Characters) == 0 { return }
	target := targetPlayer.Characters[len(targetPlayer.Characters)-1]
	if attacker.Lieu != target.Lieu { return }

	damage := attacker.Stats.CalculateDamage() - target.Stats.CalculateArmor()
	if damage < 1 { damage = 1 }
	target.CurrentPV -= damage

	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("💥 %s attaque %s ! Dégâts : %.0f. (PV : %.0f)", attacker.Stats.Nom, target.Stats.Nom, damage, target.CurrentPV),
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
	if loc == nil { return }
	if len(loc.Objects) >= 5 { return }

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
	targetPlayer, ok := gm.World.Players[targetPseudo]
	if !ok || len(targetPlayer.Characters) == 0 { return }
	target := targetPlayer.Characters[len(targetPlayer.Characters)-1]
	damage := char.Stats.Savoir * 1.5
	target.CurrentPV -= damage
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🔥 %s lance %s sur %s ! Dégâts Magiques : %.0f.", char.Stats.Nom, spellName, target.Stats.Nom, damage),
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
	if itemIdx == -1 { return }
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
	if !ok || len(player.Characters) == 0 { return nil }
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
	if currentLocation == nil || itemName == "" { return }
	itemIdx := -1
	for i, obj := range currentLocation.Objects {
		if obj.Nom == itemName {
			itemIdx = i
			break
		}
	}
	if itemIdx == -1 { return }
	item := currentLocation.Objects[itemIdx]
	char.Inventaire = append(char.Inventaire, item)
	currentLocation.Objects = append(currentLocation.Objects[:itemIdx], currentLocation.Objects[itemIdx+1:]...)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🎒 %s a ramassé %s dans %s !", char.Stats.Nom, item.Nom, char.Lieu),
	})
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
			sorts := []string{}
			for _, s := range char.Sorts { sorts = append(sorts, s.Nom) }
			inv := []string{}
			for _, i := range char.Inventaire { inv = append(inv, i.Nom) }
			syncData[pseudo] = map[string]interface{}{
				"nom":        char.Stats.Nom,
				"pv":         char.CurrentPV,
				"max_pv":     char.Stats.CalculateLifePoints(),
				"classe":     char.Stats.Background,
				"lieu":       char.Lieu,
				"sorts":      sorts,
				"inventaire": inv,
				"stats":       char.Stats,
			}
		}
	}
	gm.broadcast(map[string]interface{}{
		"type": "sync",
		"liste": syncData,
		"locations": gm.World.Locations,
	})
}

func (gm *GameManager) Disconnect(pseudo string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.Connections, pseudo)
	gm.broadcast(map[string]interface{}{
		"type": "chat",
		"msg":  fmt.Sprintf("🏃 %s a quitté le jeu.", pseudo),
	})
	gm.NotifyChange()
}
