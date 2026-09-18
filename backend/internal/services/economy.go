package services

import (
	"strings"

	"dnd-backend/internal/domain"
)

// playerChat broadcasts a player-authored chat message to the whole table.
func (gm *GameManager) playerChat(char *domain.Character, msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	gm.chat("💬 %s : %s", char.Stats.Nom, msg)
}

// dmChat broadcasts a message authored by the DM/GM.
func (gm *GameManager) dmChat(msg string) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return
	}
	gm.chat("📣 MDJ : %s", msg)
}

// actionSell trades an inventory item (by index) for its price in gold.
func (gm *GameManager) actionSell(char *domain.Character, index int) {
	if index < 0 || index >= len(char.Inventaire) {
		return
	}
	item := char.Inventaire[index]
	gain := item.Prix
	if gain <= 0 {
		gain = 1
	}
	char.Or += gain
	char.Inventaire = append(char.Inventaire[:index], char.Inventaire[index+1:]...)
	gm.chat("💰 %s a vendu %s pour %.0f pièce(s) d'or.", char.Stats.Nom, item.Nom, gain)
	gm.emit(GameEvent{Type: "event", Event: "gold", Source: char.Stats.Nom, Amount: gain, Text: "Vente de " + item.Nom})
}

// actionBuy purchases the item at the given index of the current location's
// shop stock, enforcing both funds and the character's carrying capacity.
func (gm *GameManager) actionBuy(char *domain.Character, index int) {
	loc := gm.findLocation(char.Lieu)
	if loc == nil || index < 0 || index >= len(loc.Commerce) {
		return
	}
	item := loc.Commerce[index]
	if item.Prix > char.Or {
		gm.chat("❌ %s n'a pas assez d'or pour acheter %s (%.0f or).", char.Stats.Nom, item.Nom, item.Prix)
		return
	}
	if char.CarriedWeight()+domain.ItemWeight(item) > char.Capacity() {
		gm.chat("❌ %s est trop chargé pour emporter %s.", char.Stats.Nom, item.Nom)
		return
	}
	char.Or -= item.Prix
	char.Inventaire = append(char.Inventaire, item)
	loc.Commerce = append(loc.Commerce[:index], loc.Commerce[index+1:]...)
	gm.chat("🛒 %s a acheté %s pour %.0f pièce(s) d'or.", char.Stats.Nom, item.Nom, item.Prix)
	gm.emit(GameEvent{Type: "event", Event: "buy", Target: char.Stats.Nom, Amount: item.Prix, Text: item.Nom})
	gm.persistWorld()
}

// dmEditOr allows the DM to set a character's gold directly.
func (gm *GameManager) dmEditOr(target string, or float64) {
	char := gm.getLatestCharacter(target)
	if char == nil {
		return
	}
	char.Or = or
	gm.chat("💰 Le MDJ a fixé l'or de %s à %.0f.", char.Stats.Nom, or)
}

// dmShopAdd adds an item to a location's shop stock.
func (gm *GameManager) dmShopAdd(location string, item domain.Item) {
	loc := gm.findLocation(location)
	if loc == nil || item.Nom == "" {
		return
	}
	loc.Commerce = append(loc.Commerce, item)
	gm.chat("🏪 Le MDJ a ajouté %s à l'échoppe de %s.", item.Nom, location)
	gm.persistWorld()
}