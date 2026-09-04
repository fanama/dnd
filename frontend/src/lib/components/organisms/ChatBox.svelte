<script>
    export let logs = [];
    let box;

    $: if (logs && box) {
        setTimeout(() => box.scrollTop = box.scrollHeight, 0);
    }
</script>

<div class="bg-dnd-panel border-4 border-double border-dnd-accent rounded-xl p-6 shadow-2xl relative overflow-hidden group">
    <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-dnd-gold to-transparent opacity-50"></div>
    <div class="border-b-2 border-dnd-accent pb-3 mb-5 flex justify-between items-center">
        <h3 class="text-dnd-gold font-medieval text-2xl m-0 drop-shadow-sm">📜 Journal des Événements</h3>
        <span class="text-xs text-stone-500 font-medieval uppercase tracking-widest">Archives du Monde</span>
    </div>
    <div id="box" bind:this={box} class="h-[320px] overflow-y-auto bg-dnd-paper text-dnd-ink p-5 rounded-lg border-4 border-dnd-gold/40 shadow-inner font-alegreya font-medium leading-relaxed ring-4 ring-black/20">
        {#each logs as log}
            <div class="border-b border-dnd-sepia/20 py-2 hover:bg-dnd-gold/10 transition-colors duration-200 px-2 rounded">
                {log}
            </div>
        {/each}
    </div>
</div>
