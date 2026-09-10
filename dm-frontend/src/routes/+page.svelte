<script>
    import '../app.css';
    import { dmState, dmConnect, dmSend, selectPlayer } from '../lib/stores/dm';
    import PlayerCard from '../lib/components/PlayerCard.svelte';
    import StatsEditor from '../lib/components/StatsEditor.svelte';
    import InventoryEditor from '../lib/components/InventoryEditor.svelte';
    import SpellEditor from '../lib/components/SpellEditor.svelte';
    import QuestEditor from '../lib/components/QuestEditor.svelte';
    import LocationManager from '../lib/components/LocationManager.svelte';
    import NPCManager from '../lib/components/NPCManager.svelte';
    import ChatBox from '../lib/components/ChatBox.svelte';

    let pseudo = 'dm';

    function login() {
        if (pseudo.trim()) dmConnect(pseudo.trim());
    }

    function dmAction(type, extra = {}) {
        dmSend({ type, ...extra });
    }

    function deletePlayer(target) {
        if (confirm(`Supprimer ${target} du monde ?`)) {
            dmAction('dm_delete_player', { target_player: target });
            selectPlayer(null);
        }
    }

    function addNpc(npc) {
        dmAction('dm_add_npc', {
            item_name: npc.nom,
            destination: npc.lieu,
            pv: npc.pv,
            alignement: npc.alignement,
            mob_type: npc.type || 'pnj',
            count: npc.count || 1,
        });
    }

    function editNpc(name, partialStats, pv, alignement) {
        dmAction('dm_edit_npc', {
            item_name: name,
            stats: { ...partialStats },
            pv,
            alignement,
        });
    }

    function moveNpc(name, location) {
        dmAction('dm_move_npc', { item_name: name, destination: location });
    }

    function removeNpc(name) {
        if (confirm(`Supprimer le PNJ ${name} ?`)) {
            dmAction('dm_remove_npc', { item_name: name });
            selectPlayer(null);
        }
    }

    function removeNpcs(names) {
        dmAction('dm_remove_npcs', { names });
    }

    function npcAddItem(name, item) {
        dmAction('dm_npc_add_item', { item_name: name, item });
    }

    function npcRemoveItem(name, index) {
        dmAction('dm_npc_remove_item', { item_name: name, item_index: index });
    }

    $: selected = $dmState.selectedPlayer;
    $: selectedData = selected ? ($dmState.players[selected] || $dmState.npcs[selected]) : null;
    $: selectedIsNpc = selected ? !!($dmState.npcs[selected]) : false;
    $: playerList = Object.entries($dmState.players).filter(([, v]) => !v.role);
</script>

<svelte:head>
    <title>D&D - Maître du Donjon</title>
</svelte:head>

