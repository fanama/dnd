<script>
    import LootItem from '../molecules/LootItem.svelte';
    export let locationObjects = [];
    export let onLoot = (idx) => {};
</script>

<div class="bg-dnd-panel border-2 border-dnd-gold/30 rounded-xl p-6 mb-5 shadow-lg relative overflow-hidden group transition-all hover:border-dnd-gold">
    <div class="absolute -right-4 -top-4 w-16 h-16 bg-dnd-gold/10 rounded-full blur-2xl"></div>
    <h3 class="text-dnd-gold font-medieval text-xl m-0 flex items-center gap-2">
        <span class="text-2xl">🔍</span> Exploration du Lieu
    </h3>
    {#if locationObjects && locationObjects.length > 0}
        <div class="mt-5 grid grid-cols-1 gap-3">
            {#each locationObjects as item, i}
                <LootItem {item} index={i} onLoot={onLoot} />
            {/each}
        </div>
    {:else}
        <div class="mt-6 text-center py-4">
            <span class="text-stone-500 italic block font-alegreya">Rien à ramasser ici...</span>
        </div>
    {/if}
</div>
