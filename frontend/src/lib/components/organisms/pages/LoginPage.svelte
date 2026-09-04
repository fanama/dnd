<script>
    import Button from '../../atoms/Button.svelte';
    import Input from '../../atoms/Input.svelte';

    export let pseudo = '';
    export let charName = '';
    export let charClass = 'Guerrier';
    export let onJoin = () => {};

    let pseudoError = '';
    let charNameError = '';
    let submitted = false;

    $: isValid = pseudo.trim().length >= 2 && charName.trim().length >= 2;
    $: canSubmit = isValid && !submitted;

    function validate() {
        submitted = true;
        pseudoError = '';
        charNameError = '';

        if (pseudo.trim().length < 2) {
            pseudoError = 'Minimum 2 caracteres';
        }
        if (charName.trim().length < 2) {
            charNameError = 'Minimum 2 caracteres';
        }

        if (pseudoError || charNameError) return;

        onJoin();
    }

    function handleKeydown(e) {
        if (e.key === 'Enter' && canSubmit) {
            validate();
        }
    }

    const classes = [
        { value: 'Guerrier', label: 'Guerrier', icon: '🛡️', desc: 'Force et endurance' },
        { value: 'Magicien', label: 'Magicien', icon: '🔮', desc: 'Arcanes et mystere' },
        { value: 'Voleur', label: 'Voleur', icon: '🗡️', desc: 'Agilite et ruse' },
        { value: 'Clerc', label: 'Clerc', icon: '✝️', desc: 'Foi et guerison' },
        { value: 'Barde', label: 'Barde', icon: '🎵', desc: 'Charmes et musique' },
        { value: 'Ranger', label: 'Ranger', icon: '🏹', desc: 'Nature et precision' }
    ];
</script>

<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="login-container" on:keydown={handleKeydown}>
    <!-- Decorative top flourish -->
    <div class="flourish" aria-hidden="true">
        <div class="flourish-line"></div>
        <span class="flourish-icon">⚔️</span>
        <div class="flourish-line"></div>
    </div>

    <div class="login-card fade-in">
        <!-- Header -->
        <div class="login-header">
            <h2 class="login-title">Enregistrement du Registre</h2>
            <p class="login-subtitle">Inscrivez votre nom dans la legende avant de braver le donjon</p>
        </div>

        <!-- Form Fields -->
        <div class="login-fields">
            <Input
                id="pseudo-input"
                label="Pseudo Joueur"
                bind:value={pseudo}
                placeholder="Votre nom d'aventurier..."
                error={pseudoError}
            />

            <Input
                id="char-name-input"
                label="Nom du Personnage"
                bind:value={charName}
                placeholder="Nom de votre personnage..."
                error={charNameError}
            />

            <div class="field-group">
                <label for="char-class-select" class="field-label">Classe de Depart</label>
                <div class="class-grid">
                    {#each classes as cls}
                        <button
                            type="button"
                            class="class-card"
                            class:selected={charClass === cls.value}
                            on:click={() => charClass = cls.value}
                        >
                            <span class="class-icon">{cls.icon}</span>
                            <span class="class-name">{cls.label}</span>
                            <span class="class-desc">{cls.desc}</span>
                        </button>
                    {/each}
                </div>
            </div>
        </div>

        <!-- Action -->
        <div class="login-actions">
            <Button
                onClick={validate}
                disabled={!canSubmit}
                loading={submitted}
                className="w-full py-4 text-xl"
            >
                Forger le Destin
            </Button>
            {#if !isValid && submitted}
                <p class="hint-error">Veuillez remplir tous les champs (2 caracteres minimum)</p>
            {/if}
        </div>
    </div>
</div>

<style>
    .login-container {
        max-width: 480px;
        margin: 0 auto;
        padding: 40px 16px;
    }

    .flourish {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 16px;
        margin-bottom: 24px;
    }

    .flourish-line {
        flex: 1;
        max-width: 80px;
        height: 1px;
        background: linear-gradient(to right, transparent, #c5a059, transparent);
    }

    .flourish-icon {
        font-size: 1.5rem;
        opacity: 0.7;
    }

    .login-card {
        background: linear-gradient(135deg, #0d0d0d 0%, #151210 100%);
        border: 1px solid rgba(197, 160, 89, 0.25);
        border-radius: 16px;
        padding: 32px 28px;
        box-shadow: 0 20px 60px rgba(0, 0, 0, 0.6),
                    0 0 40px rgba(197, 160, 89, 0.05);
    }

    .login-header {
        text-align: center;
        margin-bottom: 28px;
        padding-bottom: 20px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.15);
    }

    .login-title {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 1.6rem;
        margin: 0 0 8px 0;
        letter-spacing: 0.02em;
    }

    .login-subtitle {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.9rem;
        font-style: italic;
        margin: 0;
    }

    .login-fields {
        display: flex;
        flex-direction: column;
        gap: 20px;
    }

    .field-group {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .field-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        color: #c5a059;
        text-transform: uppercase;
        letter-spacing: 0.08em;
    }

    .class-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 10px;
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
        box-shadow: 0 0 20px rgba(197, 160, 89, 0.15),
                    inset 0 0 20px rgba(197, 160, 89, 0.05);
    }

    .class-icon {
        font-size: 1.6rem;
        line-height: 1;
    }

    .class-name {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.85rem;
        font-weight: bold;
    }

    .class-desc {
        font-family: 'Alegreya', serif;
        color: #7a6f5f;
        font-size: 0.65rem;
        font-style: italic;
    }

    .class-card.selected .class-desc {
        color: #a09080;
    }

    .login-actions {
        margin-top: 24px;
        display: flex;
        flex-direction: column;
        gap: 8px;
    }

    .hint-error {
        font-family: 'Alegreya', serif;
        font-size: 0.85rem;
        color: #ef4444;
        text-align: center;
        margin: 0;
    }

    @media (max-width: 480px) {
        .login-card {
            padding: 24px 20px;
        }
        .login-title {
            font-size: 1.3rem;
        }
    }
</style>