<main class="app-root">
    {#if !$dmState.me}
        <div class="login-view slide-up">
            <div class="login-branding">
                <h1 class="brand-title">
                    <span class="brand-icon">👑</span>
                    Maître du Donjon
                </h1>
                <p class="brand-subtitle">Panel d'Administration</p>
            </div>

            <div class="login-form dnd-card">
                <label class="form-label">Identifiant</label>
                <input
                    class="form-input"
                    bind:value={pseudo}
                    placeholder="Pseudo du MDJ"
                    on:keydown={(e) => e.key === 'Enter' && login()}
                />
                <button class="btn-primary" on:click={login}>
                    Entrer dans la salle du trône
                </button>
            </div>
        </div>
    {:else}
        <div class="dm-layout fade-in">
            <!-- Header -->
            <header class="dm-header">
                <div class="header-left">
                    <span class="header-icon">👑</span>
                    <div>
                        <h1 class="header-title">Maître du Donjon</h1>
                        <span class="header-sub">{playerList.length} aventurier{playerList.length !== 1 ? 's' : ''} connecté{playerList.length !== 1 ? 's' : ''}</span>
                    </div>
                </div>
                <div class="header-right">
                    <span class="dm-badge">MDJ</span>
                    <span class="dm-pseudo">{$dmState.me}</span>
                </div>
            </header>

            <div class="dm-content">
                <!-- Left: Player Roster -->
                <aside class="roster-panel dnd-section">
                    <h2 class="panel-title">
                        <span>👥</span> Aventuriers
                    </h2>
                    <div class="roster-list custom-scrollbar">
                        {#each playerList as [pseudo, data]}
                            <button
                                class="roster-item"
                                class:active={selected === pseudo}
                                on:click={() => selectPlayer(pseudo)}
                            >
                                <div class="roster-item-top">
                                    <span class="roster-name">{data.nom}</span>
                                    <span class="roster-class">{data.classe}</span>
                                </div>
                                <div class="roster-item-bottom">
                                    <span class="roster-hp">❤️ {data.pv}/{data.max_pv}</span>
                                    <span class="roster-loc">📍 {data.lieu}</span>
                                </div>
                            </button>
                        {:else}
                            <div class="empty-state">
                                <span class="empty-icon">🏜️</span>
                                <span class="empty-text">Aucun aventurier pour l'instant...</span>
                            </div>
                        {/each}
                    </div>
                </aside>

                <!-- Center: Editor -->
                <main class="editor-panel">
                    {#if selected && selectedData}
                        <div class="fade-in">
                            <div class="editor-header">
                                <h2 class="editor-title">
                                    {selectedData.nom}
                                    {#if selectedIsNpc}<span class="npc-badge">PNJ</span>{/if}
                                    {#if selectedData.mobType === 'boss'}<span class="npc-badge boss">🐲 BOSS</span>{/if}
                                    {#if selectedData.mobType === 'minion'}<span class="npc-badge minion">👹 Minion</span>{/if}
                                </h2>
                                <button class="btn-danger-sm" on:click={() => selectedIsNpc ? removeNpc(selected) : deletePlayer(selected)}>
                                    💀 Supprimer
                                </button>
                            </div>

                            <!-- Alignment -->
                            <div class="dnd-section editor-section">
                                <h3 class="section-title"><span>⚖️</span> Alignement</h3>
                                <div class="align-row">
                                    <input
                                        class="form-input flex-1"
                                        value={selectedData.alignement || 'Neutre'}
                                        on:change={(e) => selectedIsNpc
                                            ? dmAction('dm_edit_npc', { item_name: selected, stats: {}, pv: 0, alignement: e.target.value })
                                            : dmAction('dm_edit_align', { target_player: selected, alignement: e.target.value })}
                                    />
                                </div>
                            </div>

                            <!-- Location -->
                            <div class="dnd-section editor-section">
                                <h3 class="section-title"><span>📍</span> Position</h3>
                                <div class="location-row">
                                    {#each $dmState.locations as loc}
                                        <button
                                            class="loc-btn"
                                            class:current={selectedData.lieu === loc.nom}
                                            on:click={() => selectedIsNpc
                                                ? dmAction('dm_move_npc', { item_name: selected, destination: loc.nom })
                                                : dmAction('dm_move_player', { target_player: selected, destination: loc.nom })}
                                        >
                                            {loc.nom}
                                        </button>
                                    {/each}
                                </div>
                            </div>

                            <StatsEditor
                                stats={selectedData.stats}
                                pv={selectedData.pv}
                                maxPv={selectedData.max_pv}
                                isNpc={selectedIsNpc}
                                equipement={selectedData.equipement || { arme: null, armure: null }}
                                onSave={(stats) => selectedIsNpc
                                    ? dmAction('dm_edit_npc', { item_name: selected, stats, pv: 0, alignement: '', overwrite: true })
                                    : dmAction('dm_edit_stats', { target_player: selected, stats })}
                                onSetPv={(pv) => selectedIsNpc
                                    ? dmAction('dm_edit_npc', { item_name: selected, stats: {}, pv, alignement: '', overwrite: true })
                                    : dmAction('dm_set_pv', { target_player: selected, pv })}
                            />

                            <InventoryEditor
                                inventory={selectedData.inventaire || []}
                                onAdd={(item) => selectedIsNpc
                                    ? dmAction('dm_npc_add_item', { item_name: selected, item })
                                    : dmAction('dm_add_item', { target_player: selected, item })}
                                onRemove={(index) => selectedIsNpc
                                    ? dmAction('dm_npc_remove_item', { item_name: selected, item_index: index })
                                    : dmAction('dm_remove_item', { target_player: selected, item_index: index })}
                            />

                            <SpellEditor
                                spells={selectedData.sorts || []}
                                onAdd={(spell) => selectedIsNpc
                                    ? dmAction('dm_npc_add_spell', { item_name: selected, spell })
                                    : dmAction('dm_add_spell', { target_player: selected, spell })}
                                onRemove={(index) => selectedIsNpc
                                    ? dmAction('dm_npc_remove_spell', { item_name: selected, item_index: index })
                                    : dmAction('dm_remove_spell', { target_player: selected, item_index: index })}
                            />

                            <QuestEditor
                                quests={selectedData.quests || []}
                                onAdd={(quest) => dmAction('dm_add_quest_player', { target_player: selected, quest })}
                                onRemove={(index) => dmAction('dm_remove_quest_player', { target_player: selected, quest_index: index })}
                            />
                        </div>
                    {:else}
                        <div class="empty-editor">
                            <span class="empty-editor-icon">📜</span>
                            <p class="empty-editor-text">Sélectionnez un aventurier pour modifier ses données</p>
                        </div>
                    {/if}
                </main>

                <!-- Right: World + Chat -->
                <aside class="world-panel">
                    <LocationManager
                        locations={$dmState.locations}
                        players={$dmState.players}
                        onAddItem={(loc, item) => dmAction('dm_teleport_item_add', { destination: loc, item })}
                        onRemoveItem={(loc, idx) => dmAction('dm_teleport_item_remove', { destination: loc, item_index: idx })}
                        onEditLocation={(oldName, newName, newBg) => dmAction('dm_edit_location', { destination: oldName, new_name: newName, location_bg: newBg })}
                        onAddQuest={(loc, quest) => dmAction('dm_add_quest', { destination: loc, quest })}
                        onRemoveQuest={(loc, idx) => dmAction('dm_remove_quest', { destination: loc, quest_index: idx })}
                    />

                    <NPCManager
                        npcs={$dmState.npcs}
                        locations={$dmState.locations}
                        onAddNpc={addNpc}
                        onRemoveNpc={removeNpc}
                        onBulkRemove={removeNpcs}
                        onEditNpc={editNpc}
                        onMoveNpc={moveNpc}
                        onNpcAddItem={npcAddItem}
                        onNpcRemoveItem={npcRemoveItem}
                        selectNpc={(name) => selectPlayer(name)}
                    />

                    <ChatBox logs={$dmState.logs} />
                </aside>
            </div>
        </div>
    {/if}
</main>

<style>
    .app-root {
        min-height: 100vh;
        min-height: 100dvh;
        background: radial-gradient(ellipse at top, #0d0a07 0%, #000 60%);
        color: #e8e0d4;
    }

    /* Login */
    .login-view {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 40px 16px 60px;
        min-height: 100vh;
        min-height: 100dvh;
        justify-content: center;
    }

    .login-branding { text-align: center; margin-bottom: 24px; }

    .brand-title {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 2.5rem;
        margin: 0;
        font-weight: 700;
        letter-spacing: 0.05em;
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .brand-icon { font-size: 2rem; }

    .brand-subtitle {
        font-family: 'MedievalSharp', cursive;
        color: #5a5045;
        font-size: 1rem;
        margin: 4px 0 0 0;
        letter-spacing: 0.15em;
    }

    .login-form {
        padding: 32px;
        width: 100%;
        max-width: 380px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .form-label {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.85rem;
        letter-spacing: 0.05em;
    }

    .form-input {
        width: 100%;
        padding: 12px 16px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.25);
        border-radius: 8px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 1rem;
        outline: none;
        transition: border-color 0.2s ease;
    }

    .form-input:focus {
        border-color: #c5a059;
    }

    .btn-primary {
        padding: 14px 24px;
        background: linear-gradient(135deg, #c5a059, #a67c37);
        border: none;
        border-radius: 8px;
        color: #000;
        font-family: 'Cinzel', serif;
        font-weight: 700;
        font-size: 0.95rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-primary:hover {
        transform: translateY(-1px);
        box-shadow: 0 4px 16px rgba(197, 160, 89, 0.3);
    }

    .btn-danger-sm {
        padding: 6px 14px;
        background: rgba(185, 28, 28, 0.2);
        border: 1px solid rgba(239, 68, 68, 0.3);
        border-radius: 6px;
        color: #fca5a5;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-danger-sm:hover {
        background: rgba(185, 28, 28, 0.4);
    }

    /* Dashboard Layout */
    .dm-layout {
        min-height: 100vh;
        min-height: 100dvh;
        padding: 16px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .dm-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px 24px;
        background: linear-gradient(135deg, #0d0d0d, #12100e);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 14px;
    }

    .header-left {
        display: flex;
        align-items: center;
        gap: 14px;
    }

    .header-icon { font-size: 2rem; }

    .header-title {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1.4rem;
        margin: 0;
        font-weight: 700;
    }

    .header-sub {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.8rem;
    }

    .header-right {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .dm-badge {
        padding: 4px 12px;
        background: linear-gradient(135deg, #c5a059, #a67c37);
        border-radius: 100px;
        color: #000;
        font-family: 'Cinzel', serif;
        font-size: 0.7rem;
        font-weight: 700;
        letter-spacing: 0.1em;
    }

    .dm-pseudo {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.9rem;
    }

    .dm-content {
        display: grid;
        grid-template-columns: 280px 1fr 320px;
        gap: 16px;
        flex: 1;
        min-height: 0;
    }

    /* Roster Panel */
    .roster-panel {
        position: sticky;
        top: 16px;
        max-height: calc(100vh - 120px);
        display: flex;
        flex-direction: column;
    }

    .panel-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 1.05rem;
        margin: 0 0 16px 0;
        padding-bottom: 12px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.15);
    }

    .roster-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
        overflow-y: auto;
        flex: 1;
    }

    .roster-item {
        width: 100%;
        padding: 12px 14px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.1);
        border-radius: 10px;
        cursor: pointer;
        text-align: left;
        transition: all 0.2s ease;
        display: flex;
        flex-direction: column;
        gap: 6px;
        color: inherit;
    }

    .roster-item:hover {
        border-color: rgba(197, 160, 89, 0.3);
        background: rgba(26, 20, 16, 0.8);
    }

    .roster-item.active {
        border-color: #c5a059;
        background: rgba(197, 160, 89, 0.1);
        box-shadow: 0 0 12px rgba(197, 160, 89, 0.1);
    }

    .roster-item-top {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .roster-name {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 0.9rem;
        font-weight: 700;
    }

    .roster-class {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.7rem;
    }

    .roster-item-bottom {
        display: flex;
        justify-content: space-between;
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        color: #7a6f5f;
    }

    /* Editor Panel */
    .editor-panel {
        display: flex;
        flex-direction: column;
        gap: 16px;
        overflow-y: auto;
        max-height: calc(100vh - 120px);
        padding-right: 4px;
    }

    .editor-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .editor-title {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1.5rem;
        margin: 0;
        font-weight: 700;
    }

    .npc-badge {
        display: inline-block;
        margin-left: 10px;
        padding: 3px 10px;
        background: rgba(147, 197, 253, 0.12);
        border: 1px solid rgba(147, 197, 253, 0.35);
        border-radius: 100px;
        color: #93c5fd;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        vertical-align: middle;
    }

    .npc-badge.boss {
        background: rgba(239, 68, 68, 0.18);
        border-color: rgba(239, 68, 68, 0.5);
        color: #fca5a5;
    }

    .npc-badge.minion {
        background: rgba(217, 119, 6, 0.15);
        border-color: rgba(217, 119, 6, 0.4);
        color: #fbbf24;
    }

    .editor-section {
        margin-bottom: 0;
    }

    .section-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.95rem;
        margin: 0 0 12px 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.12);
    }

    .align-row {
        display: flex;
        gap: 12px;
    }

    .location-row {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .loc-btn {
        padding: 6px 14px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.15);
        border-radius: 8px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .loc-btn:hover {
        border-color: rgba(197, 160, 89, 0.4);
        color: #c5a059;
    }

    .loc-btn.current {
        background: rgba(197, 160, 89, 0.15);
        border-color: #c5a059;
        color: #c5a059;
    }

    .empty-editor {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        min-height: 400px;
        gap: 16px;
    }

    .empty-editor-icon {
        font-size: 4rem;
        opacity: 0.2;
    }

    .empty-editor-text {
        font-family: 'MedievalSharp', cursive;
        color: #5a5045;
        font-size: 1rem;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 28px 16px;
        border: 2px dashed rgba(197, 160, 89, 0.15);
        border-radius: 12px;
        background: rgba(0, 0, 0, 0.2);
    }

    .empty-icon { font-size: 2rem; opacity: 0.3; }

    .empty-text {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.85rem;
    }

    /* World Panel */
    .world-panel {
        display: flex;
        flex-direction: column;
        gap: 16px;
        position: sticky;
        top: 16px;
        max-height: calc(100vh - 120px);
        overflow-y: auto;
    }

    .flex-1 { flex: 1; }

    @media (max-width: 1200px) {
        .dm-content {
            grid-template-columns: 240px 1fr;
        }
        .world-panel {
            display: none;
        }
    }

    @media (max-width: 768px) {
        .dm-content {
            grid-template-columns: 1fr;
        }
        .roster-panel {
            position: static;
            max-height: 300px;
        }
        .editor-panel {
            max-height: none;
        }
    }
</style>
