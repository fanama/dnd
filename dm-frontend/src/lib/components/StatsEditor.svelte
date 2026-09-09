<script>
    export let stats = {};
    export let pv = 0;
    export let maxPv = 0;
    export let isNpc = false;
    export let equipement = { arme: null, armure: null };
    export let onSave = (stats) => {};
    export let onSetPv = (pv) => {};

    let editStats = {};
    let editPv = pv;
    let dirty = false;

    $: {
        if (!dirty) {
            editStats = {
                nom: stats?.nom ?? '',
                background: stats?.background ?? '',
                force: stats?.force ?? 0,
                constitution: stats?.constitution ?? 0,
                vitesse: stats?.vitesse ?? 0,
                charisme: stats?.charisme ?? 0,
                savoir: stats?.savoir ?? 0,
                instinct: stats?.instinct ?? 0,
            };
            editPv = pv;
        }
    }

    function mod(stat) {
        return Math.floor(((Number(stat) || 10) - 10) / 2);
    }

    function signedMod(stat) {
        const m = mod(stat);
        return m > 0 ? `+${m}` : `${m}`;
    }

    function normalizeName(name) {
        return (name || '').toLowerCase().normalize('NFD').replace(/[\u0300-\u036f]/g, '');
    }

    function weaponSides(item) {
        if (!item) return 2;
        const n = normalizeName(item.nom);
        if (n.includes('dague')) return 4;
        if (n.includes('arc') || n.includes('arbalete') || n.includes('carquois')) return 8;
        if (n.includes('epee')) return 6;
        return 6;
    }

    function weaponRanged(item) {
        if (!item) return false;
        const n = normalizeName(item.nom);
        return n.includes('arc') || n.includes('arbalete') || n.includes('carquois');
    }

    $: derivedMaxPv = Math.floor((Number(editStats.constitution) || 10) * 10);
    $: derivedAC = 10 + mod(editStats.vitesse) + (equipement?.armure?.bonusArmure || 0);
    $: attackStat = weaponRanged(equipement?.arme) ? editStats.vitesse : editStats.force;
    $: derivedAttack = mod(attackStat) + (equipement?.arme?.bonusDegats || 0);
    $: damageDice = `${weaponSides(equipement?.arme) ? `1d${weaponSides(equipement?.arme)}` : '1d2'}${derivedAttack >= 0 ? `+${mod(attackStat)}` : mod(attackStat)}`;

    const statFields = [
        { key: 'force', label: 'Force', icon: '💪' },
        { key: 'constitution', label: 'Constitution', icon: '🫀' },
        { key: 'vitesse', label: 'Vitesse', icon: '👟' },
        { key: 'savoir', label: 'Savoir', icon: '📚' },
        { key: 'instinct', label: 'Instinct', icon: '👁️' },
        { key: 'charisme', label: 'Charisme', icon: '✨' },
    ];

    function markDirty() {
        dirty = true;
    }

    function saveStats() {
        onSave({
            nom: (editStats.nom || '').trim(),
            background: (editStats.background || '').trim(),
            force: Number(editStats.force) || 0,
            constitution: Number(editStats.constitution) || 0,
            vitesse: Number(editStats.vitesse) || 0,
            charisme: Number(editStats.charisme) || 0,
            savoir: Number(editStats.savoir) || 0,
            instinct: Number(editStats.instinct) || 0,
        });
        dirty = false;
    }

    function savePv() {
        onSetPv(Number(editPv) || 0);
    }
</script>

