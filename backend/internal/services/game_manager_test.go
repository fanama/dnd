package services

import (
	"strings"
	"testing"

	"dnd-backend/internal/domain"
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

	// Boss: never duplicated, default preset stats + 300 PV.
	gm.dmAddNPC("Roi Goblin", "Donjon", 0, "", domain.Stats{}, "boss", 3)
	boss, ok := gm.World.NPCs["Roi Goblin"]
	if !ok {
		t.Fatalf("boss was not created")
	}
	if boss.MobType != "boss" || boss.Stats.Force != 18 || boss.CurrentPV != 300 {
		t.Errorf("boss = mobType:%q force:%v pv:%v, want mobType:boss force:18 pv:300", boss.MobType, boss.Stats.Force, boss.CurrentPV)
	}

	// Minions: 3 copies with suffixed names, default preset stats + 50 PV.
	gm.dmAddNPC("Goblin", "Donjon", 0, "", domain.Stats{}, "minion", 3)
	for _, name := range []string{"Goblin #1", "Goblin #2", "Goblin #3"} {
		npc, ok := gm.World.NPCs[name]
		if !ok {
			t.Fatalf("minion %q was not created", name)
		}
		if npc.MobType != "minion" || npc.CurrentPV != 50 {
			t.Errorf("%s = mobType:%q pv:%v, want mobType:minion pv:50", name, npc.MobType, npc.CurrentPV)
		}
	}

	// Regular PNJ keeps its previous behavior (Constitution * 10 PV, no mob type).
	gm.dmAddNPC("Boby", "Taverne", 0, "", domain.Stats{}, "", 1)
	pnj := gm.World.NPCs["Boby"]
	if pnj == nil || pnj.MobType != "" || pnj.CurrentPV != 100 {
		t.Errorf("PNJ Boby = %+v, want mobType empty and 100 PV", pnj)
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

func failIf(t *testing.T, msg, contains string) {
	t.Helper()
	if !strings.Contains(msg, contains) {
		t.Errorf("message %q should contain %q", msg, contains)
	}
}
