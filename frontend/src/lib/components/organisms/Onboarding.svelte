<script>
    export let pseudo = '';
    export let onFinalize = () => {};

    const STATS = ['force', 'constitution', 'vitesse', 'charisme', 'savoir', 'instinct'];
    const STAT_LABELS = {
        force: 'Force',
        constitution: 'Constitution',
        vitesse: 'Vitesse',
        charisme: 'Charisme',
        savoir: 'Savoir',
        instinct: 'Instinct'
    };
    const STAT_ICONS = {
        force: '💪',
        constitution: '🛡️',
        vitesse: '👟',
        charisme: '💬',
        savoir: '📖',
        instinct: '🧠'
    };
    const BONUS_POINTS = 4;

    function hitDiceSides(classe) {
        switch (classe) {
            case 'Magicien': return 6;
            case 'Voleur':
            case 'Clerc':
            case 'Barde':
            case 'Ranger': return 8;
            case 'Guerrier':
            default: return 10;
        }
    }

    const classes = {
        Guerrier: {
            icon: '🛡️',
            desc: 'Force et endurance',
            lore: 'Un rempart de fer et d\'acier, fidèle à sa promesse.',
            base: { force: 15, constitution: 12, vitesse: 10, charisme: 10, savoir: 10, instinct: 10 },
            gear: ['Épée Longue (dégâts +5)', 'Potion de Soin'],
            spells: []
        },
        Magicien: {
            icon: '🔮',
            desc: 'Arcanes et mystère',
            lore: 'L\'esprit est sa lame, la magie est son bouclier.',
            base: { force: 8, constitution: 9, vitesse: 11, charisme: 12, savoir: 16, instinct: 14 },
            gear: ['Bâton Mystique (dégâts +3)', 'Potion de Soin'],
            spells: ['Boule de Feu']
        },
        Voleur: {
            icon: '🗡️',
            desc: 'Agilité et ruse',
            lore: 'L\'ombre murmure son nom avant chaque perfidie.',
            base: { force: 10, constitution: 8, vitesse: 16, charisme: 10, savoir: 11, instinct: 15 },
            gear: ['Dague Empoisonnée (dégâts +4)', 'Potion de Soin'],
            spells: []
        },
        Clerc: {
            icon: '✝️',
            desc: 'Foi et guérison',
            lore: 'Sa lumière réconforte les alliés, son marteau écrase les ténèbres.',
            base: { force: 12, constitution: 14, vitesse: 8, charisme: 15, savoir: 12, instinct: 9 },
            gear: ['Marteau Sacré (dégâts +4)', 'Potion de Soin'],
            spells: ['Soin Divin']
        },
        Barde: {
            icon: '🎵',
            desc: 'Charmes et musique',
            lore: 'Chaque note charme, chaque mot ouvre une porte.',
            base: { force: 9, constitution: 10, vitesse: 13, charisme: 16, savoir: 10, instinct: 12 },
            gear: ['Luth Enchanté (dégâts +2)', 'Potion de Soin'],
            spells: ['Mélodie Envoûtante']
        },
        Ranger: {
            icon: '🏹',
            desc: 'Nature et précision',
            lore: 'La forêt lui a appris à viser juste et à frapper vite.',
            base: { force: 13, constitution: 11, vitesse: 14, charisme: 8, savoir: 9, instinct: 15 },
            gear: ['Arc Long (dégâts +5)', 'Potion de Soin'],
            spells: []
        }
    };

    let step = 1;
    let className = 'Guerrier';
    let charName = '';

    const allocation = { force: 0, constitution: 0, vitesse: 0, charisme: 0, savoir: 0, instinct: 0 };

    $: klass = classes[className] || classes.Guerrier;
    $: pointsLeft = BONUS_POINTS - STATS.reduce((sum, k) => sum + allocation[k], 0);
    $: finalStats = STATS.reduce((acc, k) => {
        acc[k] = (klass.base[k] || 10) + allocation[k];
        return acc;
    }, {});
    $: maxPv = Math.max(1, hitDiceSides(className) + Math.floor(((finalStats.constitution || 10) - 10) / 2));
    $: canGoNext = (step === 1) || (step === 2 && charName.trim().length >= 2);
    $: bonusMod = (v) => {
        const m = Math.floor(((v || 10) - 10) / 2);
        return m > 0 ? `+${m}` : `${m}`;
    };

    function selectClass(name) {
        className = name;
        STATS.forEach(k => allocation[k] = 0);
    }

    function allocate(key, delta) {
        const next = allocation[key] + delta;
        if (next < 0) return;
        if (next > 20) return;
        if (delta > 0 && pointsLeft <= 0) return;
        allocation[key] = next;
    }

    function next() {
        if (canGoNext && step < 3) step += 1;
    }

    function back() {
        if (step > 1) step -= 1;
    }

    function finish() {
        onFinalize({
            nom: charName.trim(),
            background: className,
            ...finalStats
        });
    }
