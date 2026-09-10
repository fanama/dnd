<script>
    export let locations = [];
    export let onAdd = (location, quest) => {};
    export let onEdit = (location, index, quest) => {};
    export let onRemove = (location, index) => {};

    let newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
    let formLoc = null;
    let editing = null;

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

    function itemsToStr(items) {
        return (items || []).map((i) => i.nom).join(', ');
    }

    function buildQuest() {
        return {
            nom: newQuest.nom.trim(),
            objectif: newQuest.objectif.trim(),
            obstacle: newQuest.obstacle.trim(),
            recompense: parseItems(newQuest.recompenseItems),
            information: newQuest.information.trim(),
        };
    }

    function openAdd(loc) {
        newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
        editing = null;
        formLoc = loc.nom;
    }

    function openEdit(loc, index) {
        const q = (loc.quests || [])[index];
        if (!q) return;
        editing = index;
        newQuest = {
            nom: q.nom || '',
            objectif: q.objectif || '',
            obstacle: (typeof q.obstacle === 'string' ? q.obstacle : '') || '',
            recompenseItems: itemsToStr(q.recompense),
            information: q.information || '',
        };
        formLoc = loc.nom;
    }

    function submit(loc) {
        if (!newQuest.nom.trim()) return;
        const quest = buildQuest();
        if (editing !== null) {
            onEdit(loc.nom, editing, quest);
        } else {
            onAdd(loc.nom, quest);
        }
        newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
        editing = null;
        formLoc = null;
    }

    function cancel() {
        newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
        editing = null;
        formLoc = null;
    }
</script>

