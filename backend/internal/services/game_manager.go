package services

import (
	"encoding/json"
	"fmt"
	"sync"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Action struct {
	Type        string `json:"type"`
	Cible       string `json:"cible,omitempty"`
	Destination string `json:"destination,omitempty"`
	Sort        string `json:"sort,omitempty"`
	ItemIndex   int    `json:"item_index,omitempty"`
	LootIndex   int    `json:"loot_index,omitempty"`
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
				{Nom: "Taverne", Background: "Ambiance chaleureuse", Objects: []domain.Item{{Nom: "Vieille Carte", IsConsumable: false, Prix: 10}}},
				{Nom: "Donjon", Background: "Sombre et humide", Objects: []domain.Item{{Nom: "Épée Rouillée", IsConsumable: false, BonusDégâts: 2, Prix: 40}, {Nom: "Potion de Soin", IsConsumable: true, Prix: 25}}},
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
	nom, classe, lieu, pv, maxPv, invStr, statsStr, err := gm.Repo.GetCharacter(pseudo)
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
		if charClass == "Magicien" {
			stats.Force = 8; stats.Constitution = 9; stats.Vitesse = 11; stats.Charisme = 12; stats.Instinct = 14; stats.Savoir = 16
		} else {
			stats.Force = 15; stats.Constitution = 12; stats.Vitesse = 10; stats.Charisme = 10; stats.Instinct = 10; stats.Savoir = 10
		}

		char := &domain.Character{
			ID:         uuid.New(),
			Alignement: "Neutre",
			Stats:      stats,
			CurrentPV:  stats.CalculateLifePoints(),
			Lieu:       "Taverne",
		}

		if charClass == "Magicien" {
			char.Sorts = append(char.Sorts, domain.Sort{Nom: "Boule de Feu", EcoleMagie: "Évocations"})
		}
		char.Inventaire = append(char.Inventaire, domain.Item{Nom: "Épée Longue", IsConsumable: false, BonusDégâts: 5, Prix: 100})
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
	case "use_consumable": gm.actionConsume(char, action.ItemIndex)
	case "loot": gm.actionLoot(char, action.LootIndex)
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

func (gm *GameManager) actionConsume(char *domain.Character, index int) {
	if index < 0 || index >= len(char.Inventaire) { return }
	item := char.Inventaire[index]
	if item.IsConsumable {
		heal := char.Stats.Constitution * 5
		char.CurrentPV += heal
		if char.CurrentPV > char.Stats.CalculateLifePoints() {
			char.CurrentPV = char.Stats.CalculateLifePoints()
		}
		char.Inventaire = append(char.Inventaire[:index], char.Inventaire[index+1:]...)
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

func (gm *GameManager) actionLoot(char *domain.Character, index int) {
	var currentLocation *domain.Location
	for i := range gm.World.Locations {
		if gm.World.Locations[i].Nom == char.Lieu {
			currentLocation = &gm.World.Locations[i]
			break
		}
	}
	if currentLocation == nil || index < 0 || index >= len(currentLocation.Objects) { return }
	item := currentLocation.Objects[index]
	char.Inventaire = append(char.Inventaire, item)
	currentLocation.Objects = append(currentLocation.Objects[:index], currentLocation.Objects[index+1:]...)
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
