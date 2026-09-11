package services

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"dnd-backend/internal/domain"
	"dnd-backend/internal/repository"
	"github.com/google/uuid"
)

func newTestCharacter(name string, stats domain.Stats, pv float64) *domain.Character {
	return &domain.Character{
		ID:         uuid.New(),
		Stats:      stats,
		CurrentPV:  pv,
		Inventaire: []domain.Item{},
	}
}

func TestAbilityModifier(t *testing.T) {
	cases := []struct {
		stat   float64
		expMod float64
	}{
		{8, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{18, 4},
		{7, -2},
	}
	for _, c := range cases {
		if got := domain.AbilityModifier(c.stat); got != c.expMod {
			t.Errorf("AbilityModifier(%v) = %v, want %v", c.stat, got, c.expMod)
		}
	}
}

func TestBaseAC(t *testing.T) {
	stats := domain.Stats{Vitesse: 14}
	if got, want := stats.BaseAC(), 12.0; got != want {
		t.Errorf("BaseAC = %v, want %v", got, want)
	}
}

func TestWeaponInfo(t *testing.T) {
	caseNameFn := func(name string, item *domain.Item, sides int, ranged bool) {
		sidesGot, rangedGot := weaponInfo(item)
		if sidesGot != sides || rangedGot != ranged {
			t.Errorf("%s: weaponInfo = (%d, %v), want (%d, %v)", name, sidesGot, rangedGot, sides, ranged)
		}
	}
	caseNameFn("nil = mains nues", nil, 2, false)
	caseNameFn("Dague", &domain.Item{Nom: "Dague"}, 4, false)
	caseNameFn("Épée", &domain.Item{Nom: "Épée"}, 6, false)
	caseNameFn("Arc", &domain.Item{Nom: "Arc"}, 8, true)
	caseNameFn("Arbalète", &domain.Item{Nom: "Arbalète"}, 8, true)
	caseNameFn("Hache", &domain.Item{Nom: "Hache de Guerre"}, 6, false)
}

func TestWeaponInfoExplicitDiceOverridesName(t *testing.T) {
	sides, ranged := weaponInfo(&domain.Item{Nom: "Dague", DesDégâts: "d10"})
	if sides != 10 || ranged {
		t.Errorf("weaponInfo = (%d, %v), want (10, false)", sides, ranged)
	}
	sides, ranged = weaponInfo(&domain.Item{Nom: "Arc Long", DesDégâts: "1d6"})
	if sides != 6 || !ranged {
		t.Errorf("weaponInfo = (%d, %v), want (6, true)", sides, ranged)
	}
	sides, ranged = weaponInfo(&domain.Item{Nom: "Marteau", DesDégâts: "bogus"})
	if sides != 6 || ranged {
		t.Errorf("invalid dice should fall back to name default, got (%d, %v)", sides, ranged)
	}
}

func TestResolvePhysicalAttackMissOnLowRoll(t *testing.T) {
	oldRoll := rollD20Fn
	rollD20Fn = func() int { return 5 }
	defer func() { rollD20Fn = oldRoll }()

	attacker := newTestCharacter("Aragorn", domain.Stats{Force: 15, Vitesse: 10}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 10}, 50)
	target.Stats.Nom = "Goblin"
	attacker.Stats.Nom = "Aragorn"

	msg := resolvePhysicalAttack(attacker, target)
	if target.CurrentPV != 50 {
		t.Errorf("PV should be unchanged on miss, got %v", target.CurrentPV)
	}
	failIf(t, msg, "Raté")
}

