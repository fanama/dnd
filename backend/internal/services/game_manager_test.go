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
	cases := []struct {
		name   string
		item   *domain.Item
		sides  int
		ranged bool
	}{
		{name: "nil = mains nues", item: nil, sides: 2, ranged: false},
		{name: "Dague", item: &domain.Item{Nom: "Dague"}, sides: 4, ranged: false},
		{name: "Épée", item: &domain.Item{Nom: "Épée"}, sides: 6, ranged: false},
		{name: "Arc", item: &domain.Item{Nom: "Arc"}, sides: 8, ranged: true},
		{name: "Arbalète", item: &domain.Item{Nom: "Arbalète"}, sides: 8, ranged: true},
		{name: "Hache", item: &domain.Item{Nom: "Hache de Guerre"}, sides: 6, ranged: false},
	}
	for _, c := range cases {
		sides, ranged := weaponInfo(c.item)
		if sides != c.sides || ranged != c.ranged {
			t.Errorf("%s: weaponInfo = (%d, %v), want (%d, %v)", c.name, sides, ranged, c.sides, c.ranged)
		}
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

	msg := resolveSpellAttack(attacker, target, "Boule de Feu")
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

	msg := resolveSpellAttack(attacker, target, "Boule de Feu")
	// normal = 24, crit = 48
	if target.CurrentPV != 2 {
		t.Errorf("PV = %v, want 2 (48 dmg)", target.CurrentPV)
	}
	failIf(t, msg, "CRITIQUE")
}

func failIf(t *testing.T, msg, contains string) {
	t.Helper()
	if !strings.Contains(msg, contains) {
		t.Errorf("message %q should contain %q", msg, contains)
	}
}
