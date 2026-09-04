<script>
    import Button from '../atoms/Button.svelte';
    export let item = '';
    export let onLoot = (name) => {};
</script>

<div class="loot-item">
    <div class="loot-info">
        <span class="loot-icon">💎</span>
        <div class="loot-details">
            <span class="loot-name">{item.nom}</span>
            {#if item.type}
                <span class="loot-type">{item.type}</span>
            {/if}
        </div>
    </div>
    <Button variant="success" onClick={() => onLoot(item.nom)} className="py-1.5 px-4 text-sm">
        Ramasser
    </Button>
</div>

<style>
    .loot-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 10px;
        transition: all 0.2s ease;
    }

    .loot-item:hover {
        border-color: rgba(197, 160, 89, 0.4);
        background: rgba(26, 20, 16, 0.8);
    }

    .loot-info {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .loot-icon {
        font-size: 1.4rem;
    }

    .loot-details {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .loot-name {
        font-family: 'MedievalSharp', cursive;
        color: #f5f0e8;
        font-size: 0.95rem;
    }

    .loot-type {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.75rem;
        font-style: italic;
    }
</style>
