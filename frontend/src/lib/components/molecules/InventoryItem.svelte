<script>
    import Button from '../atoms/Button.svelte';
    export let item = {};
    export let onUse = (name) => {};

    $: name = typeof item === 'string' ? item : item.nom;
    $: isConsumable = typeof item === 'object' ? item.isConsumable : item.toLowerCase().includes('potion');
    $: isWeapon = typeof item === 'object' && item.bonusDegats;
    $: isArmor = typeof item === 'object' && item.bonusArmure;

    $: itemIcon = isWeapon ? '⚔️' : isArmor ? '🛡️' : isConsumable ? '🧪' : '📦';
    $: itemBorder = isWeapon ? '#b71c1c' : isArmor ? '#1565c0' : isConsumable ? '#2e7d32' : '#c5a059';
</script>

<div class="inventory-item" style="border-left-color: {itemBorder}">
    <div class="item-info">
        <span class="item-icon">{itemIcon}</span>
        <div class="item-details">
            <span class="item-name">{name}</span>
            <div class="item-stats">
                {#if typeof item === 'object'}
                    {#if item.prix}
                        <span class="stat stat-gold">💰 {item.prix}p</span>
                    {/if}
                    {#if item.bonusDegats}
                        <span class="stat stat-atk">⚔️ +{item.bonusDegats} ATK</span>
                    {/if}
                    {#if item.bonusArmure}
                        <span class="stat stat-def">🛡️ +{item.bonusArmure} DEF</span>
                    {/if}
                {/if}
            </div>
        </div>
    </div>
    {#if isConsumable}
        <Button variant="success" onClick={() => onUse(name)} className="py-1 px-3 text-xs">
            Boire
        </Button>
    {/if}
</div>

<style>
    .inventory-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 14px;
        background: rgba(26, 20, 16, 0.6);
        border-left: 3px solid;
        border-radius: 8px;
        transition: all 0.2s ease;
    }

    .inventory-item:hover {
        background: rgba(26, 20, 16, 0.9);
        transform: translateX(2px);
    }

    .item-info {
        display: flex;
        align-items: center;
        gap: 10px;
        min-width: 0;
    }

    .item-icon {
        font-size: 1.3rem;
        flex-shrink: 0;
    }

    .item-details {
        display: flex;
        flex-direction: column;
        gap: 3px;
        min-width: 0;
    }

    .item-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.9rem;
        font-weight: bold;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .item-stats {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
    }

    .stat {
        font-family: 'Alegreya', serif;
        font-size: 0.7rem;
        display: flex;
        align-items: center;
        gap: 2px;
    }

    .stat-gold { color: #c5a059; }
    .stat-atk { color: #ef5350; font-weight: bold; }
    .stat-def { color: #42a5f5; font-weight: bold; }
</style>
