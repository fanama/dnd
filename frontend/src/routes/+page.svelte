<script>
    import '../app.css';
    import { gameState, connect, sendAction } from '../lib/stores/game.js';
    import LoginPage from '../lib/components/organisms/pages/LoginPage.svelte';
    import GamePage from '../lib/components/organisms/pages/GamePage.svelte';

    let pseudo = '';
    let charName = '';
    let charClass = 'Guerrier';

    function join() {
        if (pseudo && charName) {
            connect(pseudo, charName, charClass);
        }
    }

    function move(dest) {
        sendAction({ type: 'move', destination: dest });
    }

    function hit(target) {
        sendAction({ type: 'attack', cible: target });
    }

    function useItem(idx) {
        sendAction({ type: 'use_consumable', item_index: idx });
    }

    function lootItem(idx) {
        sendAction({ type: 'loot', loot_index: idx });
    }

    $: myStats = $gameState.players[$gameState.me];
</script>

<svelte:head>
    <title>D&D - Svelte Edition</title>
    <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@500;700&family=MedievalSharp&family=Alegreya:wght@400;700&display=swap" rel="stylesheet">
</svelte:head>

<main class="min-h-screen bg-dnd-dark text-stone-200 font-alegreya p-6 md:p-12 lg:p-16">
    <div class="max-w-7xl mx-auto space-y-12">
        <h1 class="text-center text-dnd-gold font-cinzel text-4xl md:text-6xl border-b-4 double border-dnd-gold pb-8 mb-4 drop-shadow-md">⚔️ Table de Jeu - Svelte & Go ⚔️</h1>

        {#if !$gameState.me}
            <div class="py-12">
                <LoginPage 
                    bind:pseudo 
                    bind:charName 
                    bind:charClass 
                    onJoin={join} 
                />
            </div>
        {:else}
            <GamePage 
                gameState={$gameState} 
                myStats={myStats} 
                onMove={move} 
                onHit={hit} 
                onUseItem={useItem} 
                onLootItem={lootItem} 
            />
        {/if}
    </div>
</main>
