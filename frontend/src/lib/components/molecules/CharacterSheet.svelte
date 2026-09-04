<script>
    import HPBar from '../atoms/HPBar.svelte';

    export let stats = {};

    $: maxPv = (stats.stats?.constitution || 10) * 10;

    $: statList = [
        { label: 'Force', val: stats.stats?.force, icon: '💪' },
        { label: 'Constitution', val: stats.stats?.constitution, icon: '🫀' },
        { label: 'Vitesse', val: stats.stats?.vitesse, icon: '👟' },
        { label: 'Savoir', val: stats.stats?.savoir, icon: '📚' },
        { label: 'Instinct', val: stats.stats?.instinct, icon: '👁️' },
        { label: 'Charisme', val: stats.stats?.charisme, icon: '✨' }
    ];
</script>

<div class="character-sheet">
    <div class="sheet-header">
        <div class="char-identity">
            <span class="char-label">Nom du Personnage</span>
            <h2 class="char-name">{stats.nom || 'Inconnu'}</h2>
        </div>
        <div class="char-class">
            <span class="char-label">Classe & Origine</span>
            <span class="char-class-name">{stats.classe || 'Sans Classe'}</span>
        </div>
    </div>

    <div class="sheet-body">
        <!-- Stats Grid -->
        <div class="stats-grid">
            {#each statList as stat}
                <div class="stat-card">
                    <span class="stat-icon">{stat.icon}</span>
                    <span class="stat-label">{stat.label}</span>
                    <span class="stat-value">{stat.val || 0}</span>
                </div>
            {/each}
        </div>

        <!-- Sidebar Info -->
        <div class="char-sidebar">
            <div class="sidebar-block">
                <span class="sidebar-label">Lieu Actuel</span>
                <span class="sidebar-value">{stats.lieu}</span>
            </div>
            <div class="sidebar-divider"></div>
            <div class="sidebar-block">
                <div class="hp-header">
                    <span class="sidebar-label">Points de Vie</span>
                    <span class="hp-value">{stats.pv} / {maxPv}</span>
                </div>
                <HPBar current={stats.pv} max={maxPv} />
            </div>
        </div>
    </div>
</div>

<style>
    .character-sheet {
        background: linear-gradient(145deg, #f4e4bc, #e8d4a9);
        border: 8px double #5c4033;
        border-radius: 6px;
        padding: 20px;
        color: #2c1e16;
        box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5),
                    inset 0 1px 0 rgba(255, 255, 255, 0.3);
        position: relative;
        overflow: hidden;
    }

    .character-sheet::before {
        content: '📜';
        position: absolute;
        top: -8px;
        right: -4px;
        font-size: 5rem;
        opacity: 0.08;
        pointer-events: none;
        transform: rotate(15deg);
    }

    .sheet-header {
        display: flex;
        flex-direction: column;
        gap: 12px;
        padding-bottom: 16px;
        border-bottom: 2px solid rgba(92, 64, 51, 0.2);
        margin-bottom: 16px;
    }

    .char-label {
        display: block;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.6rem;
        text-transform: uppercase;
        letter-spacing: 0.1em;
        color: #8b6544;
        margin-bottom: 2px;
    }

    .char-name {
        font-family: 'Cinzel', serif;
        font-size: 1.5rem;
        font-weight: 700;
        margin: 0;
        color: #2c1e16;
        letter-spacing: 0.03em;
    }

    .char-class-name {
        font-family: 'MedievalSharp', cursive;
        font-size: 1rem;
        color: #5c4033;
        font-style: italic;
    }

    .sheet-body {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .stats-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 8px;
    }

    .stat-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 2px;
        padding: 10px 4px;
        background: rgba(232, 212, 169, 0.6);
        border: 1px solid rgba(92, 64, 51, 0.15);
        border-radius: 8px;
        transition: background 0.2s ease;
    }

    .stat-card:hover {
        background: rgba(232, 212, 169, 0.9);
    }

    .stat-icon {
        font-size: 1.1rem;
        line-height: 1;
    }

    .stat-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.6rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        color: #8b6544;
    }

    .stat-value {
        font-family: 'Cinzel', serif;
        font-size: 1.3rem;
        font-weight: 700;
        color: #2c1e16;
    }

    .char-sidebar {
        background: rgba(220, 196, 149, 0.5);
        border: 1px solid rgba(92, 64, 51, 0.12);
        border-radius: 8px;
        padding: 14px;
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .sidebar-block {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .sidebar-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.6rem;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: #8b6544;
    }

    .sidebar-value {
        font-family: 'MedievalSharp', cursive;
        font-size: 1rem;
        font-weight: bold;
        color: #2c1e16;
    }

    .sidebar-divider {
        height: 1px;
        background: rgba(92, 64, 51, 0.15);
    }

    .hp-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .hp-value {
        font-family: 'Cinzel', serif;
        font-size: 0.85rem;
        font-weight: 700;
        color: #2c1e16;
    }

    @media (max-width: 480px) {
        .stats-grid {
            grid-template-columns: repeat(2, 1fr);
        }
    }
</style>
