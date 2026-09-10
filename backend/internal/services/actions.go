package services

import "dnd-backend/internal/domain"

// playerActionFn handles a player action on the given character.
type playerActionFn func(gm *GameManager, char *domain.Character, action Action)

// dmActionFn handles a DM-only action.
type dmActionFn func(gm *GameManager, action Action)

// registerActions wires the action identifiers to their implementations,
// replacing the previous switch-based dispatch in HandleAction.
func (gm *GameManager) registerActions() {
	gm.playerActions = map[string]playerActionFn{
		"attack":         func(gm *GameManager, c *domain.Character, a Action) { gm.actionAttack(c, a.Cible) },
		"move":           func(gm *GameManager, c *domain.Character, a Action) { gm.actionMove(c, a.Destination) },
		"cast_spell":     func(gm *GameManager, c *domain.Character, a Action) { gm.actionCastSpell(c, a.Sort, a.Cible) },
		"use_consumable": func(gm *GameManager, c *domain.Character, a Action) { gm.actionConsume(c, a.ItemName) },
		"loot":           func(gm *GameManager, c *domain.Character, a Action) { gm.actionLoot(c, a.LootName) },
		"equip_item":     func(gm *GameManager, c *domain.Character, a Action) { gm.actionEquip(c, a.ItemName) },
		"unequip_item":   func(gm *GameManager, c *domain.Character, a Action) { gm.actionUnequip(c, a.Slot) },
		"accept_quest":   func(gm *GameManager, c *domain.Character, a Action) { gm.actionAcceptQuest(c, a.QuestName) },
		"complete_quest": func(gm *GameManager, c *domain.Character, a Action) { gm.actionCompleteQuest(c, a.QuestName) },
		"create_character": func(gm *GameManager, c *domain.Character, a Action) {
			gm.createCharacter(c, a)
		},
	}

	gm.dmActions = map[string]dmActionFn{
		"dm_edit_stats":        func(gm *GameManager, a Action) { gm.dmEditStats(a.TargetPlayer, a.Stats) },
		"dm_set_pv":            func(gm *GameManager, a Action) { gm.dmSetPV(a.TargetPlayer, a.PV) },
		"dm_add_item":          func(gm *GameManager, a Action) { gm.dmAddItem(a.TargetPlayer, a.Item) },
		"dm_remove_item":       func(gm *GameManager, a Action) { gm.dmRemoveItem(a.TargetPlayer, a.ItemIndex) },
		"dm_add_spell":         func(gm *GameManager, a Action) { gm.dmAddSpell(a.TargetPlayer, a.Spell) },
		"dm_remove_spell":      func(gm *GameManager, a Action) { gm.dmRemoveSpell(a.TargetPlayer, a.ItemIndex) },
		"dm_move_player":       func(gm *GameManager, a Action) { gm.dmMovePlayer(a.TargetPlayer, a.Destination) },
		"dm_teleport_item_add": func(gm *GameManager, a Action) { gm.dmAddLocationItem(a.Destination, a.Item) },
		"dm_teleport_item_remove": func(gm *GameManager, a Action) {
			gm.dmRemoveLocationItem(a.Destination, a.ItemIndex)
		},
		"dm_delete_player":    func(gm *GameManager, a Action) { gm.dmDeletePlayer(a.TargetPlayer) },
		"dm_edit_align":       func(gm *GameManager, a Action) { gm.dmEditAlign(a.TargetPlayer, a.Alignement) },
		"dm_edit_location":    func(gm *GameManager, a Action) { gm.dmEditLocation(a.Destination, a.NewName, a.LocationBg) },
		"dm_add_npc":          func(gm *GameManager, a Action) { gm.dmAddNPC(a.ItemName, a.Destination, a.PV, a.Alignement, a.Stats, a.MobType, a.Count) },
		"dm_remove_npc":       func(gm *GameManager, a Action) { gm.dmRemoveNPC(a.ItemName) },
		"dm_remove_npcs":      func(gm *GameManager, a Action) { gm.dmRemoveNPCs(a.Names) },
		"dm_edit_npc":         func(gm *GameManager, a Action) { gm.dmEditNPC(a.ItemName, a.Stats, a.PV, a.Alignement, a.Overwrite) },
		"dm_move_npc":         func(gm *GameManager, a Action) { gm.dmMoveNPC(a.ItemName, a.Destination) },
		"dm_npc_add_item":     func(gm *GameManager, a Action) { gm.dmNPCAddItem(a.ItemName, a.Item) },
		"dm_npc_remove_item":  func(gm *GameManager, a Action) { gm.dmNPCRemoveItem(a.ItemName, a.ItemIndex) },
		"dm_npc_add_spell":    func(gm *GameManager, a Action) { gm.dmNPCAddSpell(a.ItemName, a.Spell) },
		"dm_npc_remove_spell": func(gm *GameManager, a Action) { gm.dmNPCRemoveSpell(a.ItemName, a.ItemIndex) },
		"dm_add_quest":        func(gm *GameManager, a Action) { gm.dmAddQuest(a.Destination, a.Quest) },
		"dm_remove_quest":     func(gm *GameManager, a Action) { gm.dmRemoveQuest(a.Destination, a.QuestIndex) },
		"dm_edit_quest":       func(gm *GameManager, a Action) { gm.dmEditQuest(a.Destination, a.QuestIndex, a.Quest) },
		"dm_export_state": func(gm *GameManager, a Action) {
			gm.dmExportState(a.Pseudo)
		},
		"dm_load_state": func(gm *GameManager, a Action) {
			gm.dmLoadState(a.Payload)
		},
		"dm_add_quest_player": func(gm *GameManager, a Action) { gm.dmAddPlayerQuest(a.TargetPlayer, a.Quest) },
		"dm_edit_quest_player": func(gm *GameManager, a Action) {
			gm.dmEditPlayerQuest(a.TargetPlayer, a.QuestIndex, a.Quest)
		},
		"dm_remove_quest_player": func(gm *GameManager, a Action) {
			gm.dmRemovePlayerQuest(a.TargetPlayer, a.QuestIndex)
		},
	}
}