func TestResolvePhysicalAttackHitOnACReached(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 18 }
	rollDiceFn = func(count, sides int) int { return 1 * count }
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	// Force 15 → mod +2 ; target AC 10 → roll 18+2 = 20 ≥ 10 → hit, 1d6+2 = 3
	attacker := newTestCharacter("Aragorn", domain.Stats{Force: 15, Vitesse: 10}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 10}, 50)
	attacker.Stats.Nom = "Aragorn"
	target.Stats.Nom = "Goblin"

	msg := resolvePhysicalAttack(attacker, target)
	if target.CurrentPV != 47 {
		t.Errorf("PV = %v, want 47 (3 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "Touché")
}

func TestResolvePhysicalAttackCriticalNatural20(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 20 }
	rollDiceFn = func(count, sides int) int { return sides * count } // max damage
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	// Unarmed 2d2 + mod(Force) : Force 18 → +4 → damage 8 for 2d2 (each d2=2)
	attacker := newTestCharacter("Aragorn", domain.Stats{Force: 18, Vitesse: 10}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 18}, 50)
	attacker.Stats.Nom = "Aragorn"
	target.Stats.Nom = "Goblin"

	msg := resolvePhysicalAttack(attacker, target)
	// Even against AC 14, nat 20 always hits. Damage = 2d2(4) + 4 = 8.
	if target.CurrentPV != 42 {
		t.Errorf("PV = %v, want 42 (8 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "CRITIQUE")
}

func TestResolvePhysicalAttackNat1AlwaysMisses(t *testing.T) {
	oldRoll := rollD20Fn
	rollD20Fn = func() int { return 1 }
	defer func() { rollD20Fn = oldRoll }()

	attacker := newTestCharacter("Aragorn", domain.Stats{Force: 20, Vitesse: 20}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 8}, 50)
	attacker.Stats.Nom = "Aragorn"
	target.Stats.Nom = "Goblin"

	msg := resolvePhysicalAttack(attacker, target)
	if target.CurrentPV != 50 {
		t.Errorf("PV = %v, want 50 (nat 1 always misses)", target.CurrentPV)
	}
	failIf(t, msg, "Raté")
}

func TestResolvePhysicalAttackMinDamageOne(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 19 }
	rollDiceFn = func(count, sides int) int { return 1 * count }
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	// Force 8 → mod -1 ; unarmed 1d2 rolls 1 → 1 - 1 = 0 → clamped to 1.
	attacker := newTestCharacter("Faible", domain.Stats{Force: 8, Vitesse: 10}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 8}, 50)
	attacker.Stats.Nom = "Faible"
	target.Stats.Nom = "Goblin"

	msg := resolvePhysicalAttack(attacker, target)
	if target.CurrentPV != 49 {
		t.Errorf("PV = %v, want 49 (min 1 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "Touché")
}

func TestResolveSpellAttackMiss(t *testing.T) {
	oldRoll := rollD20Fn
	rollD20Fn = func() int { return 2 }
	defer func() { rollD20Fn = oldRoll }()

	attacker := newTestCharacter("Magicien", domain.Stats{Savoir: 8}, 100) // mod -1 → total 1 < AC 10
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 10}, 50)
	attacker.Stats.Nom = "Magicien"
	target.Stats.Nom = "Goblin"

	msg := resolveSpellAttack(attacker, target, domain.Sort{Nom: "Boule de Feu"})
	if target.CurrentPV != 50 {
		t.Errorf("PV = %v, want 50", target.CurrentPV)
	}
	failIf(t, msg, "Raté")
}