<div class="dnd-section quest-manager">
    <h2 class="panel-title"><span>🧭</span> Quêtes du Monde</h2>
    <p class="panel-hint">Les quêtes visibles ici le sont aussi par les joueurs sur place.</p>

    <div class="qm-list custom-scrollbar">
        {#each locations as loc}
            <div class="qm-loc">
                    <div class="qm-loc-header">
                        <span class="qm-loc-name">📍 {loc.nom}</span>
                        <span class="qm-loc-count">{loc.quests ? loc.quests.length : 0} quête{loc.quests && loc.quests.length > 1 ? 's' : ''}</span>
                    </div>

                    {#each (loc.quests || []) as quest, qi}
                        <div class="qm-quest">
                            <div class="quest-info">
                                <span class="quest-name">{quest.nom}</span>
                                {#if quest.objectif}
                                    <span class="quest-objectif">{quest.objectif}</span>
                                {/if}
                                {#if quest.obstacle}
                                    <span class="quest-obstacle">⛔ {quest.obstacle}</span>
                                {/if}
                                {#if quest.recompense && quest.recompense.length > 0}
                                    <span class="quest-rewards">
                                        🎁 {quest.recompense.map((r) => r.nom).join(', ')}
                                    </span>
                                {/if}
                                {#if quest.information}
                                    <span class="quest-information">📜 {quest.information}</span>
                                {/if}
                            </div>
                            <button class="btn-edit" title="Modifier" on:click={() => openEdit(loc, qi)}>✏️</button>
                            <button class="btn-remove" title="Supprimer" on:click={() => onRemove(loc.nom, qi)}>✕</button>
                        </div>
                    {/each}

                    {#if formLoc === loc.nom}
                        <div class="add-form fade-in">
                            <input class="form-input" bind:value={newQuest.nom} placeholder="Nom de la quête" />
                            <input class="form-input" bind:value={newQuest.objectif} placeholder="Objectif" />
                            <textarea class="form-input textarea" bind:value={newQuest.obstacle} placeholder="Obstacle (description)"></textarea>
                            <input class="form-input" bind:value={newQuest.recompenseItems} placeholder="Récompenses (objets, séparés par des virgules)" />
                            <textarea class="form-input textarea" bind:value={newQuest.information} placeholder="Information révélée à la complétion"></textarea>
                            <div class="form-actions">
                                <button class="btn-cancel" on:click={cancel}>Annuler</button>
                                <button class="btn-add" on:click={() => submit(loc)}>
                                    {editing !== null ? '💾 Enregistrer' : 'Ajouter'}
                                </button>
                            </div>
                        </div>
                    {:else}
                        <button class="btn-add-quest" on:click={() => openAdd(loc)}>
                            + Ajouter une quête
                        </button>
                    {/if}
                </div>
        {:else}
            <div class="empty-qm">
                <span>Aucun lieu dans le monde...</span>
            </div>
        {/each}
    </div>
</div>

<style>
    .quest-manager {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .panel-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 1rem;
        margin: 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.12);
    }

    .panel-hint {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-size: 0.7rem;
        font-style: italic;
        margin: 0;
    }

    .qm-list {
        display: flex;
        flex-direction: column;
        gap: 10px;
        max-height: 460px;
        overflow-y: auto;
        padding-right: 2px;
    }

    .qm-loc {
        display: flex;
        flex-direction: column;
        gap: 6px;
        padding: 8px;
        background: rgba(26, 20, 16, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.12);
        border-radius: 8px;
    }

    .qm-loc-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .qm-loc-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.8rem;
    }

    .qm-loc-count {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-size: 0.65rem;
    }

    .qm-quest {
        display: flex;
        align-items: flex-start;
        gap: 8px;
        padding: 6px 8px;
        background: rgba(0, 0, 0, 0.25);
        border: 1px solid rgba(139, 92, 246, 0.25);
        border-radius: 6px;
    }

    .quest-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .quest-name {
        font-family: 'MedievalSharp', cursive;
        color: #c4b5fd;
        font-size: 0.78rem;
    }

    .quest-objectif {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
        font-style: italic;
    }

    .quest-obstacle {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
        font-style: italic;
    }

    .quest-rewards {
        font-family: 'Alegreya', serif;
        color: #c5a059;
        font-size: 0.7rem;
    }

    .quest-information {
        font-family: 'Alegreya', serif;
        color: #4ade80;
        font-size: 0.7rem;
    }

    .btn-edit {
        width: 22px;
        height: 22px;
        background: rgba(139, 92, 246, 0.15);
        border: 1px solid rgba(139, 92, 246, 0.3);
        border-radius: 4px;
        color: #c4b5fd;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.65rem;
        transition: all 0.2s ease;
        flex-shrink: 0;
    }

    .btn-edit:hover {
        background: rgba(139, 92, 246, 0.35);
    }

    .btn-remove {
        width: 22px;
        height: 22px;
        background: rgba(185, 28, 28, 0.2);
        border: 1px solid rgba(239, 68, 68, 0.2);
        border-radius: 4px;
        color: #fca5a5;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.65rem;
        transition: all 0.2s ease;
        flex-shrink: 0;
    }

    .btn-remove:hover {
        background: rgba(185, 28, 28, 0.4);
    }

    .btn-add-quest {
        width: 100%;
        padding: 5px;
        background: transparent;
        border: 1px dashed rgba(197, 160, 89, 0.15);
        border-radius: 6px;
        color: #5a5045;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-add-quest:hover {
        border-color: rgba(197, 160, 89, 0.3);
        color: #c5a059;
    }

    .add-form {
        display: flex;
        flex-direction: column;
        gap: 6px;
        padding: 8px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.15);
        border-radius: 6px;
    }

    .form-input {
        width: 100%;
        padding: 6px 8px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.8rem;
        outline: none;
        box-sizing: border-box;
    }

    .form-input:focus {
        border-color: #c5a059;
    }

    .form-input.textarea {
        min-height: 48px;
        resize: vertical;
    }

    .form-actions {
        display: flex;
        gap: 6px;
        justify-content: flex-end;
    }

    .btn-cancel {
        padding: 4px 10px;
        background: transparent;
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        cursor: pointer;
    }

    .btn-add {
        padding: 4px 10px;
        background: rgba(197, 160, 89, 0.15);
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 4px;
        color: #c5a059;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-add:hover {
        background: rgba(197, 160, 89, 0.25);
    }

    .empty-qm {
        text-align: center;
        padding: 12px;
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.75rem;
    }
</style>