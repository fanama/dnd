<script>
    import CharacterSheet from '../../molecules/CharacterSheet.svelte';
    import LocationExplorer from '../LocationExplorer.svelte';
    import ChatBox from '../ChatBox.svelte';
    import Button from '../../atoms/Button.svelte';
    import HPBar from '../../atoms/HPBar.svelte';
    import { myStats, myDerivedStats } from '../../../stores/game';

    export let gameState;
    export let onMove = (dest) => {};
    export let onHit = (target) => {};
    export let onUseItem = (name) => {};
    export let onLootItem = (name) => {};

    let activeTab = 'charsheet';
    let equips = { weapon: null, armor: null };

    $: myStatsData = $myStats;
    $: myDerived = $myDerivedStats;

    $: sameZonePlayers = Object.entries(gameState.players).filter(
        ([p, v]) => v.lieu === gameState.location && p !== gameState.me
    );
    $: zonePlayerCount = sameZonePlayers.length;

    function equipItem(item) {
        if (item.isConsumable) {
            onUseItem(item.nom);
            return;
        }
        const slot = item.bonusDegats ? 'weapon' : item.bonusArmure ? 'armor' : null;
        if (slot) {
            equips[slot] = item;
        }
    }

    function unequipSlot(slot) {
        equips[slot] = null;
    }

    function isEquipped(item) {
        return equips.weapon === item || equips.armor === item;
    }
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
        <div class="header-actions">
            <div class="location-badge">
                <span class="location-icon">📍</span>
                <span class="location-text">{gameState.location}</span>
            </div>
            <div class="player-count-badge" title="Aventuriers présents">
                <span>👥</span>
                <span class="player-count">{zonePlayerCount}</span>
            </div>
        </div>
    </header>

    <!-- Tab Switcher (Visible on all sizes; on desktop nav is via layout) -->
    <nav class="view-tabs" aria-label="Navigation">
        <button
            class="view-tab"
            class:active={activeTab === 'charsheet'}
            on:click={() => activeTab = 'charsheet'}
        >
            <span class="tab-icon">📜</span>
            <span class="tab-label">Fiche</span>
        </button>
        <button
            class="view-tab"
            class:active={activeTab === 'equipment'}
            on:click={() => activeTab = 'equipment'}
        >
            <span class="tab-icon">🎒</span>
            <span class="tab-label">Équipement</span>
        </button>
        <button
            class="view-tab"
            class:active={activeTab === 'world'}
            on:click={() => activeTab = 'world'}
        >
            <span class="tab-icon">🌍</span>
            <span class="tab-label">Monde</span>
        </button>
    </nav>

    <!-- ============ CHARACTER SHEET VIEW ============ -->
    {#if activeTab === 'charsheet'}
        <section class="view-panel fade-in">
            <div class="dnd-section">
                {#if myStatsData}
                    <CharacterSheet stats={myStatsData} derivedStats={myDerived} />
                {/if}
            </div>

            <!-- Character Sheet detail: spells & alignment -->
            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">📖</span> Détails du Personnage
                </h3>
                <div class="detail-grid">
                    <div class="detail-item">
                        <span class="detail-label">Alignement</span>
                        <span class="detail-value">{myStatsData?.alignement || '—'}</span>
                    </div>
                    <div class="detail-item">
                        <span class="detail-label">Localisation</span>
                        <span class="detail-value">{gameState.location}</span>
                    </div>
                </div>
            </div>

            {#if myStatsData?.sorts && myStatsData.sorts.length > 0}
                <div class="dnd-section">
                    <h3 class="section-title">
                        <span class="section-icon">🔮</span> Sorts ({myStatsData.sorts.length})
                    </h3>
                    <div class="spells-list">
                        {#each (myStatsData.sorts || []) as sort}
                            <div class="spell-card">
                                <div class="spell-header">
                                    <span class="spell-name">{sort.nom}</span>
                                    <span class="spell-level">Niv. {sort.niveauSort}</span>
                                </div>
                                <div class="spell-meta">
                                    <span class="spell-meta-item">🏛️ {sort.ecoleMagie || 'Arcanes'}</span>
                                    {#if sort.portee}<span class="spell-meta-item">📏 {sort.portee}</span>{/if}
                                    {#if sort.duree}<span class="spell-meta-item">⏳ {sort.duree}</span>{/if}
                                </div>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}
        </section>
    {/if}

    <!-- ============ EQUIPMENT VIEW ============ -->
    {#if activeTab === 'equipment'}
        <section class="view-panel fade-in">
            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">🎒</span> Gérer l'Équipement
                </h3>
                <div class="equip-slots">
                    {#each ['weapon', 'armor'] as slot}
                        <div class="equip-slot" class:empty={!equips[slot]}>
                            <div class="slot-header">
                                <span class="slot-name">{slot === 'weapon' ? '⚔️ Arme' : '🛡️ Armure'}</span>
                                {#if equips[slot]}
                                    <button class="slot-clear" on:click={() => unequipSlot(slot)} aria-label="Retirer">
                                        ✖ Retirer
                                    </button>
                                {/if}
                            </div>
                            {#if equips[slot]}
                                <div class="slot-item">
                                    <span class="slot-item-icon">{slot === 'weapon' ? '⚔️' : '🛡️'}</span>
                                    <div class="slot-item-info">
                                        <span class="slot-item-name">{equips[slot].nom}</span>
                                        {#if equips[slot].bonusDegats}
                                            <span class="slot-item-stat">+{equips[slot].bonusDegats} ATK</span>
                                        {/if}
                                        {#if equips[slot].bonusArmure}
                                            <span class="slot-item-stat">+{equips[slot].bonusArmure} DEF</span>
                                        {/if}
                                    </div>
                                </div>
                            {:else}
                                <div class="slot-empty">Aucune {slot === 'weapon' ? 'arme' : 'armure'} équipée</div>
                            {/if}
                        </div>
                    {/each}
                </div>
            </div>

            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">🎒</span> Sac de Voyage
                </h3>
                {#if myStatsData?.inventaire && myStatsData.inventaire.length > 0}
                    <div class="inventory-list custom-scrollbar">
                        {#each myStatsData.inventaire as item, i}
                            <div
                                class="inv-item"
                                class:equipped={isEquipped(item)}
                                style="--bd-color: {item.isConsumable ? '#2e7d32' : item.bonusDegats ? '#b71c1c' : item.bonusArmure ? '#1565c0' : '#c5a059'}"
                            >
                                <div class="inv-info">
                                    <span class="inv-icon">{item.isConsumable ? '🧪' : item.bonusDegats ? '⚔️' : item.bonusArmure ? '🛡️' : '📦'}</span>
                                    <div class="inv-details">
                                        <span class="inv-name">
                                            {item.nom}
                                            {#if isEquipped(item)}<span class="equipped-badge">Équipé</span>{/if}
                                        </span>
                                        <div class="inv-stats">
                                            {#if item.bonusDegats}<span class="stat atk">+{item.bonusDegats} ATK</span>{/if}
                                            {#if item.bonusArmure}<span class="stat def">+{item.bonusArmure} DEF</span>{/if}
                                            {#if item.prix}<span class="stat gold">💰 {item.prix}p</span>{/if}
                                        </div>
                                    </div>
                                </div>
                                <div class="inv-actions">
                                    {#if item.isConsumable}
                                        <Button variant="success" onClick={() => onUseItem(item.nom)} className="px-2 py-1 text-xs">
                                            Boire
                                        </Button>
                                    {:else if !isEquipped(item)}
                                        <Button variant="primary" onClick={() => equipItem(item)} className="px-2 py-1 text-xs">
                                            Équiper
                                        </Button>
                                    {:else}
                                        <Button variant="danger" onClick={() => unequipSlot(item.bonusDegats ? 'weapon' : 'armor')} className="px-2 py-1 text-xs">
                                            Retirer
                                        </Button>
                                    {/if}
                                </div>
                            </div>
                        {/each}
                    </div>
                {:else}
                    <div class="empty-state">
                        <span class="empty-icon">🎒</span>
                        <span class="empty-text">Votre sac est vide...</span>
                        <span class="empty-hint">Explorez le monde pour trouver des trésors</span>
                    </div>
                {/if}
            </div>
        </section>
    {/if}

    <!-- ============ WORLD VIEW ============ -->
    {#if activeTab === 'world'}
        <section class="view-panel fade-in">
            <!-- Player health & chars present -->
            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">❤️</span> Points de Vie
                </h3>
                {#if myStatsData}
                    <div class="own-hp">
                        <div class="own-hp-row">
                            <span class="own-name">{myStatsData.nom}</span>
                            <span class="own-class">{myStatsData.classe}</span>
                        </div>
                        <HPBar current={myStatsData.pv} max={myStatsData.max_pv || myDerived.maxPv} />
                    </div>
                {/if}
            </div>

            <!-- Exploration -->
            <div class="fade-in">
                <LocationExplorer
                    locationObjects={gameState.currentLocationObjects}
                    onLoot={onLootItem}
                />
            </div>

            <!-- Players Roster -->
            <div class="dnd-section">
                <h3 class="section-title">
                    <span class="section-icon">👥</span> Héros Présents ({zonePlayerCount})
                </h3>
                <div class="players-list custom-scrollbar">
                    {#each sameZonePlayers as [p, v]}
                        <div class="player-card">
                            <div class="player-header">
                                <strong class="player-name">{v.nom}</strong>
                                <span class="player-location">{v.classe}</span>
                            </div>
                            <HPBar current={v.pv} max={v.max_pv} />
                            <Button variant="danger" onClick={() => onHit(p)} className="w-full py-2 text-sm">
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

                {#if gameState.npcs && gameState.npcs.length > 0}
                    <h3 class="section-title npc-title">
                        <span class="section-icon">🤝</span> PNJ Présents ({gameState.npcs.length})
                    </h3>
                    <div class="players-list custom-scrollbar">
                        {#each gameState.npcs as npc}
                            <div class="player-card npc-card">
                                <div class="player-header">
                                    <strong class="player-name">{npc.nom}</strong>
                                    <span class="player-location">{npc.classe}</span>
                                </div>
                                <HPBar current={npc.pv} max={npc.max_pv} />
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>

            <!-- Movement -->
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

            <!-- Chat -->
            <div class="fade-in">
                <ChatBox logs={gameState.logs} />
            </div>
        </section>
    {/if}
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

    .header-actions {
        display: flex;
        align-items: center;
        gap: 12px;
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
        gap: 6px;
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.9rem;
        padding: 8px 14px;
        background: rgba(0, 0, 0, 0.5);
        border-radius: 100px;
        border: 1px solid rgba(197, 160, 89, 0.15);
    }

    .player-count {
        font-size: 1.1rem;
        font-weight: bold;
        color: #c5a059;
    }

    /* View Tab Switcher */
    .view-tabs {
        display: flex;
        gap: 8px;
        padding: 6px;
        background: #0a0a0a;
        border-radius: 12px;
        border: 1px solid rgba(197, 160, 89, 0.15);
        position: sticky;
        top: 8px;
        z-index: 20;
        backdrop-filter: blur(8px);
    }

    .view-tab {
        flex: 1;
        padding: 12px 8px;
        background: transparent;
        border: none;
        border-radius: 8px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.9rem;
        cursor: pointer;
        transition: all 0.2s ease;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 3px;
    }

    .view-tab .tab-icon {
        font-size: 1.2rem;
    }

    .view-tab.active {
        background: linear-gradient(135deg, rgba(197, 160, 89, 0.18), rgba(197, 160, 89, 0.08));
        color: #c5a059;
        border: 1px solid rgba(197, 160, 89, 0.25);
    }

    .view-tab:hover:not(.active) {
        color: #a09080;
    }

    /* View Panel */
    .view-panel {
        display: flex;
        flex-direction: column;
        gap: 20px;
        max-width: 720px;
        margin: 0 auto;
        width: 100%;
    }

    /* Sections */
    .dnd-section {
        background-color: #0a0a0a;
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 12px;
        padding: 24px;
        transition: border-color 0.3s ease;
    }

    .dnd-section:hover {
        border-color: rgba(197, 160, 89, 0.4);
    }

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

    .npc-title {
        margin-top: 20px;
        color: #93c5fd;
    }

    /* Detail Grid */
    .detail-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
    }

    .detail-item {
        display: flex;
        flex-direction: column;
        gap: 4px;
        padding: 12px;
        background: rgba(26, 20, 16, 0.6);
        border-radius: 8px;
        border: 1px solid rgba(197, 160, 89, 0.1);
    }

    .detail-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.65rem;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: #7a6f5f;
    }

    .detail-value {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1rem;
        font-weight: 700;
        word-break: break-word;
    }

    /* Spells */
    .spells-list {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .spell-card {
        padding: 12px 16px;
        background: rgba(69, 39, 160, 0.15);
        border: 1px solid rgba(139, 92, 246, 0.25);
        border-radius: 10px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .spell-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 8px;
    }

    .spell-name {
        font-family: 'MedievalSharp', cursive;
        color: #c4b5fd;
        font-size: 1rem;
        font-weight: bold;
    }

    .spell-level {
        font-family: 'Alegreya', serif;
        font-size: 0.7rem;
        background: rgba(139, 92, 246, 0.2);
        color: #c4b5fd;
        padding: 2px 8px;
        border-radius: 100px;
        white-space: nowrap;
    }

    .spell-meta {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
    }

    .spell-meta-item {
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        color: #9d8ecb;
    }

    /* Equip slots */
    .equip-slots {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
    }

    .equip-slot {
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 12px;
        padding: 12px;
        display: flex;
        flex-direction: column;
        gap: 8px;
        min-height: 90px;
        transition: all 0.2s ease;
    }

    .equip-slot:not(.empty) {
        border-color: rgba(197, 160, 89, 0.45);
        background: rgba(26, 20, 16, 0.85);
        box-shadow: 0 0 14px rgba(197, 160, 89, 0.12);
    }

    .slot-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .slot-name {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.75rem;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: #7a6f5f;
    }

    .slot-clear {
        background: transparent;
        border: none;
        color: #7a6f5f;
        cursor: pointer;
        font-size: 0.75rem;
        font-family: 'Alegreya', serif;
        opacity: 0.8;
        padding: 2px 6px;
        border-radius: 4px;
    }

    .slot-clear:hover {
        color: #ef4444;
        background: rgba(239, 68, 68, 0.1);
    }

    .slot-item {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .slot-item-icon {
        font-size: 1.6rem;
    }

    .slot-item-info {
        display: flex;
        flex-direction: column;
        gap: 3px;
        min-width: 0;
    }

    .slot-item-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.9rem;
        font-weight: bold;
        word-break: break-word;
    }

    .slot-item-stat {
        font-family: 'Alegreya', serif;
        font-size: 0.72rem;
        color: #e8d4a9;
    }

    .slot-empty {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-size: 0.85rem;
        font-style: italic;
        display: flex;
        align-items: center;
        justify-content: center;
        flex: 1;
    }

    /* Inventory */
    .inventory-list {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .inv-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 12px;
        padding: 10px 14px;
        background: rgba(26, 20, 16, 0.6);
        border-left: 3px solid var(--bd-color);
        border-radius: 8px;
        transition: all 0.2s ease;
    }

    .inv-item:hover {
        background: rgba(26, 20, 16, 0.9);
        transform: translateX(2px);
    }

    .inv-item.equipped {
        background: rgba(26, 20, 16, 0.9);
        box-shadow: 0 0 10px rgba(197, 160, 89, 0.1);
    }

    .inv-info {
        display: flex;
        align-items: center;
        gap: 10px;
        min-width: 0;
        flex: 1;
    }

    .inv-icon {
        font-size: 1.3rem;
        flex-shrink: 0;
    }

    .inv-details {
        display: flex;
        flex-direction: column;
        gap: 4px;
        min-width: 0;
    }

    .inv-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.9rem;
        font-weight: bold;
        display: flex;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .equipped-badge {
        font-family: 'Alegreya', serif;
        font-size: 0.6rem;
        color: #4ade80;
        background: rgba(34, 197, 94, 0.15);
        border: 1px solid rgba(34, 197, 94, 0.3);
        padding: 1px 6px;
        border-radius: 100px;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        white-space: nowrap;
    }

    .inv-stats {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
    }

    .stat {
        font-family: 'Alegreya', serif;
        font-size: 0.7rem;
    }

    .stat.atk { color: #ef5350; font-weight: bold; }
    .stat.def { color: #42a5f5; font-weight: bold; }
    .stat.gold { color: #c5a059; }

    .inv-actions {
        display: flex;
        gap: 6px;
        flex-shrink: 0;
    }

    /* Own HP */
    .own-hp {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .own-hp-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .own-name {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1.1rem;
        font-weight: 700;
    }

    .own-class {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.8rem;
    }

    /* Players */
    .players-list {
        display: flex;
        flex-direction: column;
        gap: 12px;
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

    .npc-card {
        border-color: rgba(147, 197, 253, 0.25);
        background: rgba(30, 41, 59, 0.4);
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

    .empty-hint {
        font-family: 'Alegreya', serif;
        color: #4a4035;
        font-size: 0.8rem;
    }

    .move-buttons {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    /* Responsive */
    @media (max-width: 480px) {
        .game-container {
            padding: 12px 8px;
        }

        .game-header {
            padding: 16px;
            flex-direction: column;
            align-items: flex-start;
        }

        .header-actions {
            width: 100%;
            justify-content: space-between;
        }

        .header-name {
            font-size: 1.1rem;
        }

        .detail-grid {
            grid-template-columns: 1fr;
        }

        .equip-slots {
            grid-template-columns: 1fr;
        }

        .inv-actions {
            flex-direction: column;
        }

        .dnd-section {
            padding: 16px;
        }
    }
</style>
