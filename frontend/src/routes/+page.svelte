<script>
    import '../app.css';
    import { onMount } from 'svelte';
    import { gameState, connect, sendAction, sendChat, finalizeCharacter, tryRestoreSession, connectionStatus, loginError, disconnect } from '../lib/stores/game';
    import LoginPage from '../lib/components/organisms/pages/LoginPage.svelte';
    import Onboarding from '../lib/components/organisms/Onboarding.svelte';
    import GamePage from '../lib/components/organisms/pages/GamePage.svelte';
    import ToastContainer from '../lib/components/organisms/ToastContainer.svelte';
    import DeathOverlay from '../lib/components/organisms/DeathOverlay.svelte';

    let pseudo = '';
    let restoring = true;

    onMount(() => {
        if (!tryRestoreSession()) {
            restoring = false;
            return;
        }
        const unsub = connectionStatus.subscribe((status) => {
            if (status === 'connected' || status === 'disconnected') {
                restoring = false;
                unsub();
            }
        });
    });

    function join() {
        if (pseudo.trim()) {
            connect(pseudo.trim(), pseudo.trim(), 'Guerrier');
        }
    }

    function logout() {
        disconnect();
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

    function sellItem(index) {
        sendAction({ type: 'sell_item', item_index: index });
    }

    function buyItem(index) {
        sendAction({ type: 'buy_item', item_index: index });
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
    {#if restoring && !$gameState.me}
        <div class="login-view">
            <div class="login-branding">
                <h1 class="brand-title">
                    <span class="brand-icon">⚔️</span>
                    Table de Jeu
                </h1>
                <p class="brand-subtitle">Reconnexion en cours...</p>
            </div>
        </div>
    {:else if !$gameState.me}
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
                connecting={$connectionStatus === 'connecting' || $connectionStatus === 'reconnecting'}
                error={$loginError}
                onJoin={join}
            />
        </div>
    {:else}
        {#if $gameState.newChar}
            <Onboarding pseudo={$gameState.me || ''} onFinalize={finalizeCharacter} />
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
                    onSellItem={sellItem}
                    onBuyItem={buyItem}
                    onSendChat={sendChat}
                    onAcceptQuest={acceptQuest}
                    onCompleteQuest={completeQuest}
                    onLogout={logout}
                />
            </div>
        {/if}
    {/if}
</main>

<ToastContainer />
<DeathOverlay />

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

    @media (max-width: 640px) {
        .brand-title {
            font-size: 1.8rem;
        }
        .login-view {
            padding: 24px 8px 40px;
        }
    }
</style>
