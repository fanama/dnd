package services

import (
	"encoding/json"

	"dnd-backend/internal/domain"
	"github.com/google/uuid"
)

func defaultLocations() []domain.Location {
	return []domain.Location{
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
	}
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

// loadOrSeedWorld restores the persisted NPCs and locations, falling back to
// the default seed the first time the database has no world state.
func (gm *GameManager) loadOrSeedWorld() {
	npcsJSON, locsJSON, err := gm.Repo.LoadWorld()
	if err == nil && npcsJSON != "" && locsJSON != "" {
		var npcs map[string]*domain.Character
		var locs []domain.Location
		if json.Unmarshal([]byte(npcsJSON), &npcs) == nil && npcs != nil &&
			json.Unmarshal([]byte(locsJSON), &locs) == nil && locs != nil {
			gm.World.NPCs = npcs
			gm.World.Locations = locs
			return
		}
	}
	gm.seedDefaultNPCs()
	gm.persistWorld()
}

// persistWorld snapshots the NPCs and locations and queues a database write.
func (gm *GameManager) persistWorld() {
	npcsJSON, _ := json.Marshal(gm.World.NPCs)
	locsJSON, _ := json.Marshal(gm.World.Locations)
	gm.persister.Enqueue(func() {
		gm.Repo.SaveWorld(string(npcsJSON), string(locsJSON))
	})
}
