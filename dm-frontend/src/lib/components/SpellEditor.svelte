<script>
    export let spells = [];
    export let onAdd = (spell) => {};
    export let onRemove = (index) => {};

    let newSpell = {
        nom: '',
        niveauSort: 1,
        ecoleMagie: '',
        portee: 'Courte',
        duree: 'Instant',
    };

    let showForm = false;

    const ecoles = ['Évocations', 'Guérison', 'Enchantement', 'Illusion', 'Nécromancie', 'Transmutation', 'Abjuration', 'Divination'];
    const portees = ['Contact', 'Courte', 'Moyenne', 'Longue', 'Illimitée'];
    const durees = ['Instant', '1 round', '1 minute', '10 minutes', '1 heure', 'Permanent'];

    function addSpell() {
        if (!newSpell.nom.trim()) return;
        onAdd({ ...newSpell });
        newSpell = { nom: '', niveauSort: 1, ecoleMagie: '', portee: 'Courte', duree: 'Instant' };
        showForm = false;
    }
</script>

<div class="dnd-section spell-editor">
    <h3 class="section-title"><span>✨</span> Sorts</h3>

    <div class="spell-list custom-scrollbar">
        {#each spells as spell, i}
            <div class="spell-item">
                <span class="spell-icon">🔥</span>
                <div class="spell-info">
                    <span class="spell-name">{spell.nom}</span>
                    <div class="spell-meta">
                        {#if spell.ecoleMagie}<span class="badge school">{spell.ecoleMagie}</span>{/if}
                        {#if spell.niveauSort}<span class="badge level">Nv.{spell.niveauSort}</span>{/if}
                        {#if spell.portee}<span class="badge range">{spell.portee}</span>{/if}
                    </div>
                </div>
                <button class="btn-remove" on:click={() => onRemove(i)}>✕</button>
            </div>
        {:else}
            <div class="empty-spells">
                <span>Aucun sort</span>
            </div>
        {/each}
    </div>

    {#if showForm}
        <div class="add-form fade-in">
            <input class="form-input" bind:value={newSpell.nom} placeholder="Nom du sort" />
            <div class="form-row">
                <div class="mini-field">
                    <label>Niveau</label>
                    <input class="form-input-sm" type="number" bind:value={newSpell.niveauSort} min="1" max="9" />
                </div>
                <div class="mini-field flex-1">
                    <label>École</label>
                    <select class="form-select" bind:value={newSpell.ecoleMagie}>
                        <option value="">—</option>
                        {#each ecoles as ecole}
                            <option value={ecole}>{ecole}</option>
                        {/each}
                    </select>
                </div>
            </div>
            <div class="form-row">
                <div class="mini-field flex-1">
                    <label>Portée</label>
                    <select class="form-select" bind:value={newSpell.portee}>
                        {#each portees as p}
                            <option value={p}>{p}</option>
                        {/each}
                    </select>
                </div>
                <div class="mini-field flex-1">
                    <label>Durée</label>
                    <select class="form-select" bind:value={newSpell.duree}>
                        {#each durees as d}
                            <option value={d}>{d}</option>
                        {/each}
                    </select>
                </div>
            </div>
            <div class="form-actions">
                <button class="btn-cancel" on:click={() => showForm = false}>Annuler</button>
                <button class="btn-add" on:click={addSpell}>Ajouter</button>
            </div>
        </div>
    {:else}
        <button class="btn-toggle-form" on:click={() => showForm = true}>
            + Ajouter un sort
        </button>
    {/if}
</div>

<style>
    .spell-editor {
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

    .spell-list {
        display: flex;
        flex-direction: column;
        gap: 6px;
        max-height: 200px;
        overflow-y: auto;
    }

    .spell-item {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 8px 12px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.08);
        border-radius: 8px;
        transition: border-color 0.2s ease;
    }

    .spell-item:hover {
        border-color: rgba(197, 160, 89, 0.25);
    }

    .spell-icon { font-size: 1.1rem; flex-shrink: 0; }

    .spell-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
    }

    .spell-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.85rem;
    }

    .spell-meta {
        display: flex;
        gap: 4px;
        flex-wrap: wrap;
    }

    .badge {
        font-family: 'Alegreya', serif;
        font-size: 0.6rem;
        padding: 1px 6px;
        border-radius: 4px;
    }

    .badge.school { background: rgba(147, 51, 234, 0.15); color: #a78bfa; }
    .badge.level { background: rgba(197, 160, 89, 0.15); color: #c5a059; }
    .badge.range { background: rgba(66, 165, 245, 0.15); color: #42a5f5; }

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

    .empty-spells {
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

    .mini-field {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .mini-field label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.65rem;
    }

    .flex-1 { flex: 1; }

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

    .form-select {
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

    .form-select:focus {
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
