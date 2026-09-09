<script>
    import Button from '../atoms/Button.svelte';

    export let myQuests = [];
    export let locationQuests = [];
    export let inventory = [];
    export let onAccept = (name) => {};
    export let onComplete = (name) => {};

    $: activeNames = new Set(myQuests.map(q => q.nom));

    function hasItem(name) {
        return inventory.some(i => i.nom === name);
    }

    function obstaclesOk(quest) {
        return !quest.obstacle || quest.obstacle.every(o => hasItem(o.nom));
    }
</script>

<div class="quest-panel">
    <div class="dnd-section">
        <h3 class="section-title">
            <span class="section-icon">🧭</span> Mes Quêtes ({myQuests.length})
        </h3>
        {#if myQuests.length > 0}
            <div class="quest-list custom-scrollbar">
                {#each myQuests as quest}
                    <div class="quest-card active">
                        <div class="quest-header">
                            <span class="quest-name">{quest.nom}</span>
                            <span class="quest-status">🔷 En cours</span>
                        </div>
                        <p class="quest-objectif">{quest.objectif}</p>
                        {#if quest.obstacle && quest.obstacle.length > 0}
                            <div class="quest-requires">
                                <span class="requires-label">Requis :</span>
                                {#each quest.obstacle as ob}
                                    <span class="req-item" class:ok={hasItem(ob.nom)}>
                                        {hasItem(ob.nom) ? '✅' : '❌'} {ob.nom}
                                    </span>
                                {/each}
                            </div>
                        {/if}
                        {#if quest.recompense && quest.recompense.length > 0}
                            <div class="quest-rewards">
                                <span class="rewards-label">Récompense :</span>
                                {#each quest.recompense as rew}
                                    <span class="reward-name">🎁 {rew.nom}</span>
                                {/each}
                            </div>
                        {/if}
                        <Button
                            variant="success"
                            onClick={() => onComplete(quest.nom)}
                            disabled={!obstaclesOk(quest)}
                            className="w-full py-2 text-sm"
                        >
                            🏆 Terminer la quête
                        </Button>
                    </div>
                {/each}
            </div>
        {:else}
            <div class="empty-state">
                <span class="empty-icon">🧭</span>
                <span class="empty-text">Aucune quête active...</span>
                <span class="empty-hint">Acceptez une quête de la région pour commencer</span>
            </div>
        {/if}
    </div>

    <div class="dnd-section">
        <h3 class="section-title">
            <span class="section-icon">📜</span> Quêtes disponibles ici
        </h3>
        {#if locationQuests.length > 0}
            <div class="quest-list custom-scrollbar">
                {#each locationQuests as quest}
                    {#if !activeNames.has(quest.nom)}
                        <div class="quest-card">
                            <div class="quest-header">
                                <span class="quest-name">{quest.nom}</span>
                                <span class="quest-status available">📌 Disponible</span>
                            </div>
                            <p class="quest-objectif">{quest.objectif}</p>
                            {#if quest.obstacle && quest.obstacle.length > 0}
                                <div class="quest-requires">
                                    <span class="requires-label">Requis :</span>
                                    {#each quest.obstacle as ob}
                                        <span class="req-item">{ob.nom}</span>
                                    {/each}
                                </div>
                            {/if}
                            {#if quest.recompense && quest.recompense.length > 0}
                                <div class="quest-rewards">
                                    <span class="rewards-label">Récompense :</span>
                                    {#each quest.recompense as rew}
                                        <span class="reward-name">🎁 {rew.nom}</span>
                                    {/each}
                                </div>
                            {/if}
                            <Button variant="primary" onClick={() => onAccept(quest.nom)} className="w-full py-2 text-sm">
                                ✍️ Accepter la quête
                            </Button>
                        </div>
                    {/if}
                {/each}
            </div>
        {:else}
            <div class="empty-state">
                <span class="empty-icon">📜</span>
                <span class="empty-text">Rien à faire par ici...</span>
                <span class="empty-hint">Le Maître du Donjon peut ajouter des quêtes à ce lieu</span>
            </div>
        {/if}
    </div>
</div>

<style>
    .quest-panel {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .dnd-section {
        background-color: #0a0a0a;
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 12px;
        padding: 24px;
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

    .quest-list {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .quest-card {
        padding: 16px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.15);
        border-radius: 12px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .quest-card.active {
        border-color: rgba(139, 92, 246, 0.35);
        background: rgba(69, 39, 160, 0.1);
    }

    .quest-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        gap: 8px;
        flex-wrap: wrap;
    }

    .quest-name {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1rem;
        font-weight: 700;
    }

    .quest-card.active .quest-name {
        color: #c4b5fd;
    }

    .quest-status {
        font-family: 'Alegreya', serif;
        font-size: 0.7rem;
        background: rgba(139, 92, 246, 0.2);
        color: #c4b5fd;
        padding: 2px 10px;
        border-radius: 100px;
        white-space: nowrap;
    }

    .quest-status.available {
        background: rgba(197, 160, 89, 0.15);
        color: #c5a059;
    }

    .quest-objectif {
        font-family: 'Alegreya', serif;
        color: #a09080;
        font-size: 0.9rem;
        margin: 0;
        font-style: italic;
    }

    .quest-requires,
    .quest-rewards {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 8px;
        font-family: 'Alegreya', serif;
        font-size: 0.8rem;
    }

    .requires-label,
    .rewards-label {
        color: #7a6f5f;
        text-transform: uppercase;
        font-size: 0.65rem;
        letter-spacing: 0.06em;
    }

    .req-item {
        color: #ef5350;
        background: rgba(239, 68, 68, 0.1);
        padding: 2px 8px;
        border-radius: 100px;
    }

    .req-item.ok {
        color: #4ade80;
        background: rgba(34, 197, 94, 0.12);
    }

    .reward-name {
        color: #c5a059;
        background: rgba(197, 160, 89, 0.12);
        padding: 2px 8px;
        border-radius: 100px;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 28px 16px;
        border: 2px dashed rgba(197, 160, 89, 0.15);
        border-radius: 12px;
        background: rgba(0, 0, 0, 0.2);
    }

    .empty-icon {
        font-size: 2rem;
        opacity: 0.3;
    }

    .empty-text {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.9rem;
    }

    .empty-hint {
        font-family: 'Alegreya', serif;
        color: #4a4035;
        font-size: 0.8rem;
    }

    @media (max-width: 480px) {
        .dnd-section {
            padding: 16px;
        }
    }
</style>