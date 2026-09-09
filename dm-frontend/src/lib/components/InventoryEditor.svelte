<script>
    export let inventory = [];
    export let onAdd = (item) => {};
    export let onRemove = (index) => {};

    let newItem = {
        nom: '',
        prix: 0,
        type: 'arme',
        bonusDegats: 0,
        bonusArmure: 0,
        desDegats: 'd6',
        isConsumable: false,
    };

    let showForm = false;

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

    function resetItemForm() {
        newItem = {
            nom: '',
            prix: 0,
            type: 'arme',
            bonusDegats: 0,
            bonusArmure: 0,
            desDegats: 'd6',
            isConsumable: false,
        };
    }

    function addItem() {
        if (!newItem.nom.trim()) return;
        onAdd(buildItem());
        resetItemForm();
        showForm = false;
    }

    function itemIcon(item) {
        if (item.bonusDegats || item.desDegats) return '⚔️';
        if (item.bonusArmure) return '🛡️';
        if (item.isConsumable) return '🧪';
        return '📦';
    }
</script>

<div class="dnd-section inv-editor">
    <h3 class="section-title"><span>🎒</span> Inventaire</h3>

    <div class="inv-list custom-scrollbar">
        {#each inventory as item, i}
            <div class="inv-item">
                <span class="inv-icon">{itemIcon(item)}</span>
                <div class="inv-info">
                    <span class="inv-name">{item.nom}</span>
                    <div class="inv-badges">
                        {#if item.prix}<span class="badge gold">💰 {item.prix}p</span>{/if}
                        {#if item.bonusDegats}<span class="badge atk">⚔️ +{item.bonusDegats}</span>{/if}
                        {#if item.desDegats}<span class="badge dmg">🎲 1{item.desDegats}</span>{/if}
                        {#if item.bonusArmure}<span class="badge def">🛡️ +{item.bonusArmure}</span>{/if}
                        {#if item.isConsumable}<span class="badge con">🧪</span>{/if}
                    </div>
                </div>
                <button class="btn-remove" on:click={() => onRemove(i)}>✕</button>
            </div>
        {:else}
            <div class="empty-inv">
                <span>Vide</span>
            </div>
        {/each}
    </div>

    {#if showForm}
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
                <div class="mini-field">
                    <label>Prix</label>
                    <input class="form-input-sm" type="number" bind:value={newItem.prix} min="0" />
                </div>
                {#if newItem.type === 'arme'}
                    <div class="mini-field">
                        <label>Dés de dégâts</label>
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
                <button class="btn-cancel" on:click={() => showForm = false}>Annuler</button>
                <button class="btn-add" on:click={addItem}>Ajouter</button>
            </div>
        </div>
    {:else}
        <button class="btn-toggle-form" on:click={() => showForm = true}>
            + Ajouter un objet
        </button>
    {/if}
</div>

<style>
    .inv-editor {
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

    .inv-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 240px;
        overflow-y: auto;
    }

    .inv-item {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 8px 12px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.08);
        border-radius: 8px;
        transition: border-color 0.2s ease;
    }

    .inv-item:hover {
        border-color: rgba(197, 160, 89, 0.25);
    }

    .inv-icon { font-size: 1.1rem; flex-shrink: 0; }

    .inv-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .inv-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.85rem;
    }

    .inv-badges {
        display: flex;
        gap: 6px;
        flex-wrap: wrap;
    }

    .badge {
        font-family: 'Alegreya', serif;
        font-size: 0.65rem;
        padding: 1px 6px;
        border-radius: 4px;
    }

    .badge.gold { background: rgba(197, 160, 89, 0.15); color: #c5a059; }
    .badge.atk { background: rgba(239, 68, 68, 0.15); color: #ef5350; }
    .badge.def { background: rgba(66, 165, 245, 0.15); color: #42a5f5; }
    .badge.con { background: rgba(46, 125, 50, 0.15); color: #66bb6a; }
    .badge.dmg { background: rgba(139, 92, 246, 0.15); color: #c4b5fd; }

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

    .empty-inv {
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
    }

    .form-input:focus {
        border-color: #c5a059;
    }

    .form-row {
        display: flex;
        gap: 10px;
        align-items: center;
    }

    .type-row {
        display: flex;
        gap: 6px;
    }

    .type-btn {
        flex: 1;
        padding: 7px 6px;
        background: rgba(0, 0, 0, 0.3);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #7a6f5f;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.75rem;
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
        padding: 6px 8px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.85rem;
        text-align: center;
        outline: none;
    }

    .form-check {
        display: flex;
        align-items: center;
        gap: 6px;
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.8rem;
        cursor: pointer;
    }

    .form-check input[type="checkbox"] {
        accent-color: #c5a059;
    }

    .mini-field {
        display: flex;
        align-items: center;
        gap: 4px;
    }

    .mini-field label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.7rem;
    }

    .form-input-sm {
        width: 60px;
        padding: 6px 8px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 4px;
        color: #e8e0d4;
        font-family: 'Alegreya', serif;
        font-size: 0.85rem;
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
