<script>
    // Props
    export let current = 0;
    export let max = 100;
    export let showNumbers = true;

    // Calcul du pourcentage sécurisé
    $: safeMax = max > 0 ? max : 100;
    $: numericCurrent = Number.isFinite(Number(current)) ? Math.max(0, Number(current)) : 0;
    $: percent = Math.round(Math.max(0, Math.min(100, (numericCurrent / safeMax) * 100)));

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

<!-- Conteneur principal avec hauteur standardisée (h-6 = 24px) -->
<div class="w-full bg-stone-900 h-[2rem] rounded-lg border border-amber-900/60 overflow-hidden relative shadow-inner">
    <div
        class="h-full transition-all duration-500 ease-out"
        style={fillStyle}
    ></div>
</div>

{#if showNumbers}
    <div class="text-xs text-stone-300 font-bold text-right mt-1 tracking-wide">
        {percent}%
    </div>
{/if}
