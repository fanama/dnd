<script>
    export let npcs = {};
    export let locations = [];
    export let onAddNpc = (npc) => {};
    export let onRemoveNpc = (name) => {};
    export let onEditNpc = (name, stats, pv, alignement) => {};
    export let onMoveNpc = (name, location) => {};
    export let onNpcAddItem = (name, item) => {};
    export let onNpcRemoveItem = (name, index) => {};
    export let selectNpc = (name) => {};

    let selected = null;
    let showAdd = false;

    let newNpc = {
        nom: '',
        lieu: 'Taverne',
        pv: 100,
        alignement: 'Neutre',
        force: 10,
        constitution: 10,
        vitesse: 10,
        charisme: 10,
        savoir: 10,
        instinct: 10,
        classe: 'PNJ',
    };

    let newItem = { nom: '', prix: 0, isConsumable: false, bonusDegats: 0, bonusArmure: 0 };
    let showItemForm = false;

    function addNpc() {
        if (!newNpc.nom.trim()) return;
        onAddNpc({
            nom: newNpc.nom.trim(),
            lieu: newNpc.lieu,
            pv: newNpc.pv,
            alignement: newNpc.alignement,
            force: newNpc.force,
            constitution: newNpc.constitution,
            vitesse: newNpc.vitesse,
            charisme: newNpc.charisme,
            savoir: newNpc.savoir,
            instinct: newNpc.instinct,
            classe: newNpc.classe,
        });
        newNpc = { ...newNpc, nom: '', pv: 100 };
        showAdd = false;
    }

    function select(name) {
        selected = selected === name ? null : name;
        if (selected) selectNpc(selected);
    }

    function npcStat(npc, key) {
        return npc.stats ? npc.stats[key] : npc[key];
    }

    function addItem() {
        if (!newItem.nom.trim() || !selected) return;
        onNpcAddItem(selected, { ...newItem });
        newItem = { nom: '', prix: 0, isConsumable: false, bonusDegats: 0, bonusArmure: 0 };
        showItemForm = false;
    }

    function itemIcon(item) {
        if (item.bonusDegats) return '⚔️';
        if (item.bonusArmure) return '🛡️';
        if (item.isConsumable) return '🧪';
        return '📦';
    }
</script>

