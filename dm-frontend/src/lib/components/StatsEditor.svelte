<script>
    export let stats = {};
    export let pv = 0;
    export let maxPv = 0;
    export let onSave = (stats) => {};
    export let onSetPv = (pv) => {};

    let editStats = { ...stats };
    let editPv = pv;

    $: editStats = { ...stats };
    $: editPv = pv;

    const statFields = [
        { key: 'force', label: 'Force', icon: '💪' },
        { key: 'constitution', label: 'Constitution', icon: '🫀' },
        { key: 'vitesse', label: 'Vitesse', icon: '👟' },
        { key: 'savoir', label: 'Savoir', icon: '📚' },
        { key: 'instinct', label: 'Instinct', icon: '👁️' },
        { key: 'charisme', label: 'Charisme', icon: '✨' },
    ];

    function saveStats() {
        onSave({ ...editStats });
    }

    function savePv() {
        onSetPv(Number(editPv));
    }
</script>

<div class="dnd-section stats-editor">
    <h3 class="section-title"><span>📊</span> Statistiques</h3>

    <div class="stats-grid">
        {#each statFields as field}
            <div class="stat-field">
                <label class="stat-label">
                    <span class="stat-icon">{field.icon}</span>
                    {field.label}
                </label>
                <input
                    class="stat-input"
                    type="number"
                    bind:value={editStats[field.key]}
                    min="1"
                    max="30"
                />
            </div>
        {/each}
    </div>

    <button class="btn-save" on:click={saveStats}>
        💾 Sauvegarder les Stats
    </button>

    <div class="pv-section">
        <h3 class="section-title"><span>❤️</span> Points de Vie</h3>
        <div class="pv-row">
            <div class="pv-bar-wrap">
                <div class="pv-bar">
                    <div class="pv-bar-fill" style="width: {maxPv ? (editPv / maxPv * 100) : 100}%"></div>
                </div>
                <span class="pv-text">{editPv} / {maxPv}</span>
            </div>
            <input
                class="stat-input pv-input"
                type="number"
                bind:value={editPv}
                min="0"
                max={maxPv}
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
    }
</style>
