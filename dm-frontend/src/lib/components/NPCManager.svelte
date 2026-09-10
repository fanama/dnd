<script>
    export let npcs = {};
    export let locations = [];
    export let onAddNpc = (npc) => {};
    export let onRemoveNpc = (name) => {};
    export let onEditNpc = (name, stats, pv, alignement) => {};
    export let onMoveNpc = (name, location) => {};
    export let onNpcAddItem = (name, item) => {};
    export let onNpcRemoveItem = (name, index) => {};
    export let onBulkRemove = (names) => {};
    export let selectNpc = (name) => {};

    let selected = null;
    let showAdd = false;
    let selecting = false;
    let selectedNames = [];

    let newNpc = {
        nom: '',
        lieu: 'Taverne',
        pv: 100,
        alignement: 'Neutre',
        type: 'pnj',
        count: 1,
    };

    let newItem = { nom: '', prix: 0, type: 'arme', bonusDegats: 0, bonusArmure: 0, desDegats: 'd6', isConsumable: false };
    let showItemForm = false;

    const diceOptions = ['d4', 'd6', 'd8', 'd10', 'd12', 'd20'];

    function setType(t) {
        const pv = t === 'boss' ? 300 : t === 'minion' ? 50 : 100;
        newNpc = { ...newNpc, type: t, pv };
    }

    function addNpc() {
        if (!newNpc.nom.trim()) return;
        onAddNpc({
            nom: newNpc.nom.trim(),
            lieu: newNpc.lieu,
            pv: newNpc.pv,
            alignement: newNpc.alignement,
            type: newNpc.type,
            count: newNpc.count,
        });
        newNpc = { ...newNpc, nom: '', pv: 100, count: 1 };
        showAdd = false;
    }

    function select(name) {
        selected = selected === name ? null : name;
        if (selected) selectNpc(selected);
    }

    function isSelected(name) {
        return selectedNames.includes(name);
    }

    function toggleSelect(name) {
        selectedNames = isSelected(name)
            ? selectedNames.filter((n) => n !== name)
            : [...selectedNames, name];
    }

    function exitSelection() {
        selecting = false;
        selectedNames = [];
    }

    function bulkRemove() {
        if (!selectedNames.length) return;
        if (confirm(`Supprimer ${selectedNames.length} entité(s) sélectionnée(s) ?`)) {
            onBulkRemove([...selectedNames]);
        }
        exitSelection();
    }

    function npcStat(npc, key) {
        return npc.stats ? npc.stats[key] : npc[key];
    }

    function buildItem() {
        const base = {
            nom: newItem.nom.trim(),
            prix: newItem.prix,
            encombrement: 0,
            isConsumable: false,
            bonusDegats: 0,
            bonusArmure: 0,
        };
        if (newItem.type === 'arme') {
            return { ...base, bonusDegats: newItem.bonusDegats, desDegats: newItem.desDegats };
        }
        if (newItem.type === 'armure') {
            return { ...base, bonusArmure: newItem.bonusArmure };
        }
        return { ...base, isConsumable: newItem.isConsumable };
    }

    function addItem() {
        if (!newItem.nom.trim() || !selected) return;
        onNpcAddItem(selected, buildItem());
        newItem = { nom: '', prix: 0, type: 'arme', bonusDegats: 0, bonusArmure: 0, desDegats: 'd6', isConsumable: false };
        showItemForm = false;
    }

    function itemIcon(item) {
        if (item.bonusDegats || item.desDegats) return '⚔️';
        if (item.bonusArmure) return '🛡️';
        if (item.isConsumable) return '🧪';
        return '📦';
    }

    function mobIcon(npc) {
        return npc.mobType === 'boss' ? '🐲' : npc.mobType === 'minion' ? '👹' : '🤝';
    }

    function mobLabel(npc) {
        return npc.mobType === 'boss' ? 'BOSS' : npc.mobType === 'minion' ? 'Minion' : null;
    }
</script>

