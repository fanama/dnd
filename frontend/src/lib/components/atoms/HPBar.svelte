<script>
    import { onDestroy } from 'svelte';
    import { gameEvents } from '../../stores/game';

    // Props
    export let current = 0;
    export let max = 100;
    export let showNumbers = true;
    // who matches the event.target to flash this bar when the entity is hit
    export let who = '';

    let flash = false;
    let flashTimer;

    function onEvents(events) {
        if (!who) return;
        for (const ev of events) {
            if (ev.target === who && ev.event === 'damage') {
                flash = true;
                clearTimeout(flashTimer);
                flashTimer = setTimeout(() => { flash = false; }, 450);
            }
        }
    }
    const unsub = gameEvents.subscribe(onEvents);
    onDestroy(() => { unsub(); clearTimeout(flashTimer); });

    // Calcul du pourcentage sécurisé
    $: safeMax = max > 0 ? max : 100;
    $: numericCurrent = Number.isFinite(Number(current)) ? Math.max(0, Number(current)) : 0;
    $: displayCurrent = Math.min(numericCurrent, safeMax);
    $: percent = Math.round(Math.max(0, Math.min(100, (displayCurrent / safeMax) * 100)));

    // Styles dynamiques selon le pourcentage
    $: fill = percent > 50
        ? 'linear-gradient(180deg, #4ade80 0%, #22c55e 55%, #16a34a 100%)'
        : percent > 20
            ? 'linear-gradient(180deg, #fde047 0%, #facc15 55%, #d97706 100%)'
            : 'linear-gradient(180deg, #f87171 0%, #ef4444 55%, #b91c1c 100%)';

    $: glow = percent > 50
        ? 'rgba(34, 197, 94, 0.85)'
        : percent > 20
            ? 'rgba(250, 204, 21, 0.85)'
            : 'rgba(248, 113, 113, 0.85)';

    $: fillStyle = `width: ${percent}%; background: ${fill}; box-shadow: 0 0 10px ${glow}, inset 0 1px 0 rgba(255, 255, 255, 0.35);`;
</script>

<div
    class="w-full bg-stone-900 h-[2rem] rounded-lg border border-amber-900/60 overflow-hidden relative shadow-inner"
    class:hp-flash={flash}
>
    <div
        class="h-full transition-all duration-500 ease-out"
        style={fillStyle}
    ></div>
    {#if showNumbers}
        <span class="hp-numbers">{Math.ceil(displayCurrent)} / {Math.round(safeMax)}</span>
    {/if}
</div>

<style>
    .hp-numbers {
        position: absolute;
        inset: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        font-family: 'Cinzel', serif;
        font-size: 0.75rem;
        font-weight: 700;
        color: #f5efe6;
        text-shadow: 0 1px 3px rgba(0, 0, 0, 0.9);
        pointer-events: none;
    }

    .hp-flash {
        animation: hp-flash-anim 0.45s ease-out;
    }

    @keyframes hp-flash-anim {
        0% { box-shadow: 0 0 0 rgba(239, 68, 68, 0); border-color: rgba(239, 68, 68, 0); }
        30% { box-shadow: 0 0 24px rgba(239, 68, 68, 0.9), inset 0 0 16px rgba(239, 68, 68, 0.6); border-color: #ef4444; }
        100% { box-shadow: 0 0 0 rgba(239, 68, 68, 0); border-color: rgba(239, 68, 68, 0); }
    }
</style>