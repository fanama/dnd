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
    export let onUseItem = (idx) => {};
    export let onLootItem = (idx) => {};
</script>

<div id="game" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
    <!-- Header Banner -->
    <div class="flex flex-col md:flex-row justify-between items-center dnd-card p-8 gap-6 text-white bg-gradient-to-r from-dnd-panel via-dnd-panel to-dnd-accent/30 rounded-2xl shadow-xl">
        <div class="flex items-center gap-5">
            <div class="w-14 h-14 bg-dnd-gold rounded-full flex items-center justify-center text-dnd-dark font-bold text-2xl shadow-lg ring-4 ring-dnd-gold/30 shrink-0">👤</div>
            <div class="flex flex-col gap-0.5">
                <p class="text-xs uppercase tracking-widest text-stone-400 font-medieval">Aventurier</p>
                <h2 class="text-2xl sm:text-3xl font-bold leading-tight">Joueur : <span class="text-dnd-gold font-cinzel">{gameState.me}</span></h2>
            </div>
        </div>
        <div class="flex items-center gap-3 bg-dnd-dark/70 px-8 py-3.5 rounded-full border border-dnd-gold/30 shadow-inner">
            <span class="text-2xl">📍</span>
            <h3 class="text-xl font-medieval text-dnd-gold tracking-wide">{gameState.location}</h3>
        </div>
    </div>

    <!-- Main Content Layout -->
    <div class="grid grid-cols-2 gap-10 ">
        <!-- Sidebar: Character Info -->
        <aside class="lg:col-span-4 flex flex-col gap-8">
            {#if myStats}
                <div class="transform transition-transform duration-200 hover:scale-[1.01]">
                    <CharacterSheet stats={myStats} />
                </div>
            {/if}

            <!-- Inventory Section (Moved from Main Workspace) -->
            <div class="dnd-card p-6 text-white flex flex-col ">
                <h3 class="text-dnd-gold font-medieval text-xl flex items-center gap-3 border-b border-stone-700/50 pb-3">
                    <span class="text-2xl opacity-80">🎒</span> Sac de Voyage
                </h3>
                <div class="flex flex-col gap-3 mt-4 max-h-[320px] overflow-y-auto pr-2 custom-scrollbar">
                    {#if myStats && myStats.inventaire && myStats.inventaire.length > 0}
                        {#each myStats.inventaire as item, i}
                            <InventoryItem {item} index={i} onUse={onUseItem} />
                        {/each}
                    {:else}
                        <div class="text-center py-12 px-4 border-2 border-dashed border-stone-700/60 rounded-xl bg-dnd-dark/20">
                            <span class="text-stone-500 italic block font-alegreya text-lg">Votre sac est vide...</span>
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Move / Travel Options -->
            <div class="dnd-card p-6 m-6 text-white space-y-5">
                <h3 class="text-dnd-gold font-medieval text-xl flex items-center gap-3 border-b border-stone-700/50 pb-3">
                    <span class="text-2xl opacity-80">🗺️</span> Ordre de Marche
                </h3>
                <div class="grid grid-cols-1 gap-3.5 pt-1">
                    <Button onClick={() => onMove('Taverne')} class="w-full py-3.5 text-base tracking-wide shadow-md">Rallier la Taverne</Button>
                    <Button variant="danger" onClick={() => onMove('Donjon')} class="w-full py-3.5 text-base tracking-wide shadow-md">S'enfoncer dans le Donjon</Button>
                </div>
            </div>
        </aside>

        <!-- Main Workspace -->
        <main class="lg:col-span-8 flex flex-col gap-8 xl:gap-10">
            <!-- Grid for Interactive Panels -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
                <!-- Left Panel: Exploration -->
                <div class="flex flex-col gap-8">
                    <div class="transform transition-transform duration-200 hover:scale-[1.01]">
                        <LocationExplorer 
                            locationObjects={gameState.currentLocationObjects} 
                            onLoot={onLootItem} 
                        />
                    </div>
                </div>

                <!-- Right Panel: Players Roster -->
                <div class="flex flex-col">
                    <div class="dnd-card p-6 text-white h-full flex flex-col">
                        <h3 class="text-dnd-gold font-medieval text-xl mb-5 flex items-center gap-3 border-b border-stone-700/50 pb-3">
                            <span class="text-2xl opacity-80">👥</span> Chroniques des Héros
                        </h3>
                        <div class="grid grid-cols-1 gap-4 overflow-y-auto max-h-[520px] pr-1 custom-scrollbar">
                            {#each Object.entries(gameState.players) as [p, v]}
                                <div class="bg-dnd-dark/40 border-2 {p === gameState.me ? 'border-green-500/60 ring-2 ring-green-500/20' : 'border-stone-700/70'} p-5 rounded-xl transition-all hover:border-dnd-gold shadow-sm flex flex-col gap-3">
                                    <div class="flex justify-between items-start gap-2">
                                        <strong class="text-dnd-gold text-lg font-cinzel leading-snug">{v.nom}</strong>
                                        <span class="text-xs px-2.5 py-1 bg-dnd-accent/80 rounded-md text-stone-300 font-medieval shrink-0 border border-stone-600/40">{v.lieu}</span>
                                    </div>
                                    <HPBar current={v.pv} max={v.pv} />
                                    {#if p !== gameState.me && v.lieu === gameState.location}
                                        <Button variant="danger" onClick={() => onHit(p)} class="mt-2 w-full py-2.5 text-sm shadow">Attaquer</Button>
                                    {/if}
                                </div>
                            {/each}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Chat Box Section -->
            <div class="transform transition-transform duration-200 hover:scale-[1.002]">
                <ChatBox logs={gameState.logs} />
            </div>
        </main>
    </div>
</div>
