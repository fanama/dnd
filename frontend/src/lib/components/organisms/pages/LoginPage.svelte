<script>
    import Button from '../../atoms/Button.svelte';
    import Input from '../../atoms/Input.svelte';

    export let pseudo = '';
    export let onJoin = () => {};

    let pseudoError = '';
    let submitted = false;

    $: isValid = pseudo.trim().length >= 2;
    $: canSubmit = isValid && !submitted;

    function validate() {
        submitted = true;
        pseudoError = '';

        if (pseudo.trim().length < 2) {
            pseudoError = 'Minimum 2 caracteres';
        }

        if (pseudoError) return;

        onJoin();
    }

    function handleKeydown(e) {
        if (e.key === 'Enter' && canSubmit) {
            validate();
        }
    }
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
            <h2 class="login-title">Le Registre des Aventuriers</h2>
            <p class="login-subtitle">Inscrivez votre nom dans la legende... la forge de votre heros vous attend</p>
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
        </div>

        <!-- Action -->
        <div class="login-actions">
            <Button
                onClick={validate}
                disabled={!canSubmit}
                loading={submitted}
                className="w-full py-4 text-xl"
            >
                Se présenter
            </Button>
            {#if !isValid && submitted}
                <p class="hint-error">Veuillez indiquer un pseudo (2 caracteres minimum)</p>
            {/if}
        </div>
    </div>
</div>

<style>
    .login-container {
        max-width: 440px;
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