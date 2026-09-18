<script>
    import { toasts, dismissToast } from '../stores/toasts';

    const KIND_META = {
        success: { icon: '✅', label: 'Succès' },
        error: { icon: '❌', label: 'Erreur' },
        info: { icon: '📜', label: 'Information' },
        warn: { icon: '⚠️', label: 'Attention' },
        damage: { icon: '💥', label: 'Dégâts' },
        heal: { icon: '❤️', label: 'Soin' },
        crit: { icon: '💥', label: 'CRITIQUE' },
    };
</script>

<div class="toast-stack" aria-live="polite">
    {#each $toasts as toast (toast.id)}
        <div
            class="toast toast-{toast.kind} fade-slide"
            role="status"
        >
            <span class="toast-icon">{KIND_META[toast.kind]?.icon || '📜'}</span>
            <div class="toast-body">
                <span class="toast-title">{toast.title}</span>
                {#if toast.msg}
                    <span class="toast-msg">{toast.msg}</span>
                {/if}
            </div>
            <button class="toast-close" on:click|stopPropagation={() => dismissToast(toast.id)} aria-label="Fermer">✕</button>
        </div>
    {/each}
</div>

<style>
    .toast-stack {
        position: fixed;
        top: 16px;
        right: 16px;
        z-index: 200;
        display: flex;
        flex-direction: column;
        gap: 10px;
        max-width: min(360px, calc(100vw - 32px));
        pointer-events: none;
    }

    .toast {
        pointer-events: auto;
        display: flex;
        align-items: flex-start;
        gap: 12px;
        padding: 12px 14px;
        background: rgba(10, 10, 10, 0.92);
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-left-width: 4px;
        border-radius: 10px;
        box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(8px);
        transition: box-shadow 0.2s ease;
    }

    .toast:hover {
        box-shadow: 0 12px 34px rgba(0, 0, 0, 0.75);
    }

    .toast-success { border-left-color: #22c55e; }
    .toast-error { border-left-color: #ef4444; }
    .toast-info { border-left-color: #3b82f6; }
    .toast-warn { border-left-color: #f59e0b; }
    .toast-damage { border-left-color: #ef4444; }
    .toast-heal { border-left-color: #22c55e; }
    .toast-crit { border-left-color: #f59e0b; }

    .toast-icon {
        font-size: 1.3rem;
        line-height: 1.2;
        flex-shrink: 0;
    }

    .toast-body {
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-width: 0;
        flex: 1;
    }

    .toast-title {
        font-family: 'Cinzel', serif;
        color: #e8e0d4;
        font-size: 0.85rem;
        font-weight: 700;
        word-break: break-word;
    }

    .toast-damage .toast-title { color: #fca5a5; }
    .toast-heal .toast-title { color: #86efac; }
    .toast-crit .toast-title { color: #fbbf24; }

    .toast-msg {
        font-family: 'Alegreya', serif;
        color: #a09080;
        font-size: 0.78rem;
        line-height: 1.4;
        word-break: break-word;
    }

    .toast-close {
        background: transparent;
        border: none;
        color: #5a5045;
        font-size: 0.8rem;
        cursor: pointer;
        padding: 2px 4px;
        flex-shrink: 0;
        line-height: 1;
    }

    .toast-close:hover { color: #e8e0d4; }

    .fade-slide {
        animation: toast-in 0.35s ease-out;
    }

    @keyframes toast-in {
        from { opacity: 0; transform: translateX(24px); }
        to { opacity: 1; transform: translateX(0); }
    }

    @media (max-width: 480px) {
        .toast-stack {
            top: 12px;
            right: 12px;
            left: 12px;
            max-width: none;
        }
    }
</style>