func TestResolveSpellAttackCritDoubleDamage(t *testing.T) {
	oldRoll := rollD20Fn
	rollD20Fn = func() int { return 20 }
	defer func() { rollD20Fn = oldRoll }()

	attacker := newTestCharacter("Magicien", domain.Stats{Savoir: 16}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 20}, 50)
	attacker.Stats.Nom = "Magicien"
	target.Stats.Nom = "Goblin"

	msg := resolveSpellAttack(attacker, target, domain.Sort{Nom: "Boule de Feu"})
	// normal = 24, crit = 48
	if target.CurrentPV != 2 {
		t.Errorf("PV = %v, want 2 (48 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "CRITIQUE")
}

func TestResolveSpellAttackBonusAndDice(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 10 }
	rollDiceFn = func(count, sides int) int { return sides * count } // max damage
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	// Savoir 8 → mod -1 ; bonus +4 → total 13 ≥ CA 10 (Vitesse 10) → touche
	attacker := newTestCharacter("Magicien", domain.Stats{Savoir: 8}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 10}, 50)
	attacker.Stats.Nom = "Magicien"
	target.Stats.Nom = "Goblin"

	msg := resolveSpellAttack(attacker, target, domain.Sort{
		Nom:       "Éclair de Givre",
		Bonus:     4,
		DesDégâts: "d8",
	})
	// 1d8 max = 8
	if target.CurrentPV != 42 {
		t.Errorf("PV = %v, want 42 (8 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "Touché")
	failIf(t, msg, "[1d8")
}

func TestResolveSpellAttackDiceCritTwoDice(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 20 }
	rollDiceFn = func(count, sides int) int { return sides * count } // max damage
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	attacker := newTestCharacter("Magicien", domain.Stats{Savoir: 10}, 100)
	target := newTestCharacter("Goblin", domain.Stats{Vitesse: 20}, 50)
	attacker.Stats.Nom = "Magicien"
	target.Stats.Nom = "Goblin"

	msg := resolveSpellAttack(attacker, target, domain.Sort{
		Nom:       "Boule de Feu",
		DesDégâts: "d6",
	})
	// crit → 2d6 max = 12
	if target.CurrentPV != 38 {
		t.Errorf("PV = %v, want 38 (12 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "[2d6")
}

func TestApplyStatsBuff(t *testing.T) {
	stats := domain.Stats{Force: 10, Savoir: 12}
	ok, label := applyStatsBuff(&stats, domain.SortBuff{Stat: "force", Valeur: 5})
	if !ok || label != "Force" || stats.Force != 15 {
		t.Errorf("buff force = ok:%v label:%q force:%v, want ok:true label:Force force:15", ok, label, stats.Force)
	}

	ok, _ = applyStatsBuff(&stats, domain.SortBuff{Stat: "inconnu", Valeur: 5})
	if ok {
		t.Errorf("applyStatsBuff should reject unknown stat")
	}
}

func TestActionCastSpellOnNPC(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 17 }
	rollDiceFn = func(count, sides int) int { return sides * count } // max damage
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Savoir: 10}, 100)},
	}
	attacker := gm.World.Players["arya"].Characters[0]
	attacker.Stats.Nom = "Arya"
	attacker.Sorts = []domain.Sort{{Nom: "Boule de Feu", DesDégâts: "d6"}}

	npc := gm.World.NPCs["Grum le Garde"]
	npc.CurrentPV = 50
	attacker.Lieu = npc.Lieu

	gm.actionCastSpell(attacker, "Boule de Feu", "Grum le Garde")
	// 1d6 max = 6 dégâts → 44
	if npc.CurrentPV != 44 {
		t.Errorf("PV du PNJ = %v, want 44", npc.CurrentPV)
	}
}

func TestActionCastSpellBuffOnSelf(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Nom: "Arya", Force: 10}, 100)},
	}
	attacker := gm.World.Players["arya"].Characters[0]
	attacker.Sorts = []domain.Sort{{Nom: "Bénédiction", Buff: &domain.SortBuff{Stat: "Force", Valeur: 2}}}

	forceBefore := attacker.Stats.Force
	gm.actionCastSpell(attacker, "Bénédiction", "arya")
	if attacker.Stats.Force != forceBefore+2 {
		t.Errorf("Force = %v, want %v", attacker.Stats.Force, forceBefore+2)
	}
}

func TestActionCastSpellNPCDeathClampedToZero(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 16 }
	rollDiceFn = func(count, sides int) int { return 100 } // massive damage
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Savoir: 10}, 100)},
	}
	attacker := gm.World.Players["arya"].Characters[0]
	attacker.Stats.Nom = "Arya"
	attacker.Sorts = []domain.Sort{{Nom: "Éclair", DesDégâts: "d12"}}

	npc := gm.World.NPCs["Zorra la Sorciere"]
	npc.CurrentPV = 5
	attacker.Lieu = npc.Lieu

	gm.actionCastSpell(attacker, "Éclair", "Zorra la Sorciere")
	if npc.CurrentPV != 0 {
		t.Errorf("PV du PNJ = %v, want 0 (clamp)", npc.CurrentPV)
	}
}

