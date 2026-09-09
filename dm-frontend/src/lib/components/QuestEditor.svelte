<script>
    export let quests = [];
    export let onAdd = (quest) => {};
    export let onRemove = (index) => {};

    let newQuest = {
        nom: '',
        objectif: '',
        obstacle: '',
        recompense: '',
    };
    let showForm = false;

    function parseItems(str) {
        return (str || '')
            .split(',')
            .map((s) => s.trim())
            .filter(Boolean)
            .map((nom) => ({
                nom,
                prix: 0,
                isConsumable: false,
                bonusDegats: 0,
                bonusArmure: 0,
            }));
    }

    function addQuest() {
        if (!newQuest.nom.trim()) return;
        onAdd({
            nom: newQuest.nom.trim(),
            objectif: newQuest.objectif.trim(),
            obstacle: parseItems(newQuest.obstacle),
            recompense: parseItems(newQuest.recompense),
        });
        newQuest = { nom: '', objectif: '', obstacle: '', recompense: '' };
        showForm = false;
    }
</script>

<div class="dnd-section quest-editor">
    <h3 class="section-title"><span>🧭</span> Quêtes</h3>

    <div class="quest-list custom-scrollbar">
        {#each quests as quest, i}
            <div class="quest-item">
                <div class="quest-info">
                    <span class="quest-name">{quest.nom}</span>
                    {#if quest.objectif}
                        <span class="quest-objectif">{quest.objectif}</span>
                    {/if}
                    {#if quest.obstacle && quest.obstacle.length > 0}
                        <div class="quest-meta">
                            <span class="meta-label">Requis :</span>
                            {#each quest.obstacle as ob}
                                <span class="req-name">❌ {ob.nom}</span>
                            {/each}
                        </div>
                    {/if}
                    {#if quest.recompense && quest.recompense.length > 0}
                        <div class="quest-meta">
                            <span class="meta-label">Récompense :</span>
                            {#each quest.recompense as rew}
                                <span class="req-name reward">🎁 {rew.nom}</span>
                            {/each}
                        </div>
                    {/if}
                </div>
                <button class="btn-remove" on:click={() => onRemove(i)}>✕</button>
            </div>
        {:else}
            <div class="empty-quest">
                <span>Aucune quête</span>
            </div>
        {/each}
    </div>

    {#if showForm}
        <div class="add-form fade-in">
            <input class="form-input" bind:value={newQuest.nom} placeholder="Nom de la quête" />
            <input class="form-input" bind:value={newQuest.objectif} placeholder="Objectif" />
            <input class="form-input" bind:value={newQuest.obstacle} placeholder="Requis (objets, séparés par des virgules)" />
            <input class="form-input" bind:value={newQuest.recompense} placeholder="Récompenses (objets, séparés par des virgules)" />
            <div class="form-actions">
                <button class="btn-cancel" on:click={() => showForm = false}>Annuler</button>
                <button class="btn-add" on:click={addQuest}>Ajouter</button>
            </div>
        </div>
    {:else}
        <button class="btn-toggle-form" on:click={() => showForm = true}>
            + Ajouter une quête
        </button>
    {/if}
</div>

<style>
    .quest-editor {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .section-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.95rem;
        margin: 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.12);
    }

    .quest-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 260px;
        overflow-y: auto;
    }

    .quest-item {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        padding: 8px 12px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(139, 92, 246, 0.2);
        border-radius: 8px;
        transition: border-color 0.2s ease;
    }

    .quest-item:hover {
        border-color: rgba(139, 92, 246, 0.4);
    }

    .quest-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 3px;
        min-width: 0;
    }

    .quest-name {
        font-family: 'MedievalSharp', cursive;
        color: #c4b5fd;
        font-size: 0.85rem;
    }

    .quest-objectif {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.75rem;
        font-style: italic;
    }

    .quest-meta {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        align-items: center;
    }

    .meta-label {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-size: 0.65rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .req-name {
        font-family: 'Alegreya', serif;
        font-size: 0.7rem;
        color: #ef5350;
        background: rgba(239, 68, 68, 0.1);
        padding: 1px 6px;
        border-radius: 4px;
    }

    .req-name.reward {
        color: #c5a059;
        background: rgba(197, 160, 89, 0.12);
    }

    .btn-remove {
        width: 24px;
        height: 24px;
        background: rgba(185, 28, 28, 0.2);
        border: 1px solid rgba(239, 68, 68, 0.2);
        border-radius: 4px;
        color: #fca5a5;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.75rem;
        transition: all 0.2s ease;
        flex-shrink: 0;
    }

    .btn-remove:hover {
        background: rgba(185, 28, 28, 0.4);
    }

    .empty-quest {
        text-align: center;
        padding: 16px;
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.85rem;
    }

    .add-form {
        display: flex;
        flex-direction: column;
        gap: 8px;
        padding: 12px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.15);
        border-radius: 8px;
    }

    .form-input {
        width: 100%;
        padding: 8px 10px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.9rem;
        outline: none;
        box-sizing: border-box;
    }

    .form-input:focus {
        border-color: #c5a059;
    }

    .form-actions {
        display: flex;
        gap: 8px;
        justify-content: flex-end;
    }

    .btn-cancel {
        padding: 6px 14px;
        background: transparent;
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
    }

    .btn-add {
        padding: 6px 14px;
        background: rgba(197, 160, 89, 0.15);
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 6px;
        color: #c5a059;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-add:hover {
        background: rgba(197, 160, 89, 0.25);
    }

    .btn-toggle-form {
        width: 100%;
        padding: 10px;
        background: rgba(197, 160, 89, 0.08);
        border: 1px dashed rgba(197, 160, 89, 0.2);
        border-radius: 8px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.85rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-toggle-form:hover {
        border-color: rgba(197, 160, 89, 0.4);
        color: #c5a059;
    }
</style>