<div class="dnd-section stats-editor">
    <h3 class="section-title"><span>📊</span> Statistiques</h3>

    <div class="identity-fields">
        <div class="field">
            <label class="field-label">⚜️ Nom du personnage</label>
            <input
                class="form-input"
                type="text"
                bind:value={editStats.nom}
                placeholder="Nom"
                on:input={markDirty}
            />
        </div>
        {#if isNpc}
            <div class="field">
                <label class="field-label">🎭 Classe / Rôle</label>
                <input
                    class="form-input"
                    type="text"
                    bind:value={editStats.background}
                    placeholder="Classe (ex: Garde)"
                    on:input={markDirty}
                />
            </div>
        {/if}
    </div>

    <div class="derived-grid">
        <div class="derived-card">
            <span class="derived-icon">🛡️</span>
            <span class="derived-value">{derivedAC}</span>
            <span class="derived-label">CA</span>
        </div>
        <div class="derived-card">
            <span class="derived-icon">🎯</span>
            <span class="derived-value">{derivedAttack >= 0 ? `+${derivedAttack}` : derivedAttack}</span>
            <span class="derived-label">Attaque</span>
        </div>
        <div class="derived-card">
            <span class="derived-icon">⚔️</span>
            <span class="derived-value">{damageDice}</span>
            <span class="derived-label">Dégâts</span>
        </div>
        <div class="derived-card">
            <span class="derived-icon">❤️</span>
            <span class="derived-value">{derivedMaxPv}</span>
            <span class="derived-label">PV Max</span>
        </div>
    </div>

    <div class="stats-grid">
        {#each statFields as field}
            <div class="stat-field">
                <label class="stat-label">
                    <span class="stat-icon">{field.icon}</span>
                    {field.label}
                    <span class="stat-mod">{signedMod(editStats[field.key])}</span>
                </label>
                <input
                    class="stat-input"
                    type="number"
                    bind:value={editStats[field.key]}
                    min="0"
                    max="30"
                    on:input={markDirty}
                />
            </div>
        {/each}
    </div>

    {#if dirty}
        <div class="dirty-hint">
            <span>✏️ Modifications non sauvegardées</span>
        </div>
    {/if}

    <button class="btn-save" on:click={saveStats}>
        💾 Sauvegarder les Stats
    </button>

    <div class="pv-section">
        <h3 class="section-title"><span>❤️</span> Points de Vie</h3>
        <div class="pv-row">
            <div class="pv-bar-wrap">
                <div class="pv-bar">
                    <div class="pv-bar-fill" style="width: {derivedMaxPv ? (editPv / derivedMaxPv * 100) : 100}%"></div>
                </div>
                <span class="pv-text">{editPv} / {derivedMaxPv}</span>
            </div>
            <input
                class="stat-input pv-input"
                type="number"
                bind:value={editPv}
                min="0"
                on:input={markDirty}
            />
            <button class="btn-save-sm" on:click={savePv}>OK</button>
        </div>
    </div>
</div>

<style>
    .stats-editor {
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .section-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.95rem;
        margin: 0 0 12px 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.12);
    }

    /* Identity fields */
    .identity-fields {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 10px;
    }

    .field {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .field-label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.7rem;
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
        transition: border-color 0.2s ease;
    }

    .form-input:focus {
        border-color: #c5a059;
    }

    /* Derived stats preview */
    .derived-grid {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 8px;
    }

    .derived-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 2px;
        padding: 10px 6px;
        background: rgba(197, 160, 89, 0.08);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 8px;
    }

    .derived-icon {
        font-size: 1rem;
        line-height: 1;
    }

    .derived-value {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1.4rem;
        font-weight: 700;
        line-height: 1.1;
    }

    .derived-label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.65rem;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    /* Stats grid */
    .stats-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 10px;
    }

    .stat-field {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .stat-label {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.75rem;
        display: flex;
        align-items: center;
        gap: 4px;
    }

    .stat-mod {
        font-family: 'Cinzel', serif;
        font-size: 0.65rem;
        font-weight: 700;
        color: #c5a059;
        background: rgba(197, 160, 89, 0.1);
        border: 1px solid rgba(197, 160, 89, 0.25);
        border-radius: 100px;
        padding: 0 6px;
        margin-left: auto;
        line-height: 1.4;
    }

    .stat-icon { font-size: 0.85rem; }

    .stat-input {
        width: 100%;
        padding: 8px 10px;
        background: rgba(0, 0, 0, 0.5);
        border: 1px solid rgba(197, 160, 89, 0.2);
        border-radius: 6px;
        color: #e8e0d4;
        font-family: 'Cinzel', serif;
        font-size: 1.1rem;
        font-weight: 700;
        text-align: center;
        outline: none;
        transition: border-color 0.2s ease;
        -moz-appearance: textfield;
    }

    .stat-input::-webkit-outer-spin-button,
    .stat-input::-webkit-inner-spin-button {
        -webkit-appearance: none;
        margin: 0;
    }

    .stat-input:focus {
        border-color: #c5a059;
    }

    .dirty-hint {
        display: flex;
        align-items: center;
        gap: 6px;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.75rem;
        color: #fbbf24;
        padding: 8px 10px;
        background: rgba(251, 191, 36, 0.08);
        border: 1px solid rgba(251, 191, 36, 0.2);
        border-radius: 6px;
    }

    .btn-save {
        width: 100%;
        padding: 10px;
        background: rgba(197, 160, 89, 0.15);
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 8px;
        color: #c5a059;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.9rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-save:hover {
        background: rgba(197, 160, 89, 0.25);
    }

    .btn-save:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }

    .pv-section {
        margin-top: 8px;
    }

    .pv-row {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .pv-bar-wrap {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .pv-bar {
        height: 12px;
        background: rgba(0, 0, 0, 0.5);
        border-radius: 6px;
        overflow: hidden;
    }

    .pv-bar-fill {
        height: 100%;
        max-width: 100%;
        background: linear-gradient(90deg, #22c55e, #16a34a);
        border-radius: 6px;
        transition: width 0.3s ease;
    }

    .pv-text {
        font-family: 'Cinzel', serif;
        color: #e8e0d4;
        font-size: 0.8rem;
        text-align: center;
    }

    .pv-input {
        width: 70px;
    }

    .btn-save-sm {
        padding: 8px 16px;
        background: rgba(197, 160, 89, 0.15);
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 6px;
        color: #c5a059;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.85rem;
        cursor: pointer;
        transition: all 0.2s ease;
    }

    .btn-save-sm:hover {
        background: rgba(197, 160, 89, 0.25);
    }

    @media (max-width: 600px) {
        .stats-grid {
            grid-template-columns: repeat(2, 1fr);
        }

        .derived-grid {
            grid-template-columns: repeat(3, 1fr);
        }

        .identity-fields {
            grid-template-columns: 1fr;
        }

        .pv-row {
            flex-wrap: wrap;
        }
    }
</style>