<script>
    export let locations = [];
    export let players = {};
    export let onAddItem = (location, item) => {};
    export let onRemoveItem = (location, index) => {};
    export let onEditLocation = (oldName, newName, newBg) => {};
    export let onAddQuest = (location, quest) => {};
    export let onEditQuest = (location, index, quest) => {};
    export let onRemoveQuest = (location, index) => {};

    let expandedLoc = null;
    let newItem = { nom: '', prix: 0, type: 'arme', bonusDegats: 0, bonusArmure: 0, desDegats: 'd6', isConsumable: false };
    let showAddFor = null;
    let editName = '';
    let editBg = '';
    let editingLoc = null;
    let newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
    let showQuestFor = null;
    let editingQuest = null;

    const diceOptions = ['d4', 'd6', 'd8', 'd10', 'd12', 'd20'];

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

    function addItem(loc) {
        if (!newItem.nom.trim()) return;
        onAddItem(loc, buildItem());
        newItem = { nom: '', prix: 0, type: 'arme', bonusDegats: 0, bonusArmure: 0, desDegats: 'd6', isConsumable: false };
        showAddFor = null;
    }

    function startEdit(loc) {
        editingLoc = editingLoc === loc ? null : loc;
        editName = loc;
        editBg = '';
        const found = locations.find((l) => l.nom === loc);
        if (found) editBg = found.background || '';
    }

    function saveEdit(loc) {
        if (editName.trim() && editName.trim() !== loc) {
            onEditLocation(loc, editName.trim(), editBg.trim());
        } else if (editBg.trim()) {
            onEditLocation(loc, loc, editBg.trim());
        }
        editingLoc = null;
    }

    function playersAt(locationName) {
        return Object.entries(players).filter(([, v]) => v.lieu === locationName && !v.role).length;
    }

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

    function addQuest(loc) {
        if (!newQuest.nom.trim()) return;
        const quest = {
            nom: newQuest.nom.trim(),
            objectif: newQuest.objectif.trim(),
            obstacle: newQuest.obstacle.trim(),
            recompense: parseItems(newQuest.recompenseItems),
            information: newQuest.information.trim(),
        };
        if (editingQuest) {
            onEditQuest(loc, editingQuest.index, quest);
        } else {
            onAddQuest(loc, quest);
        }
        newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
        editingQuest = null;
        showQuestFor = null;
    }

    function startQuestEdit(loc, index) {
        const locData = locations.find((l) => l.nom === loc);
        const q = locData && locData.quests ? locData.quests[index] : null;
        if (!q) return;
        editingQuest = { loc, index };
        newQuest = {
            nom: q.nom || '',
            objectif: q.objectif || '',
            obstacle: (typeof q.obstacle === 'string' ? q.obstacle : '') || '',
            recompenseItems: (q.recompense || []).map((i) => i.nom).join(', '),
            information: q.information || '',
        };
        showQuestFor = loc;
    }

    function cancelQuestForm() {
        showQuestFor = null;
        newQuest = { nom: '', objectif: '', obstacle: '', recompenseItems: '', information: '' };
        editingQuest = null;
    }

    function locIcon(nom) {
        if (nom === 'Taverne') return '🏠';
        if (nom === 'Donjon') return '⚔️';
        if (nom === 'Foret Enchantee') return '🌿';
        if (nom === 'Montagne Rocheuse') return '⛰️';
        if (nom === 'Marais Hante') return '👻';
        if (nom === 'Plaine des Conflits') return '🚩';
        if (nom === 'Temple Abandonne') return '🏛️';
        return '📍';
    }

    function itemIcon(item) {
        if (item.bonusDegats || item.desDegats) return '⚔️';
        if (item.bonusArmure) return '🛡️';
        if (item.isConsumable) return '🧪';
        return '📦';
    }
</script>