<div class="dnd-section npc-manager">
    <h2 class="panel-title"><span>🤝</span> PNJ</h2>

    {#if !showAdd}
        <button class="btn-add-npc" on:click={() => showAdd = true}>
            + Ajouter un PNJ
        </button>
    {:else}
        <div class="add-form fade-in">
            <label class="field-label">Nom</label>
            <input class="form-input" bind:value={newNpc.nom} placeholder="Nom du PNJ" />
            <div class="form-row">
                <div class="mini-field flex-1">
                    <label>Lieu</label>
                    <select class="form-select" bind:value={newNpc.lieu}>
                        {#each locations as loc}
                            <option value={loc.nom}>{loc.nom}</option>
                        {/each}
                    </select>
                </div>
                <div class="mini-field">
                    <label>PV</label>
                    <input class="form-input-sm" type="number" bind:value={newNpc.pv} min="1" />
                </div>
            </div>
            <label class="field-label">Alignement</label>
            <input class="form-input" bind:value={newNpc.alignement} />
            <div class="form-actions">
                <button class="btn-cancel" on:click={() => showAdd = false}>Annuler</button>
                <button class="btn-add" on:click={addNpc}>Créer</button>
            </div>
        </div>
    {/if}

    <div class="npc-list custom-scrollbar">
        {#each Object.entries(npcs) as [name, npc]}
            <div class="npc-card">
                <button class="npc-header" on:click={() => select(name)}>
                    <span class="npc-icon">🤖</span>
                    <div class="npc-info">
                        <span class="npc-name">{npc.nom}</span>
                        <span class="npc-loc">📍 {npc.lieu} · ❤️ {npc.pv}/{npc.max_pv}</span>
                    </div>
                    <span class="npc-chevron" class:open={selected === name}>▾</span>
                </button>

                {#if selected === name}
                    <div class="npc-body fade-in">
                        <!-- Edit form -->
                        <div class="edit-grid">
                            {#each [
                                { key: 'force', label: 'Force' },
                                { key: 'constitution', label: 'Const.' },
                                { key: 'vitesse', label: 'Vit.' },
                                { key: 'charisme', label: 'Char.' },
                                { key: 'savoir', label: 'Savoir' },
                                { key: 'instinct', label: 'Inst.' },
                            ] as f}
                                <div class="mini-field">
                                    <label>{f.label}</label>
                                    <input
                                        class="form-input-sm"
                                        type="number"
                                        value={npcStat(npc, f.key)}
                                        on:change={(e) => onEditNpc(name, { [f.key]: Number(e.target.value) }, 0, '')}
                                    />
                                </div>
                            {/each}
                        </div>

                        <!-- Move -->
                        <div class="move-row">
                            <label class="field-label">Déplacer vers :</label>
                            <select
                                class="form-select"
                                value={npc.lieu}
                                on:change={(e) => onMoveNpc(name, e.target.value)}
                            >
                                {#each locations as loc}
                                    <option value={loc.nom} selected={npc.lieu === loc.nom}>{loc.nom}</option>
                                {/each}
                            </select>
                        </div>

                        <div class="danger-actions">
                            <button class="btn-danger" on:click={() => onRemoveNpc(name)}>
                                🗑️ Supprimer
                            </button>
                        </div>

                        <!-- Inventory -->
                        <div class="npc-inv">
                            <label class="field-label">Inventaire</label>
                            {#each (npc.inventaire || []) as item, i}
                                <div class="npc-item">
                                    <span class="item-icon">{itemIcon(item)}</span>
                                    <span class="item-name">{item.nom}</span>
                                    <button class="btn-remove" on:click={() => onNpcRemoveItem(name, i)}>✕</button>
                                </div>
                            {:else}
                                <div class="empty-inv">Aucun objet</div>
                            {/each}

                            {#if showItemForm}
                                <div class="add-form fade-in">
                                    <input class="form-input" bind:value={newItem.nom} placeholder="Nom" />
                                    <div class="form-row">
                                        <label class="form-check">
                                            <input type="checkbox" bind:checked={newItem.isConsumable} />
                                            <span>Consommable</span>
                                        </label>
                                        <div class="mini-field">
                                            <label>Prix</label>
                                            <input class="form-input-sm" type="number" bind:value={newItem.prix} min="0" />
                                        </div>
                                    </div>
                                    <div class="form-row">
                                        <div class="mini-field">
                                            <label>ATK</label>
                                            <input class="form-input-sm" type="number" bind:value={newItem.bonusDegats} min="0" />
                                        </div>
                                        <div class="mini-field">
                                            <label>DEF</label>
                                            <input class="form-input-sm" type="number" bind:value={newItem.bonusArmure} min="0" />
                                        </div>
                                    </div>
                                    <div class="form-actions">
                                        <button class="btn-cancel" on:click={() => showItemForm = false}>Annuler</button>
                                        <button class="btn-add" on:click={addItem}>Ajouter</button>
                                    </div>
                                </div>
                            {:else}
                                <button class="btn-add-item" on:click={() => showItemForm = true}>+ Objet</button>
                            {/if}
                        </div>
                    </div>
                {/if}
            </div>
        {:else}
            <div class="empty-state">
                <span class="empty-icon">🤝</span>
                <span class="empty-text">Aucun PNJ pour l'instant...</span>
            </div>
        {/each}
    </div>
</div>

<style>
    .npc-manager {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .panel-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 1.05rem;
        margin: 0;
        padding-bottom: 12px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.15);
    }

    .btn-add-npc {
        width: 100%;
        padding: 10px;
        background: rgba(197, 160, 89, 0.08);
        border: 1px dashed rgba(197, 160, 89, 0.25);
        border-radius: 8px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.85rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-add-npc:hover {
        border-color: rgba(197, 160, 89, 0.4);
        color: #c5a059;
    }

    .npc-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 420px;
        overflow-y: auto;
    }

    .npc-card {
        border: 1px solid rgba(197, 160, 89, 0.1);
        border-radius: 8px;
        overflow: hidden;
        transition: border-color 0.2s ease;
    }

    .npc-card:hover {
        border-color: rgba(197, 160, 89, 0.25);
    }

    .npc-header {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 10px 12px;
        background: rgba(26, 20, 16, 0.6);
        border: none;
        cursor: pointer;
        color: inherit;
        text-align: left;
    }

    .npc-icon { font-size: 1.1rem; flex-shrink: 0; }

    .npc-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 1px;
        min-width: 0;
    }

    .npc-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.85rem;
    }

    .npc-loc {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
    }

    .npc-chevron {
        color: #7a6f5f;
        font-size: 0.8rem;
        transition: transform 0.2s ease;
    }

    .npc-chevron.open {
        transform: rotate(180deg);
    }

    .npc-body {
        padding: 10px 12px;
        background: rgba(0, 0, 0, 0.2);
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .edit-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 6px;
    }

    .mini-field {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .mini-field label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.6rem;
    }

    .flex-1 { flex: 1; }

    .move-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .field-label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.65rem;
    }

    .danger-actions {
        display: flex;
        justify-content: flex-end;
    }

    .btn-danger {
        padding: 5px 12px;
        background: rgba(185, 28, 28, 0.2);
        border: 1px solid rgba(239, 68, 68, 0.3);
        border-radius: 6px;
        color: #fca5a5;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-danger:hover {
        background: rgba(185, 28, 28, 0.4);
    }

    .npc-inv {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .npc-item {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 5px 8px;
        background: rgba(26, 20, 16, 0.4);
        border-radius: 6px;
    }

    .item-icon { font-size: 0.85rem; }

    .item-name {
        flex: 1;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.75rem;
    }

    .btn-remove {
        width: 18px;
        height: 18px;
        background: rgba(185, 28, 28, 0.2);
        border: 1px solid rgba(239, 68, 68, 0.2);
        border-radius: 4px;
        color: #fca5a5;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.6rem;
        transition: all 0.2s ease;
        flex-shrink: 0;
    }

    .btn-remove:hover {
        background: rgba(185, 28, 28, 0.4);
    }

    .btn-add-item {
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

    .btn-add-item:hover {
        border-color: rgba(197, 160, 89, 0.3);
        color: #c5a059;
    }

    .empty-inv {
        text-align: center;
        padding: 8px;
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.7rem;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 24px 16px;
        border: 2px dashed rgba(197, 160, 89, 0.15);
        border-radius: 12px;
        background: rgba(0, 0, 0, 0.2);
    }

    .empty-icon { font-size: 2rem; opacity: 0.3; }

    .empty-text {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.8rem;
    }

    .add-form {
        display: flex;
        flex-direction: column;
        gap: 6px;
        padding: 8px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.12);
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
    }

    .form-input:focus { border-color: #c5a059; }

    .form-row {
        display: flex;
        gap: 8px;
        align-items: center;
    }

    .form-check {
        display: flex;
        align-items: center;
        gap: 4px;
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
        cursor: pointer;
    }

    .form-check input[type="checkbox"] { accent-color: #c5a059; }

    .form-input-sm {
        width: 50px;
        padding: 4px 6px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        text-align: center;
        outline: none;
        -moz-appearance: textfield;
    }

    .form-input-sm::-webkit-outer-spin-button,
    .form-input-sm::-webkit-inner-spin-button {
        -webkit-appearance: none;
    }

    .form-select {
        width: 100%;
        padding: 4px 6px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        outline: none;
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
</style>