</script>

<div class="onboarding-backdrop">
    <div class="onboarding-card fade-in">
        <!-- Header -->
        <div class="ob-header">
            <span class="ob-icon">⚔️</span>
            <h2 class="ob-title">Forge de l'Héro</h2>
            <p class="ob-subtitle">Façonnez le destin de {pseudo}</p>
        </div>

        <!-- Stepper -->
        <ol class="ob-steps">
            <li class:active={step === 1} class:done={step > 1}><span>1</span> Classe</li>
            <li class:active={step === 2} class:done={step > 2}><span>2</span> Répartition</li>
            <li class:active={step === 3}><span>3</span> Table du Héros</li>
        </ol>

        <!-- Step 1: Classe -->
        {#if step === 1}
            <div class="ob-step">
                <p class="ob-label">1. Choisissez votre classe</p>
                <div class="class-grid">
                    {#each Object.entries(classes) as [name, cls]}
                        <button
                            type="button"
                            class="class-card"
                            class:selected={className === name}
                            on:click={() => selectClass(name)}
                        >
                            <span class="class-icon">{cls.icon}</span>
                            <span class="class-name">{name}</span>
                            <span class="class-desc">{cls.desc}</span>
                        </button>
                    {/each}
                </div>
                <div class="class-preview">
                    <p class="preview-lore">« {klass.lore} »</p>
                    <div class="preview-stats">
                        {#each STATS as k}
                            <div class="preview-row">
                                <span class="ps-label">{STAT_ICONS[k]} {STAT_LABELS[k]}</span>
                                <span class="ps-val">{klass.base[k]}</span>
                                <span class="ps-mod">{bonusMod(klass.base[k])}</span>
                            </div>
                        {/each}
                    </div>
                </div>
            </div>
        {:else if step === 2}
            <!-- Step 2: Nom + répartition -->
            <div class="ob-step">
                <p class="ob-label">2. Nom du personnage</p>
                <input class="name-input" bind:value={charName} placeholder="Ex : Lancelot, Mila, Dorn..." maxlength="24" />

                <p class="ob-label">3. Répartissez vos {BONUS_POINTS} points de destin</p>
                <div class="alloc-grid">
                    {#each STATS as k}
                        <div class="alloc-row">
                            <span class="alloc-name">{STAT_ICONS[k]} {STAT_LABELS[k]}</span>
                            <div class="alloc-controls">
                                <button class="alloc-btn" on:click={() => allocate(k, -1)} disabled={allocation[k] <= 0}>−</button>
                                <span class="alloc-value">{finalStats[k]}</span>
                                <button class="alloc-btn" on:click={() => allocate(k, 1)} disabled={pointsLeft <= 0}>+</button>
                            </div>
                            <span class="alloc-mod">{bonusMod(finalStats[k])}</span>
                        </div>
                    {/each}
                </div>
                <p class="alloc-left" class:empty={pointsLeft === 0}>
                    {pointsLeft > 0 ? `${pointsLeft} point${pointsLeft > 1 ? 's' : ''} restant${pointsLeft > 1 ? 's' : ''}` : 'Destin accompli ✓'}
                </p>
                <div class="pv-preview">❤️ Points de vie : <strong>{maxPv}</strong></div>
            </div>
        {:else}
            <!-- Step 3: Synthèse -->
            <div class="ob-step">
                <p class="ob-label">La Table du Héros</p>
                <div class="sheet">
                    <div class="sheet-head">
                        <span class="sheet-icon">{klass.icon}</span>
                        <div>
                            <div class="sheet-name">{charName.trim()}</div>
                            <div class="sheet-class">{className}</div>
                        </div>
                    </div>
                    <div class="sheet-row">
                        <span>❤️ Point de vie</span><strong>{maxPv}</strong>
                    </div>
                    <div class="sheet-row">
                        <span>🛡️ Défense</span><strong>{10 + Math.floor(((finalStats.vitesse || 10) - 10) / 2)}</strong>
                    </div>
                    {#each STATS as k}
                        <div class="sheet-row">
                            <span>{STAT_ICONS[k]} {STAT_LABELS[k]}</span>
                            <strong>{finalStats[k]} <em>({bonusMod(finalStats[k])})</em></strong>
                        </div>
                    {/each}
                    {#if klass.spells.length}
                        <div class="sheet-row">
                            <span>✨ Sorts de départ</span>
                            <strong>{klass.spells.join(', ')}</strong>
                        </div>
                    {/if}
                    <div class="sheet-row">
                        <span>🎒 Équipement de départ</span>
                        <strong>{klass.gear.join(' · ')}</strong>
                    </div>
                </div>
            </div>
        {/if}

        <!-- Actions -->
        <div class="ob-actions">
            {#if step > 1}
                <button class="ob-btn ghost" on:click={back}>← Retour</button>
            {/if}
            {#if step < 3}
                <button class="ob-btn" on:click={next} disabled={!canGoNext}>Continuer →</button>
            {:else}
                <button class="ob-btn cta" on:click={finish}>⚔️ Entrer dans le monde</button>
            {/if}
        </div>
    </div>
</div>

<style>
    .onboarding-backdrop {
        position: fixed;
        inset: 0;
        z-index: 60;
        display: flex;
        align-items: flex-start;
        justify-content: center;
        padding: 24px 12px 60px;
        overflow-y: auto;
        background: radial-gradient(ellipse at top, #0d0a07 0%, #000 60%);
    }

    .onboarding-card {
        width: 100%;
        max-width: 560px;
        background: linear-gradient(135deg, #0d0d0d 0%, #151210 100%);
        border: 1px solid rgba(197, 160, 89, 0.25);
        border-radius: 16px;
        padding: 28px 26px;
        box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6), 0 0 40px rgba(197, 160, 89, 0.05);
    }

    .ob-header {
        text-align: center;
        margin-bottom: 20px;
        padding-bottom: 16px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.15);
    }
    .ob-icon { font-size: 1.6rem; }
    .ob-title {
        margin: 4px 0 4px;
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 1.5rem;
        letter-spacing: 0.05em;
    }
    .ob-subtitle {
        margin: 0;
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-style: italic;
        font-size: 0.9rem;
    }

    .ob-steps {
        display: flex;
        justify-content: center;
        gap: 8px;
        list-style: none;
        padding: 0;
        margin: 0 0 20px;
    }
    .ob-steps li {
        display: flex;
        align-items: center;
        gap: 5px;
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        color: #5a5045;
        text-transform: uppercase;
        letter-spacing: 0.06em;
    }
    .ob-steps li span {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        border-radius: 50%;
        background: #1a1410;
        border: 1px solid #3a3028;
        color: #7a6f5f;
        font-size: 0.7rem;
    }
    .ob-steps li.active { color: #c5a059; }
    .ob-steps li.active span {
        background: #c5a059;
        color: #14100a;
        border-color: #c5a059;
    }
    .ob-steps li.done span {
        background: rgba(197, 160, 89, 0.2);
        color: #c5a059;
        border-color: rgba(197, 160, 89, 0.4);
    }

    .ob-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.85rem;
        color: #c5a059;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        margin: 0 0 10px;
    }

    .class-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 10px;
        margin-bottom: 14px;
    }
    .class-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 3px;
        padding: 14px 8px;
        background: #1a1410;
        border: 2px solid #3a3028;
        border-radius: 12px;
        cursor: pointer;
        transition: all 0.25s ease;
        text-align: center;
    }
    .class-card:hover {
        border-color: rgba(197, 160, 89, 0.5);
        background: #211c15;
        transform: translateY(-2px);
    }
    .class-card.selected {
        border-color: #c5a059;
        background: linear-gradient(135deg, #1f1a12, #2a2218);
        box-shadow: 0 0 20px rgba(197, 160, 89, 0.15), inset 0 0 20px rgba(197, 160, 89, 0.05);
    }
    .class-icon { font-size: 1.6rem; line-height: 1; }
    .class-name { font-family: 'MedievalSharp', cursive; color: #c5a059; font-size: 0.85rem; font-weight: bold; }
    .class-desc { font-family: 'Alegreya', serif; color: #7a6f5f; font-size: 0.65rem; font-style: italic; }

    .class-preview {
        border: 1px dashed rgba(197, 160, 89, 0.25);
        border-radius: 10px;
        padding: 12px 14px;
        background: rgba(0, 0, 0, 0.25);
    }
    .preview-lore {
        margin: 0 0 8px;
        font-family: 'Alegreya', serif;
        font-style: italic;
        color: #a09080;
        font-size: 0.85rem;
    }
    .preview-stats { display: flex; flex-direction: column; gap: 4px; }
    .preview-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-family: 'Alegreya', serif;
        font-size: 0.85rem;
        color: #e8e0d4;
    }
    .ps-val { color: #c5a059; font-weight: bold; }
    .ps-mod { color: #7a6f5f; font-size: 0.8rem; }

    .name-input {
        width: 100%;
        padding: 12px 14px;
        margin-bottom: 18px;
        background: #14100c;
        border: 1px solid #3a3028;
        border-radius: 8px;
        color: #e8e0d4;
        font-family: 'MedievalSharp', cursive;
        font-size: 1rem;
        outline: none;
    }
    .name-input:focus { border-color: #c5a059; }

    .alloc-grid {
        display: flex;
        flex-direction: column;
        gap: 8px;
        margin-bottom: 10px;
    }
    .alloc-row {
        display: grid;
        grid-template-columns: 1fr auto 1fr;
        align-items: center;
        gap: 12px;
        padding: 8px 12px;
        background: #100d0b;
        border: 1px solid #241e18;
        border-radius: 8px;
    }
    .alloc-name { font-family: 'Alegreya', serif; color: #cbb295; font-size: 0.9rem; }
    .alloc-controls { display: flex; align-items: center; gap: 10px; }
    .alloc-btn {
        width: 30px;
        height: 30px;
        border-radius: 6px;
        border: 1px solid rgba(197, 160, 89, 0.4);
        background: rgba(197, 160, 89, 0.1);
        color: #c5a059;
        font-size: 1.1rem;
        cursor: pointer;
    }
    .alloc-btn:disabled { opacity: 0.35; cursor: not-allowed; }
    .alloc-value { width: 28px; text-align: center; font-family: 'Cinzel', serif; font-weight: 700; color: #e8e0d4; }
    .alloc-mod { text-align: right; font-family: 'Alegreya', serif; color: #7a6f5f; font-size: 0.8rem; }

    .alloc-left {
        margin: 0 0 12px;
        font-family: 'Alegreya', serif;
        font-size: 0.85rem;
        color: #c5a059;
        text-align: center;
    }
    .alloc-left.empty { color: #4ade80; }
    .pv-preview {
        text-align: center;
        font-family: 'Alegreya', serif;
        color: #e8e0d4;
        font-size: 0.95rem;
    }

    .sheet {
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 10px;
        background: rgba(0, 0, 0, 0.3);
        padding: 14px 16px;
    }
    .sheet-head {
        display: flex;
        align-items: center;
        gap: 12px;
        padding-bottom: 12px;
        margin-bottom: 10px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.2);
    }
    .sheet-icon { font-size: 2rem; }
    .sheet-name { font-family: 'Cinzel', serif; color: #c5a059; font-size: 1.3rem; font-weight: 700; }
    .sheet-class { font-family: 'Alegreya', serif; color: #7a6f5f; font-style: italic; }
    .sheet-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 5px 0;
        font-family: 'Alegreya', serif;
        color: #cbb295;
        font-size: 0.9rem;
    }
    .sheet-row strong { color: #e8e0d4; font-family: 'Cinzel', serif; }
    .sheet-row em { color: #7a6f5f; font-style: normal; font-size: 0.8rem; }

    .ob-actions {
        display: flex;
        justify-content: flex-end;
        gap: 10px;
        margin-top: 20px;
    }
    .ob-btn {
        padding: 10px 18px;
        border: 2px solid #7c6243;
        border-radius: 8px;
        background: transparent;
        color: #d4af37;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.95rem;
        cursor: pointer;
        transition: all 0.2s ease;
        box-shadow: 0 4px 0 #4a3b2c;
    }
    .ob-btn:hover:not(:disabled) { transform: translateY(-1px); }
    .ob-btn:disabled { opacity: 0.4; cursor: not-allowed; }
    .ob-btn.ghost { border-color: #3a3028; color: #a09080; box-shadow: none; }
    .ob-btn.cta {
        background: linear-gradient(135deg, #c5a059, #a67c37);
        color: #14100a;
        border-color: #c5a059;
        box-shadow: 0 4px 0 #6e5122;
    }

    @media (max-width: 520px) {
        .class-grid { grid-template-columns: repeat(2, 1fr); }
        .ob-steps li { font-size: 0.65rem; }
    }
</style>