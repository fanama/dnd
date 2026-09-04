<script>
    import Button from '../atoms/Button.svelte';
    export let item = {};
    export let index = 0;
    export let onUse = (idx) => {};

    $: name = typeof item === 'string' ? item : item.nom;
    $: isConsumable = typeof item === 'object' ? item.isConsumable : item.toLowerCase().includes('potion');
</script>

<div class="flex justify-between items-center bg-dnd-panel p-3 rounded-lg border-l-4 border-dnd-gold mb-3 shadow-sm transition-all hover:translate-x-1 hover:bg-dnd-accent">
    <div class="flex flex-col gap-1">
        <span class="font-bold text-dnd-gold">📦 {name}</span>
        <div class="text-[0.7rem] text-stone-400 flex flex-wrap gap-x-3 gap-y-1">
            {#if typeof item === 'object'}
                {#if item.prix}<span class="flex items-center">💰 {item.prix}p</span>{/if}
                {#if item.bonusDegats} <span class="text-dnd-gold font-bold flex items-center">⚔️ +{item.bonusDegats} ATK</span>{/if}
                {#if item.bonusArmure} <span class="text-dnd-gold font-bold flex items-center">🛡️ +{item.bonusArmure} DEF</span>{/if}
            {/if}
        </div>
    </div>
    {#if isConsumable}
        <Button variant="success" onClick={() => onUse(index)} class="px-3 py-1 text-sm h-auto">Boire</Button>
    {/if}
</div>
