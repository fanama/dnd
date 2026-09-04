<script>
    import LootItem from '../molecules/LootItem.svelte';
    export let locationObjects = [];
    export let onLoot = (name) => {};
</script>

<div class="explorer-card dnd-section">
    <h3 class="section-title">
        <span class="section-icon">🔍</span> Exploration du Lieu
    </h3>

    {#if locationObjects && locationObjects.length > 0}
        <div class="loot-grid">
            {#each locationObjects as item}
                <LootItem {item} onLoot={onLoot} />
            {/each}
        </div>
    {:else}
        <div class="empty-state">
            <span class="empty-icon">🔍</span>
            <span class="empty-text">Rien a ramasser ici...</span>
            <span class="empty-hint">Deplacez-vous pour decouvrir des tresors</span>
        </div>
    {/if}
</div>

<style>
    .explorer-card {
        height: 100%;
        display: flex;
        flex-direction: column;
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

    .loot-grid {
        display: flex;
        flex-direction: column;
        gap: 10px;
        flex: 1;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 32px 16px;
        border: 2px dashed rgba(197, 160, 89, 0.12);
        border-radius: 12px;
        background: rgba(0, 0, 0, 0.2);
    }

    .empty-icon {
        font-size: 2.5rem;
        opacity: 0.25;
    }

    .empty-text {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.95rem;
    }

    .empty-hint {
        font-family: 'Alegreya', serif;
        color: #4a4035;
        font-size: 0.8rem;
    }
</style>