func TestDmAddMobSpawnsBossAndMinions(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	// Boss: never duplicated, default preset stats + 30 PV.
	gm.dmAddNPC("Roi Goblin", "Donjon", 0, "", domain.Stats{}, "boss", 3)
	boss, ok := gm.World.NPCs["Roi Goblin"]
	if !ok {
		t.Fatalf("boss was not created")
	}
	if boss.MobType != "boss" || boss.Stats.Force != 18 || boss.CurrentPV != 30 {
		t.Errorf("boss = mobType:%q force:%v pv:%v, want mobType:boss force:18 pv:30", boss.MobType, boss.Stats.Force, boss.CurrentPV)
	}
	if boss.MaxPV != 30 {
		t.Errorf("boss MaxPV = %v, want 30 (normalisé)", boss.MaxPV)
	}

	// Minions: 3 copies with suffixed names, default preset stats + 8 PV.
	gm.dmAddNPC("Goblin", "Donjon", 0, "", domain.Stats{}, "minion", 3)
	for _, name := range []string{"Goblin #1", "Goblin #2", "Goblin #3"} {
		npc, ok := gm.World.NPCs[name]
		if !ok {
			t.Fatalf("minion %q was not created", name)
		}
		if npc.MobType != "minion" || npc.CurrentPV != 8 {
			t.Errorf("%s = mobType:%q pv:%v, want mobType:minion pv:8", name, npc.MobType, npc.CurrentPV)
		}
		if npc.MaxPV != 8 {
			t.Errorf("%s MaxPV = %v, want 8 (normalisé)", name, npc.MaxPV)
		}
	}

	// Regular PNJ keeps its previous behavior (D&D HD formula PV, no mob type).
	gm.dmAddNPC("Boby", "Taverne", 0, "", domain.Stats{}, "", 1)
	pnj := gm.World.NPCs["Boby"]
	if pnj == nil || pnj.MobType != "" || pnj.CurrentPV != 10 {
		t.Errorf("PNJ Boby = %+v, want mobType empty and 10 PV (d10 + CON mod)", pnj)
	}
	if pnj.MaxPV != 10 {
		t.Errorf("PNJ Boby MaxPV = %v, want 10 (normalisé)", pnj.MaxPV)
	}
}

func TestActionAttackOnNPCWithWeapon(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 17 }
	rollDiceFn = func(count, sides int) int { return sides * count }
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Force: 10}, 100)},
	}
	attacker := gm.World.Players["arya"].Characters[0]
	attacker.Stats.Nom = "Arya"
	attacker.Lieu = "Donjon"
	attacker.Equipement.Arme = &domain.Item{Nom: "Épée", DesDégâts: "d8"}

	npc := gm.World.NPCs["Grum le Garde"]
	npc.CurrentPV = 50

	gm.actionAttack(attacker, "Grum le Garde")
	// hit → 1d8 max = 8, mod Force 10 = 0, pas de bonus magique → PV 42
	if npc.CurrentPV != 42 {
		t.Errorf("PV du PNJ = %v, want 42", npc.CurrentPV)
	}
}

