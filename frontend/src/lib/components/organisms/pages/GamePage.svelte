<script lang="ts">
    import CharacterSheet from '../../molecules/CharacterSheet.svelte';
    import LocationExplorer from '../LocationExplorer.svelte';
    import ChatBox from '../ChatBox.svelte';
    import InventoryItem from '../../molecules/InventoryItem.svelte';
    import Button from '../../atoms/Button.svelte';
    import HPBar from '../../atoms/HPBar.svelte';

    export let gameState;
    export let myStats;
    export let onMove = (dest) => {};
    export let onHit = (target) => {};
    export let onUseItem = (name) => {};
    export let onLootItem = (name) => {};

    let showMobilePanel = 'game';

    $: sameZonePlayers = Object.entries(gameState.players).filter(
        ([p, v]) => v.lieu === gameState.location && p !== gameState.me
    );
    $: zonePlayerCount = sameZonePlayers.length;
</script>

<div class="game-container fade-in">
    <!-- Header Banner -->
    <header class="game-header">
        <div class="header-left">
            <div class="avatar-ring">
                <span class="avatar-icon">👤</span>
            </div>
            <div class="header-info">
                <span class="header-label">Aventurier</span>
                <h2 class="header-name">{gameState.me}</h2>
            </div>
        </div>
        <div class="location-badge">
            <span class="location-icon">📍</span>
            <span class="location-text">{gameState.location}</span>
        </div>
        <div class="player-count-badge">
            <span>{zonePlayerCount}</span>
            <span class="player-count-label">{zonePlayerCount === 1 ? 'here' : 'here'}</span>
        </div>
    </header>

    <!-- Mobile Tab Switcher -->
    <div class="mobile-tabs">
        <button
            class="mobile-tab"
            class:active={showMobilePanel === 'game'}
            on:click={() => showMobilePanel = 'game'}
        >
            ⚔️ Jeu
        </button>
        <button
            class="mobile-tab"
            class:active={showMobilePanel === 'chat'}
            on:click={() => showMobilePanel = 'chat'}
        >
            📜 Journal
        </button>
        <button
            class="mobile-tab"
            class:active={showMobilePanel === 'players'}
            on:click={() => showMobilePanel = 'players'}
        >
            👥 Héros
        </button>
    </div>

    <!-- Main Content Layout -->
    <div class="game-layout" class:show-chat={showMobilePanel === 'chat'} class:show-players={showMobilePanel === 'players'}>
        <!-- Sidebar -->
        <aside class="sidebar">
            {#if myStats}
                <div class="fade-in">
                    <CharacterSheet stats={myStats} />
                </div>
            {/if}

            <!-- Inventory -->
            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">🎒</span> Sac de Voyage
                </h3>
                <div class="inventory-list custom-scrollbar">
                    {#if myStats && myStats.inventaire && myStats.inventaire.length > 0}
                        {#each myStats.inventaire as item, i}
                            <InventoryItem {item} onUse={onUseItem} />
                        {/each}
                    {:else}
                        <div class="empty-state">
                            <span class="empty-icon">🎒</span>
                            <span class="empty-text">Votre sac est vide...</span>
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Move / Travel -->
            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">🗺️</span> Ordre de Marche
                </h3>
                <div class="move-buttons">
                    {#each gameState.locations as loc}
                        {#if loc.nom !== gameState.location}
                            <Button
                                onClick={() => onMove(loc.nom)}
                                variant={loc.nom === 'Taverne' ? 'primary' : loc.nom === 'Donjon' ? 'danger' : 'arcane'}
                                className="w-full text-sm"
                            >
                                {loc.nom === 'Taverne' ? '🏠' : loc.nom === 'Donjon' ? '⚔️' : loc.nom === 'Foret Enchantee' ? '🌿' : loc.nom === 'Montagne Rocheuse' ? '⛰️' : loc.nom === 'Marais Hante' ? '👻' : loc.nom === 'Plaine des Conflits' ? '🚩' : loc.nom === 'Temple Abandonne' ? '🏛️' : '📍'}
                                {loc.nom}
                            </Button>
                        {/if}
                    {/each}
                </div>
            </div>
        </aside>

        <!-- Main Workspace -->
        <main class="workspace">
            <!-- Top Row: Explorer + Players -->
            <div class="workspace-top">
                <div class="fade-in">
                    <LocationExplorer
                        locationObjects={gameState.currentLocationObjects}
                        onLoot={onLootItem}
                    />
                </div>

                <!-- Players Roster -->
                <div class="dnd-section players-section">
                    <h3 class="section-title">
                        <span class="section-icon">👥</span> Heros Present ({zonePlayerCount})
                    </h3>
                    <div class="players-list custom-scrollbar">
                        {#each sameZonePlayers as [p, v]}
                            <div class="player-card">
                                <div class="player-header">
                                    <strong class="player-name">{v.nom}</strong>
                                    <span class="player-location">{v.lieu}</span>
                                </div>
                                <HPBar current={v.pv} max={v.max_pv} />
                                <Button variant="danger" onClick={() => onHit(p)} className="w-full mt-2 py-2 text-sm">
                                    ⚔️ Attaquer
                                </Button>
                            </div>
                        {:else}
                            <div class="empty-state">
                                <span class="empty-icon">👥</span>
                                <span class="empty-text">Personne d'autre ici...</span>
                            </div>
                        {/each}
                    </div>
                </div>
            </div>

            <!-- Chat -->
            <div class="fade-in">
                <ChatBox logs={gameState.logs} />
            </div>
        </main>
    </div>
</div>

<style>
    .game-container {
        max-width: 1280px;
        margin: 0 auto;
        padding: 24px 16px;
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    /* Header */
    .game-header {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        padding: 20px 28px;
        background: linear-gradient(135deg, #0d0d0d, #12100e);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 16px;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
    }

    .header-left {
        display: flex;
        align-items: center;
        gap: 16px;
    }

    .avatar-ring {
        width: 52px;
        height: 52px;
        background: linear-gradient(135deg, #c5a059, #a67c37);
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.5rem;
        box-shadow: 0 0 20px rgba(197, 160, 89, 0.3);
        flex-shrink: 0;
    }

    .avatar-icon {
        filter: brightness(0) invert(1);
    }

    .header-info {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .header-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        color: #7a6f5f;
        text-transform: uppercase;
        letter-spacing: 0.12em;
    }

    .header-name {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1.4rem;
        margin: 0;
        font-weight: 700;
    }

    .location-badge {
        display: flex;
        align-items: center;
        gap: 8px;
        background: rgba(0, 0, 0, 0.5);
        padding: 10px 20px;
        border-radius: 100px;
        border: 1px solid rgba(197, 160, 89, 0.2);
    }

    .location-icon {
        font-size: 1.1rem;
    }

    .location-text {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.95rem;
    }

    .player-count-badge {
        display: flex;
        align-items: center;
        gap: 4px;
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.85rem;
    }

    .player-count-badge span:first-child {
        font-size: 1.2rem;
        font-weight: bold;
        color: #c5a059;
    }

    .player-count-label {
        font-size: 0.75rem;
    }

    /* Mobile Tabs */
    .mobile-tabs {
        display: none;
        gap: 8px;
        padding: 4px;
        background: #0a0a0a;
        border-radius: 12px;
        border: 1px solid rgba(197, 160, 89, 0.15);
    }

    .mobile-tab {
        flex: 1;
        padding: 10px 8px;
        background: transparent;
        border: none;
        border-radius: 8px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.85rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .mobile-tab.active {
        background: rgba(197, 160, 89, 0.12);
        color: #c5a059;
    }

    /* Layout */
    .game-layout {
        display: grid;
        grid-template-columns: 320px 1fr;
        gap: 24px;
        align-items: start;
    }

    .sidebar {
        display: flex;
        flex-direction: column;
        gap: 20px;
        position: sticky;
        top: 16px;
    }

    .workspace {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .workspace-top {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
    }

    /* Sections */
    .section-title {
        display: flex;
        align-items: center;
        gap: 10px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 1.1rem;
        margin: 0 0 16px 0;
        padding-bottom: 12px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.15);
    }

    .section-icon {
        font-size: 1.3rem;
        opacity: 0.8;
    }

    .inventory-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
        max-height: 280px;
        overflow-y: auto;
        padding-right: 4px;
    }

    .move-buttons {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    /* Empty State */
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

    .empty-icon {
        font-size: 2rem;
        opacity: 0.3;
    }

    .empty-text {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.9rem;
    }

    /* Players */
    .players-section {
        height: 100%;
        display: flex;
        flex-direction: column;
    }

    .players-list {
        display: flex;
        flex-direction: column;
        gap: 12px;
        flex: 1;
        overflow-y: auto;
        max-height: 440px;
        padding-right: 4px;
    }

    .player-card {
        padding: 16px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.1);
        border-radius: 12px;
        transition: all 0.2s ease;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .player-card:hover {
        border-color: rgba(197, 160, 89, 0.3);
        background: rgba(26, 20, 16, 0.8);
    }

    .player-card.self {
        border-color: rgba(34, 197, 94, 0.3);
        box-shadow: 0 0 12px rgba(34, 197, 94, 0.08);
    }

    .player-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 8px;
    }

    .player-name {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1rem;
    }

    .player-location {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        color: #7a6f5f;
        padding: 3px 10px;
        background: rgba(197, 160, 89, 0.08);
        border-radius: 100px;
        border: 1px solid rgba(197, 160, 89, 0.15);
        white-space: nowrap;
    }

    /* Mobile */
    @media (max-width: 1024px) {
        .game-layout {
            grid-template-columns: 1fr;
        }

        .sidebar {
            position: static;
        }

        .workspace-top {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 640px) {
        .game-container {
            padding: 12px 8px;
        }

        .game-header {
            padding: 16px;
        }

        .header-name {
            font-size: 1.1rem;
        }

        .mobile-tabs {
            display: flex;
        }

        .game-layout :global(> .sidebar) {
            display: none;
        }

        .game-layout.show-chat .workspace > :first-child {
            display: none;
        }
    }
</style>
