<script>
    import { onDestroy } from 'svelte';
    import { gameEvents } from '../../stores/game';

    export let targetKey = '';

    let current = null;
    let lastId = 0;
    let timer;

    function handleEvents(events) {
        for (const ev of events) {
            if (ev.id <= lastId) continue;
            if (ev.target === targetKey && (ev.event === 'damage' || ev.event === 'heal' || ev.event === 'death')) {
                lastId = ev.id;
                current = ev;
                clearTimeout(timer);
                timer = setTimeout(() => { current = null; }, 1400);
            }
        }
    }

    const unsub = gameEvents.subscribe(handleEvents);
    onDestroy(() => { unsub(); clearTimeout(timer); });
</script>

{#if current}
    {#key current.id}
        <div class="float-dmg {current.event} {current.crit ? 'crit' : ''}" aria-hidden="true">
            {#if current.event === 'death'}
                ☠️
            {:else}
                {current.event === 'heal' ? '+' : '−'}{Math.round(current.amount)}
            {/if}
        </div>
    {/key}
{/if}

<style>
    .float-dmg {
        position: absolute;
        top: 12px;
        right: 14px;
        z-index: 5;
        font-family: 'Cinzel', serif;
        font-size: 1.5rem;
        font-weight: 700;
        line-height: 1;
        color: #f87171;
        text-shadow: 0 2px 6px rgba(0, 0, 0, 0.8), 0 0 12px rgba(239, 68, 68, 0.6);
        pointer-events: none;
        animation: float-up 1.4s ease-out forwards;
        white-space: nowrap;
    }

    .float-dmg.heal {
        color: #4ade80;
        text-shadow: 0 2px 6px rgba(0, 0, 0, 0.8), 0 0 12px rgba(34, 197, 94, 0.6);
    }

    .float-dmg.crit {
        font-size: 2rem;
        color: #fbbf24;
        text-shadow: 0 2px 6px rgba(0, 0, 0, 0.85), 0 0 18px rgba(245, 158, 11, 0.7);
    }

    .float-dmg.death {
        font-size: 2rem;
    }

    @keyframes float-up {
        0% { opacity: 0; transform: translateY(6px) scale(0.7); }
        15% { opacity: 1; transform: translateY(0) scale(1.15); }
        30% { transform: translateY(-4px) scale(1); }
        100% { opacity: 0; transform: translateY(-34px) scale(0.95); }
    }
</style>