func TestActionAttackKillsMOBDropsInventoryAsLoot(t *testing.T) {
	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 16 }
	rollDiceFn = func(count, sides int) int { return sides * count }
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	gm := newTestGameManager(t)
	defer gm.Close()

	gm.World.Players["arya"] = &domain.Player{
		Pseudo:     "arya",
		Characters: []*domain.Character{newTestCharacter("Arya", domain.Stats{Force: 10}, 100)},
	}
	attacker := gm.World.Players["arya"].Characters[0]
	attacker.Stats.Nom = "Arya"
	attacker.Lieu = "Donjon"
	attacker.Equipement.Arme = &domain.Item{Nom: "Marteau", DesDégâts: "d12", BonusDégâts: 5}

	gm.dmAddNPC("Rat Sinistre", "Donjon", 3, "Chaotique Mauvais", domain.Stats{}, "minion", 1)
	mob := gm.World.NPCs["Rat Sinistre"]
	mob.Inventaire = []domain.Item{{Nom: "Morceau de Viande", Prix: 5}}

	gm.actionAttack(attacker, "Rat Sinistre")
	// 1d12 max = 12 + 0 + 5 = 17 ≥ 3 PV → mort
	if mob.CurrentPV != 0 {
		t.Errorf("PV du MOB = %v, want 0 (mort)", mob.CurrentPV)
	}
	if len(mob.Inventaire) != 0 {
		t.Errorf("MOB mort garde son inventaire : %v", mob.Inventaire)
	}
	if _, still := gm.World.NPCs["Rat Sinistre"]; still {
		t.Error("MOB mort devrait être retiré du monde")
	}
	loc := gm.findLocation("Donjon")
	if loc == nil || !gm.hasItem(loc.Objects, "Morceau de Viande") {
		t.Errorf("le butin devrait être au sol à Donjon : %+v", loc)
	}
}

func TestDmRemoveNPCsBulk(t *testing.T) {
	gm := newTestGameManager(t)
	defer gm.Close()

	gm.dmAddNPC("Roi Goblin", "Donjon", 0, "", domain.Stats{}, "boss", 1)
	gm.dmAddNPC("Goblin", "Donjon", 0, "", domain.Stats{}, "minion", 3)

	gm.dmRemoveNPCs([]string{"Goblin #1", "Goblin #3", "Roi Goblin", "Inexistant"})

	for _, gone := range []string{"Goblin #1", "Goblin #3", "Roi Goblin"} {
		if _, ok := gm.World.NPCs[gone]; ok {
			t.Errorf("%q devrait être supprimé", gone)
		}
	}
	if _, ok := gm.World.NPCs["Goblin #2"]; !ok {
		t.Errorf("Goblin #2 devrait rester (non sélectionné)")
	}
}

func TestAttackDamagePersistedOnTarget(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "game.db")
	repo1, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	gm := NewGameManager(repo1)

	ceasar := newTestCharacter("César", domain.Stats{Force: 15, Vitesse: 10}, 100)
	ceasar.Stats.Nom = "César"
	ceasar.Lieu = "Taverne"
	brutus := newTestCharacter("Brutus", domain.Stats{Vitesse: 10}, 50)
	brutus.Stats.Nom = "Brutus"
	brutus.Lieu = "Taverne"

	gm.World.Players["ceasar"] = &domain.Player{Pseudo: "ceasar", Characters: []*domain.Character{ceasar}}
	gm.World.Players["brutus"] = &domain.Player{Pseudo: "brutus", Characters: []*domain.Character{brutus}}

	gm.saveCharacterState("ceasar", ceasar)
	gm.saveCharacterState("brutus", brutus)

	ceasar.Equipement.Arme = &domain.Item{Nom: "Épée Longue", BonusDégâts: 3}

	oldRoll := rollD20Fn
	oldDice := rollDiceFn
	rollD20Fn = func() int { return 18 }
	rollDiceFn = func(count, sides int) int { return 1 * count }
	defer func() { rollD20Fn = oldRoll; rollDiceFn = oldDice }()

	gm.HandleAction("ceasar", Action{Type: "attack", Cible: "brutus"})
	if brutus.CurrentPV >= 50 {
		t.Fatalf("target should have taken damage, PV=%v", brutus.CurrentPV)
	}
	gm.Close()

	repo2, err := repository.NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer repo2.Close()
	_, _, _, pv, _, _, _, _, _, err := repo2.GetCharacter("brutus")
	if err != nil {
		t.Fatalf("target character missing from DB: %v", err)
	}
	if int64(pv) >= 50 {
		t.Errorf("damage not persisted for target: reloaded PV = %d (want < 50)", pv)
	}
}

