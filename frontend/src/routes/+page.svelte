<script>
    import '../app.css';
    import { gameState, connect, sendAction } from '../lib/stores/game';
    import LoginPage from '../lib/components/organisms/pages/LoginPage.svelte';
    import GamePage from '../lib/components/organisms/pages/GamePage.svelte';

    let pseudo = '';
    let charName = '';
    let charClass = 'Guerrier';
    let connectionError = '';

    function join() {
        if (pseudo.trim() && charName.trim()) {
            connectionError = '';
            connect(pseudo.trim(), charName.trim(), charClass);
        }
    }

    function move(dest) {
        sendAction({ type: 'move', destination: dest });
    }

    function hit(target) {
        sendAction({ type: 'attack', cible: target });
    }

    function castSpell(spell, target) {
        sendAction({ type: 'cast_spell', sort: spell, cible: target });
    }

    function useItem(name) {
        sendAction({ type: 'use_consumable', item_name: name });
    }

    function lootItem(name) {
        sendAction({ type: 'loot', loot_name: name });
    }

    function equipItem(name) {
        sendAction({ type: 'equip_item', item_name: name });
    }

    function unequipItem(slot) {
        sendAction({ type: 'unequip_item', slot: slot });
    }

    function acceptQuest(name) {
        sendAction({ type: 'accept_quest', quest_name: name });
    }

    function completeQuest(name) {
        sendAction({ type: 'complete_quest', quest_name: name });
    }
</script>

<svelte:head>
    <title>D&D - Table de Jeu</title>
    <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@500;700&family=MedievalSharp&family=Alegreya:wght@400;700&display=swap" rel="stylesheet">
</svelte:head>

<main class="app-root">
    {#if !$gameState.me}
        <div class="login-view slide-up">
            <div class="login-branding">
                <h1 class="brand-title">
                    <span class="brand-icon">⚔️</span>
                    Table de Jeu
                </h1>
                <p class="brand-subtitle">Svelte & Go</p>
            </div>

            <LoginPage
                bind:pseudo
                bind:charName
                bind:charClass
                onJoin={join}
            />

            {#if connectionError}
                <div class="error-toast fade-in">
                    {connectionError}
                </div>
            {/if}
        </div>
    {:else}
        <div class="game-view">
            <GamePage
                gameState={$gameState}
                onMove={move}
                onHit={hit}
                onCastSpell={castSpell}
                onUseItem={useItem}
                onLootItem={lootItem}
                onEquip={equipItem}
                onUnequip={unequipItem}
                onAcceptQuest={acceptQuest}
                onCompleteQuest={completeQuest}
            />
        </div>
    {/if}
</main>

<style>
    :global(body) {
        margin: 0;
        padding: 0;
        background: #000;
    }

    .app-root {
        min-height: 100vh;
        min-height: 100dvh;
        background: radial-gradient(ellipse at top, #0d0a07 0%, #000 60%);
        color: #e8e0d4;
    }

    .login-view {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 40px 16px 60px;
        min-height: 100vh;
        min-height: 100dvh;
        justify-content: center;
    }

    .login-branding {
        text-align: center;
        margin-bottom: 8px;
    }

    .brand-title {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 2.5rem;
        margin: 0;
        font-weight: 700;
        letter-spacing: 0.05em;
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .brand-icon {
        font-size: 2rem;
    }

    .brand-subtitle {
        font-family: 'MedievalSharp', cursive;
        color: #5a5045;
        font-size: 1rem;
        margin: 4px 0 0 0;
        letter-spacing: 0.15em;
    }

    .game-view {
        min-height: 100vh;
        min-height: 100dvh;
    }

    .error-toast {
        position: fixed;
        bottom: 24px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(185, 28, 28, 0.9);
        color: #fecdd3;
        padding: 12px 24px;
        border-radius: 8px;
        font-family: 'Alegreya', serif;
        border: 1px solid rgba(239, 68, 68, 0.3);
        backdrop-filter: blur(8px);
        z-index: 100;
    }

    @media (max-width: 640px) {
        .brand-title {
            font-size: 1.8rem;
        }
        .login-view {
            padding: 24px 8px 40px;
        }
    }
</style>
