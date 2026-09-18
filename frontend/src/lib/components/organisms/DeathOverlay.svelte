<script>
    import { deathOverlay } from '../../stores/game';
</script>

{#if $deathOverlay.active}
    <div class="death-overlay" role="alert" aria-live="assertive">
        <div class="death-card">
            <div class="death-icon">☠️</div>
            <h2 class="death-title">Vous êtes tombé</h2>
            <p class="death-text">Au combat, vous avez succombé...</p>
            <p class="death-text">La Taverne vous tend les bras, aventurier.</p>
        </div>
    </div>
{/if}

<style>
    .death-overlay {
        position: fixed;
        inset: 0;
        z-index: 300;
        display: flex;
        align-items: center;
        justify-content: center;
        background: radial-gradient(ellipse at center, rgba(60, 5, 5, 0.55), rgba(0, 0, 0, 0.85));
        pointer-events: none;
        animation: death-fade 1s ease-in forwards;
    }

    .death-card {
        text-align: center;
        padding: 40px 48px;
        background: rgba(10, 10, 10, 0.9);
        border: 1px solid rgba(239, 68, 68, 0.5);
        border-radius: 16px;
        box-shadow: 0 0 60px rgba(239, 68, 68, 0.25), inset 0 0 40px rgba(120, 10, 10, 0.2);
        animation: death-card-pop 1.2s ease-out;
    }

    .death-icon {
        font-size: 4rem;
        margin-bottom: 12px;
        animation: death-pulse 2s ease-in-out infinite;
    }

    .death-title {
        font-family: 'Cinzel', serif;
        color: #fca5a5;
        font-size: 1.8rem;
        margin: 0 0 8px 0;
        letter-spacing: 0.06em;
    }

    .death-text {
        font-family: 'Alegreya', serif;
        color: #a09080;
        font-size: 0.95rem;
        font-style: italic;
        margin: 4px 0;
    }

    @keyframes death-fade {
        from { opacity: 0; }
        to { opacity: 1; }
    }

    @keyframes death-card-pop {
        0% { transform: scale(0.6); opacity: 0; }
        60% { transform: scale(1.05); opacity: 1; }
        100% { transform: scale(1); }
    }

    @keyframes death-pulse {
        0%, 100% { transform: scale(1); }
        50% { transform: scale(1.15); }
    }
</style>