func TestComputeDerivedCombat(t *testing.T) {
	mêlée := newTestCharacter("Zelda", domain.Stats{Force: 15, Vitesse: 10, Constitution: 12}, 10)
	mêlée.Stats.Nom = "Zelda"
	mêlée.Stats.Background = "Guerrier"
	mêlée.Equipement.Arme = &domain.Item{Nom: "Épée Longue", BonusDégâts: 3}
	mêlée.Equipement.Armure = &domain.Item{Nom: "Plastron", BonusArmure: 2}

	c := computeDerivedCombat(mêlée)
	if c.AC != 12 {
		t.Errorf("AC = %v, want 12 (BaseAC(10) + armure 2)", c.AC)
	}
	if c.AttackMod != 5 {
		t.Errorf("AttackMod = %v, want 5 (+2 force, +3 arme)", c.AttackMod)
	}
	if c.DamageMod != 2 {
		t.Errorf("DamageMod = %v, want 2", c.DamageMod)
	}
	if c.DamageDice != "1d6+2" {
		t.Errorf("DamageDice = %q, want %q", c.DamageDice, "1d6+2")
	}
	if c.HitDice != 10 {
		t.Errorf("HitDice = %v, want 10 (Guerrier)", c.HitDice)
	}
	if c.WeaponName != "Épée Longue" || c.Ranged {
		t.Errorf("WeaponName/Ranged = (%q, %v), want (%q, false)", c.WeaponName, c.Ranged, "Épée Longue")
	}

	archer := newTestCharacter("Legolas", domain.Stats{Force: 8, Vitesse: 16, Constitution: 12, Background: "Ranger"}, 10)
	archer.Stats.Nom = "Legolas"
	archer.Equipement.Arme = &domain.Item{Nom: "Arc Long"}

	c2 := computeDerivedCombat(archer)
	if c2.Ranged != true {
		t.Errorf("Arc should be ranged")
	}
	if c2.AttackMod != 3 {
		t.Errorf("Archer AttackMod = %v, want 3 (+3 vitesse)", c2.AttackMod)
	}
	if c2.DamageDice != "1d8+3" {
		t.Errorf("Archer DamageDice = %q, want %q", c2.DamageDice, "1d8+3")
	}
	if c2.HitDice != 8 {
		t.Errorf("Archer HitDice = %v, want 8 (Ranger)", c2.HitDice)
	}

	nue := newTestCharacter("Bare", domain.Stats{Force: 10, Vitesse: 10, Constitution: 10}, 10)
	nue.Stats.Nom = "Bare"
	c3 := computeDerivedCombat(nue)
	if c3.WeaponName != "Mains nues" || c3.DamageDice != "1d2+0" || c3.AC != 10 {
		t.Errorf("bare hands combat = %+v", c3)
	}
}

func TestNpcEntrySyncCarriesMobTypeAndNormalizedMaxPV(t *testing.T) {
	npc := &domain.Character{
		Stats:     domain.Stats{Nom: "Roi Goblin", Background: "Boss", Constitution: 16},
		CurrentPV: 30,
		MaxPV:     30,
		MobType:   "boss",
	}
	entry := NPCEntry{
		Nom:     npc.Stats.Nom,
		PV:      npc.CurrentPV,
		MaxPV:   npcMaxPV(npc),
		Stats:   npc.Stats,
		Combat:  computeDerivedCombat(npc),
		MobType: npc.MobType,
		IsNPC:   true,
	}
	if entry.MaxPV != 30 {
		t.Errorf("MaxPV = %v, want 30 (normalisé, pas 13 dérivé)", entry.MaxPV)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"mobType":"boss"`) {
		t.Errorf("sync JSON should carry mobType for the DM frontend: %s", data)
	}
}

func failIf(t *testing.T, msg, contains string) {
	t.Helper()
	if !strings.Contains(msg, contains) {
		t.Errorf("message %q should contain %q", msg, contains)
	}
}