<div class="dnd-section npc-manager">
    <h2 class="panel-title"><span>👥</span> PNJ & MOB</h2>

    {#if !showAdd}
        <button class="btn-add-npc" on:click={() => showAdd = true}>
            + Ajouter un PNJ / MOB
        </button>
    {:else}
        <div class="add-form fade-in">
            <label class="field-label">Nom</label>
            <input class="form-input" bind:value={newNpc.nom} placeholder="Nom du PNJ ou du MOB" />
            <label class="field-label">Type</label>
            <div class="type-row">
                {#each [
                    { key: 'pnj', label: '🤝 PNJ' },
                    { key: 'boss', label: '🐲 Boss' },
                    { key: 'minion', label: '👹 Minion' },
                ] as t}
                    <button
                        type="button"
                        class="type-btn"
                        class:active={newNpc.type === t.key}
                        on:click={() => setType(t.key)}
                    >
                        {t.label}
                    </button>
                {/each}
            </div>
            {#if newNpc.type === 'minion'}
                <div class="form-row">
                    <div class="mini-field flex-1">
                        <label>Nombre de minions</label>
                        <input class="form-input-sm" type="number" bind:value={newNpc.count} min="1" max="20" />
                    </div>
                </div>
            {/if}
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

    <div class="bulk-bar">
        {#if selecting}
            <span class="bulk-count">{selectedNames.length} sélectionné{selectedNames.length > 1 ? 's' : ''}</span>
            <button class="bulk-delete" disabled={selectedNames.length === 0} on:click={bulkRemove}>
                🗑️ Supprimer
            </button>
            <button class="bulk-cancel" on:click={exitSelection}>✕ Annuler</button>
        {:else}
            <button class="bulk-enter" on:click={() => selecting = true}>🗑️ Supprimer plusieurs</button>
        {/if}
    </div>

    <div class="npc-list custom-scrollbar">
        {#each Object.entries(npcs) as [name, npc]}
            <div class="npc-card" class:boss={npc.mobType === 'boss'} class:minion={npc.mobType === 'minion'}>
                <div class="npc-head-row">
                    {#if selecting}
                        <label class="npc-check">
                            <input type="checkbox" checked={isSelected(name)} on:change={() => toggleSelect(name)} />
                        </label>
                    {/if}
                    <button class="npc-header" on:click={() => selecting ? toggleSelect(name) : select(name)}>
                        <span class="npc-icon">{mobIcon(npc)}</span>
                        <div class="npc-info">
                            <span class="npc-name">{npc.nom}</span>
                            {#if mobLabel(npc)}
                                <span class="npc-badge" class:boss={npc.mobType === 'boss'}>{mobLabel(npc)}</span>
                            {/if}
                            <span class="npc-loc">📍 {npc.lieu} · ❤️ {npc.pv}/{npc.max_pv}</span>
                        </div>
                        <span class="npc-chevron" class:open={selected === name}>▾</span>
                    </button>
                </div>

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
                                    {#if item.desDegats}<span class="item-dice">1{item.desDegats}</span>{/if}
                                    <button class="btn-remove" on:click={() => onNpcRemoveItem(name, i)}>✕</button>
                                </div>
                            {:else}
                                <div class="empty-inv">Aucun objet</div>
                            {/each}

                            {#if showItemForm}
                                <div class="add-form fade-in">
                                    <input class="form-input" bind:value={newItem.nom} placeholder="Nom" />
                                    <div class="type-row">
                                        {#each ['arme', 'armure', 'objet'] as t}
                                            <button
                                                type="button"
                                                class="type-btn"
                                                class:active={newItem.type === t}
                                                on:click={() => newItem.type = t}
                                            >
                                                {t === 'arme' ? '⚔️ Arme' : t === 'armure' ? '🛡️ Armure' : '📦 Objet'}
                                            </button>
                                        {/each}
                                    </div>
                                    <div class="form-row">
                                        <div class="mini-field flex-1">
                                            <label>Prix</label>
                                            <input class="form-input-sm" type="number" bind:value={newItem.prix} min="0" />
                                        </div>
                                        {#if newItem.type === 'arme'}
                                            <div class="mini-field">
                                                <label>Dés dégâts</label>
                                                <select class="form-select-sm" bind:value={newItem.desDegats}>
                                                    {#each diceOptions as d}
                                                        <option value={d}>1{d}</option>
                                                    {/each}
                                                </select>
                                            </div>
                                        {/if}
                                    </div>
                                    {#if newItem.type === 'arme'}
                                        <div class="form-row">
                                            <div class="mini-field">
                                                <label>Bonus ATK</label>
                                                <input class="form-input-sm" type="number" bind:value={newItem.bonusDegats} min="0" />
                                            </div>
                                        </div>
                                    {:else if newItem.type === 'armure'}
                                        <div class="form-row">
                                            <div class="mini-field">
                                                <label>Bonus DEF</label>
                                                <input class="form-input-sm" type="number" bind:value={newItem.bonusArmure} min="0" />
                                            </div>
                                        </div>
                                    {:else}
                                        <div class="form-row">
                                            <label class="form-check">
                                                <input type="checkbox" bind:checked={newItem.isConsumable} />
                                                <span>Consommable</span>
                                            </label>
                                        </div>
                                    {/if}
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
                <span class="empty-icon">👥</span>
                <span class="empty-text">Aucun PNJ ou MOB pour l'instant...</span>
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
        color: #a9a090;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.88rem;
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

    .bulk-bar {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .bulk-enter {
        width: 100%;
        padding: 9px;
        background: rgba(185, 28, 28, 0.08);
        border: 1px dashed rgba(239, 68, 68, 0.3);
        border-radius: 8px;
        color: #fca5a5;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .bulk-enter:hover {
        border-color: rgba(239, 68, 68, 0.5);
        color: #fecaca;
    }

    .bulk-count {
        flex: 1;
        font-family: 'Alegreya', serif;
        color: #e8d9b0;
        font-size: 0.88rem;
    }

    .bulk-delete {
        padding: 7px 14px;
        background: rgba(185, 28, 28, 0.25);
        border: 1px solid rgba(239, 68, 68, 0.4);
        border-radius: 6px;
        color: #fecaca;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .bulk-delete:hover:not(:disabled) {
        background: rgba(185, 28, 28, 0.45);
    }

    .bulk-delete:disabled {
        opacity: 0.4;
        cursor: not-allowed;
    }

    .bulk-cancel {
        padding: 7px 14px;
        background: transparent;
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #a9a090;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
    }

    .bulk-cancel:hover {
        color: #c5a059;
        border-color: rgba(197, 160, 89, 0.4);
    }

    .npc-head-row {
        display: flex;
        align-items: center;
    }

    .npc-check {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 0 8px 0 10px;
        cursor: pointer;
    }

    .npc-check input[type="checkbox"] {
        width: 16px;
        height: 16px;
        accent-color: #dc2626;
        cursor: pointer;
    }

    .npc-card {
        border: 1px solid rgba(197, 160, 89, 0.1);
        border-radius: 8px;
        overflow: hidden;
        max-width: 100%;
        min-width: 0;
        transition: border-color 0.2s ease;
    }

    .npc-card.boss {
        border-color: rgba(239, 68, 68, 0.45);
        background: linear-gradient(135deg, rgba(239, 68, 68, 0.05), transparent);
    }

    .npc-card.minion {
        border-color: rgba(217, 119, 6, 0.35);
    }

    .npc-card:hover {
        border-color: rgba(197, 160, 89, 0.25);
    }

    .npc-header {
        flex: 1;
        min-width: 0;
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

    .npc-icon { font-size: 1.25rem; flex-shrink: 0; }

    .npc-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .npc-name {
        font-family: 'MedievalSharp', cursive;
        color: #e8d9b0;
        font-size: 0.92rem;
        overflow-wrap: anywhere;
    }

    .npc-badge {
        align-self: flex-start;
        padding: 2px 10px;
        background: rgba(217, 119, 6, 0.15);
        border: 1px solid rgba(217, 119, 6, 0.4);
        border-radius: 100px;
        color: #fbbf24;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.68rem;
        letter-spacing: 0.06em;
        white-space: nowrap;
    }

    .npc-badge.boss {
        background: rgba(239, 68, 68, 0.18);
        border-color: rgba(239, 68, 68, 0.5);
        color: #fca5a5;
    }

    .npc-loc {
        font-family: 'Alegreya', serif;
        color: #a9a090;
        font-size: 0.78rem;
        overflow-wrap: anywhere;
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
        color: #a9a090;
        font-size: 0.7rem;
    }

    .flex-1 { flex: 1; }

    .move-row {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .field-label {
        font-family: 'MedievalSharp', cursive;
        color: #a9a090;
        font-size: 0.72rem;
    }

    .danger-actions {
        display: flex;
        justify-content: flex-end;
    }

    .btn-danger {
        padding: 6px 14px;
        background: rgba(185, 28, 28, 0.2);
        border: 1px solid rgba(239, 68, 68, 0.3);
        border-radius: 6px;
        color: #fca5a5;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.78rem;
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
        min-width: 0;
        font-family: 'MedievalSharp', cursive;
        color: #e8d9b0;
        font-size: 0.8rem;
        overflow-wrap: anywhere;
    }

    .item-dice {
        font-family: 'Alegreya', serif;
        font-size: 0.68rem;
        color: #c4b5fd;
        background: rgba(139, 92, 246, 0.15);
        padding: 2px 6px;
        border-radius: 4px;
        white-space: nowrap;
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
        font-size: 0.72rem;
        transition: all 0.2s ease;
        flex-shrink: 0;
    }

    .btn-remove:hover {
        background: rgba(185, 28, 28, 0.4);
    }

    .btn-add-item {
        width: 100%;
        padding: 7px;
        background: transparent;
        border: 1px dashed rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #8b8171;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.76rem;
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
        color: #8b8171;
        font-style: italic;
        font-size: 0.76rem;
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
        color: #8b8171;
        font-style: italic;
        font-size: 0.86rem;
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
        padding: 7px 9px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #f2ead8;
        font-family: 'Alegreya', serif;
        font-size: 0.85rem;
        outline: none;
    }

    .form-input:focus { border-color: #c5a059; }

    .form-row {
        display: flex;
        gap: 8px;
        align-items: center;
    }

    .type-row {
        display: flex;
        gap: 6px;
    }

    .type-btn {
        flex: 1;
        padding: 7px 4px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #a9a090;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.76rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .type-btn.active {
        background: rgba(197, 160, 89, 0.18);
        border-color: #c5a059;
        color: #c5a059;
    }

    .type-btn:hover:not(.active) {
        border-color: rgba(197, 160, 89, 0.4);
        color: #a09080;
    }

    .form-select-sm {
        padding: 5px 7px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #f2ead8;
        font-family: 'Alegreya', serif;
        font-size: 0.8rem;
        text-align: center;
        outline: none;
        -moz-appearance: textfield;
    }

    .form-check {
        display: flex;
        align-items: center;
        gap: 4px;
        font-family: 'Alegreya', serif;
        color: #a9a090;
        font-size: 0.76rem;
        cursor: pointer;
    }

    .form-check input[type="checkbox"] { accent-color: #c5a059; }

    .form-input-sm {
        width: 62px;
        padding: 5px 7px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #f2ead8;
        font-family: 'Alegreya', serif;
        font-size: 0.8rem;
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
        padding: 5px 7px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #f2ead8;
        font-family: 'Alegreya', serif;
        font-size: 0.8rem;
        outline: none;
    }

    .form-actions {
        display: flex;
        gap: 6px;
        justify-content: flex-end;
    }

    .btn-cancel {
        padding: 5px 12px;
        background: transparent;
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #a9a090;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.76rem;
        cursor: pointer;
    }

    .btn-add {
        padding: 5px 12px;
        background: rgba(197, 160, 89, 0.15);
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 4px;
        color: #e8d9b0;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.76rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-add:hover {
        background: rgba(197, 160, 89, 0.25);
    }
</style>