<div class="dnd-section loc-manager">
    <h2 class="panel-title"><span>🗺️</span> Lieux du Monde</h2>

    <div class="loc-list custom-scrollbar">
        {#each locations as loc}
            <div class="loc-card">
                <button class="loc-header" on:click={() => toggle(loc.nom)}>
                    <span class="loc-icon">{locIcon(loc.nom)}</span>
                    <div class="loc-info">
                        <span class="loc-name">{loc.nom}</span>
                        <span class="loc-desc">{loc.background}</span>
                    </div>
                    <span class="loc-count">{loc.objects.length} objets</span>
                    <span class="loc-people">{playersAt(loc.nom)} 👥</span>
                    <span class="loc-chevron" class:open={expandedLoc === loc.nom}>▾</span>
                </button>

                {#if expandedLoc === loc.nom}
                    <div class="loc-body fade-in">
                        <!-- Edit name/description -->
                        <button
                            class="btn-edit-loc"
                            on:click={() => startEdit(loc.nom)}
                            on:click|stopPropagation
                        >
                            ✏️ Modifier le lieu
                        </button>

                        {#if editingLoc === loc.nom}
                            <div class="edit-form fade-in">
                                <label class="field-label">Nom du lieu</label>
                                <input class="form-input" bind:value={editName} />
                                <label class="field-label">Description</label>
                                <input class="form-input" bind:value={editBg} />
                                <div class="form-actions">
                                    <button class="btn-cancel" on:click={() => editingLoc = null}>Annuler</button>
                                    <button class="btn-add" on:click={() => saveEdit(loc.nom)}>💾 Sauvegarder</button>
                                </div>
                            </div>
                        {/if}

                        <!-- Loot -->
                        {#each loc.objects as item, i}
                            <div class="loc-item">
                                <span class="item-icon">{itemIcon(item)}</span>
                                <span class="item-name">{item.nom}</span>
                                {#if item.desDegats}<span class="item-dice">1{item.desDegats}</span>{/if}
                                <button class="btn-remove" on:click={() => onRemoveItem(loc.nom, i)}>✕</button>
                            </div>
                        {:else}
                            <div class="empty-loc">Aucun objet sur le sol</div>
                        {/each}

                        {#if showAddFor === loc.nom}
                            <div class="add-form fade-in">
                                <input class="form-input" bind:value={newItem.nom} placeholder="Nom de l'objet" />
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
                                    <button class="btn-cancel" on:click={() => showAddFor = null}>Annuler</button>
                                    <button class="btn-add" on:click={() => addItem(loc.nom)}>Ajouter</button>
                                </div>
                            </div>
                        {:else}
                            <button class="btn-add-obj" on:click={() => showAddFor = loc.nom}>
                                + Ajouter un objet
                            </button>
                        {/if}

                        <!-- Quests -->
                        {#each (loc.quests || []) as quest, qi}
                            <div class="loc-item quest">
                                <span class="item-icon">🧭</span>
                                <div class="quest-info">
                                    <span class="item-name">{quest.nom}</span>
                                    {#if quest.objectif}<span class="quest-objectif">{quest.objectif}</span>{/if}
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
                                <button class="btn-edit" on:click={() => startQuestEdit(loc.nom, qi)}>✏️</button>
                                <button class="btn-remove" on:click={() => onRemoveQuest(loc.nom, qi)}>✕</button>
                            </div>
                        {/each}

                        {#if showQuestFor === loc.nom}
                            <div class="add-form fade-in">
                                <input class="form-input" bind:value={newQuest.nom} placeholder="Nom de la quête" />
                                <input class="form-input" bind:value={newQuest.objectif} placeholder="Objectif" />
                                <textarea class="form-input quest-textarea" bind:value={newQuest.obstacle} placeholder="Obstacle (description)"></textarea>
                                <input class="form-input" bind:value={newQuest.recompenseItems} placeholder="Récompenses (objets, séparés par des virgules)" />
                                <textarea class="form-input quest-textarea" bind:value={newQuest.information} placeholder="Information révélée à la complétion"></textarea>
                                <div class="form-actions">
                                    <button class="btn-cancel" on:click={cancelQuestForm}>Annuler</button>
                                    <button class="btn-add" on:click={() => addQuest(loc.nom)}>
                                        {editingQuest ? '💾 Enregistrer' : 'Ajouter'}
                                    </button>
                                </div>
                            </div>
                        {:else}
                            <button class="btn-add-obj" on:click={() => showQuestFor = loc.nom}>
                                + Ajouter une quête
                            </button>
                        {/if}
                    </div>
                {/if}
            </div>
        {/each}
    </div>
</div>

<style>
    .loc-manager {
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

    .loc-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 420px;
        overflow-y: auto;
    }

    .loc-card {
        border: 1px solid rgba(197, 160, 89, 0.1);
        border-radius: 8px;
        overflow: hidden;
        transition: border-color 0.2s ease;
    }

    .loc-card:hover {
        border-color: rgba(197, 160, 89, 0.25);
    }

    .loc-header {
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

    .loc-icon { font-size: 1.1rem; flex-shrink: 0; }

    .loc-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 1px;
        min-width: 0;
    }

    .loc-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.85rem;
    }

    .loc-desc {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-size: 0.65rem;
        font-style: italic;
    }

    .loc-count {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
        white-space: nowrap;
    }

    .loc-people {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
        white-space: nowrap;
    }

    .loc-chevron {
        color: #7a6f5f;
        font-size: 0.8rem;
        transition: transform 0.2s ease;
    }

    .loc-chevron.open {
        transform: rotate(180deg);
    }

    .loc-body {
        padding: 8px 12px;
        background: rgba(0, 0, 0, 0.2);
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .btn-edit-loc {
        width: 100%;
        padding: 6px;
        background: rgba(197, 160, 89, 0.08);
        border: 1px dashed rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.75rem;
        cursor: pointer;
        transition: all 0.2s ease;
        margin-bottom: 4px;
    }

    .btn-edit-loc:hover {
        border-color: rgba(197, 160, 89, 0.4);
        color: #c5a059;
    }

    .edit-form {
        display: flex;
        flex-direction: column;
        gap: 6px;
        padding: 8px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.12);
        border-radius: 6px;
        margin-bottom: 6px;
    }

    .field-label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.65rem;
    }

    .loc-item {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 6px 10px;
        background: rgba(26, 20, 16, 0.4);
        border-radius: 6px;
    }

    .item-icon { font-size: 0.9rem; }

    .item-name {
        flex: 1;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.8rem;
    }

    .item-dice {
        font-family: 'Alegreya', serif;
        font-size: 0.6rem;
        color: #c4b5fd;
        background: rgba(139, 92, 246, 0.15);
        padding: 1px 5px;
        border-radius: 4px;
        white-space: nowrap;
    }

    .loc-item.quest {
        border: 1px solid rgba(139, 92, 246, 0.25);
        background: rgba(69, 39, 160, 0.1);
    }

    .quest-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .quest-objectif {
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

    .quest-obstacle {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.7rem;
        font-style: italic;
    }

    .quest-information {
        font-family: 'Alegreya', serif;
        color: #4ade80;
        font-size: 0.7rem;
    }

    .btn-edit {
        width: 20px;
        height: 20px;
        background: rgba(139, 92, 246, 0.15);
        border: 1px solid rgba(139, 92, 246, 0.3);
        border-radius: 4px;
        color: #c4b5fd;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.6rem;
        transition: all 0.2s ease;
        flex-shrink: 0;
    }

    .btn-edit:hover {
        background: rgba(139, 92, 246, 0.35);
    }

    .form-input.quest-textarea {
        min-height: 48px;
        resize: vertical;
        font-family: 'Alegreya', serif;
    }

    .btn-remove {
        width: 20px;
        height: 20px;
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

    .empty-loc {
        text-align: center;
        padding: 10px;
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.75rem;
    }

    .btn-add-obj {
        width: 100%;
        padding: 6px;
        background: transparent;
        border: 1px dashed rgba(197, 160, 89, 0.15);
        border-radius: 6px;
        color: #5a5045;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.75rem;
        cursor: pointer;
        transition: all 0.2s ease;
        margin-top: 4px;
    }

    .btn-add-obj:hover {
        border-color: rgba(197, 160, 89, 0.3);
        color: #c5a059;
    }

    .add-form {
        display: flex;
        flex-direction: column;
        gap: 6px;
        padding: 8px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.12);
        border-radius: 6px;
        margin-top: 4px;
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

    .type-row {
        display: flex;
        gap: 6px;
    }

    .type-btn {
        flex: 1;
        padding: 6px 4px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
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
        padding: 4px 6px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        text-align: center;
        outline: none;
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

    .mini-field {
        display: flex;
        align-items: center;
        gap: 4px;
    }

    .mini-field label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.6rem;
    }

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