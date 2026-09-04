<script>
    import { onMount } from 'svelte';

    export let current = 100;
    export let max = 100;

    let displayPercentage = 0;
    let trailPercentage = 0;
    let showFlash = false;
    let flashTimeout;
    let trailTimeout;

    $: safeMax = max > 0 ? max : 100;
    $: numericCurrent = Math.min(Math.max(Number(current), 0), safeMax);
    $: percentage = Math.min(Math.max((numericCurrent / safeMax) * 100, 0), 100);

    $: isCritical = percentage <= 25 && percentage > 0;

    $: {
        const newPct = percentage;
        const oldPct = displayPercentage;

        if (newPct < oldPct) {
            // Damage taken — snap green bar down, keep trail at old value
            clearTimeout(flashTimeout);
            clearTimeout(trailTimeout);

            showFlash = true;
            trailPercentage = oldPct;

            // Flash for 300ms
            flashTimeout = setTimeout(() => { showFlash = false; }, 300);

            // After a short delay, start trail catching up
            trailTimeout = setTimeout(() => {
                trailPercentage = newPct;
            }, 600);

            displayPercentage = newPct;
        } else {
            // Healing or same — no trail needed
            displayPercentage = newPct;
            trailPercentage = newPct;
        }
    }

    onMount(() => {
        displayPercentage = percentage;
        trailPercentage = percentage;
    });
</script>

<div class="hp-bar-root select-none">
    <div class="hp-frame">
        <div class="hp-cap hp-cap-left"></div>

        <div class="hp-track">
            <!-- Dark background -->
            <div class="absolute inset-0 bg-stone-950 rounded-sm"></div>

            <!-- Damage trail (red) — stays at old value, catches up slowly -->
            <div
                class="absolute inset-y-0 left-0 trail-bar rounded-sm"
                style="width: {trailPercentage}%;"
            ></div>

            <!-- Green HP fill — snaps immediately -->
            <div
                class="absolute inset-y-0 left-0 fill-bar rounded-sm"
                style="width: {displayPercentage}%;"
            >
                <div class="absolute inset-x-0 top-0 h-[45%] bg-gradient-to-b from-white/40 to-transparent rounded-t-sm"></div>
                <div class="absolute inset-x-0 bottom-0 h-[35%] bg-black/30"></div>
                <div class="absolute inset-x-2 top-[20%] h-[2px] bg-white/20 rounded-full"></div>
            </div>

            <!-- Red flash overlay on damage -->
            {#if showFlash}
                <div class="absolute inset-0 damage-flash rounded-sm"></div>
            {/if}

            <!-- Segment grid -->
            <div class="absolute inset-0 flex justify-between pointer-events-none opacity-20">
                {#each Array(9) as _}
                    <div class="h-full w-[1px] bg-black/80 shadow-[1px_0_0_rgba(255,255,255,0.08)]"></div>
                {/each}
            </div>

            <!-- Critical pulsing overlay -->
            {#if isCritical}
                <div class="absolute inset-0 bg-red-600/20 pointer-events-none animate-ping opacity-80 rounded-sm"></div>
            {/if}

            <!-- HP text -->
            <div class="absolute inset-0 flex items-center justify-between px-3 z-10">
                <div class="hp-badge-left">
                    <span class="hp-current">{numericCurrent}</span>
                    <span class="hp-label">PV</span>
                </div>
                <div class="hp-badge-right">
                    <span class="hp-max">{safeMax}</span>
                    <span class="hp-label">MAX</span>
                </div>
            </div>
        </div>

        <div class="hp-cap hp-cap-right"></div>
    </div>

    <div class="hp-percent">
        <span class:critical={isCritical}>{Math.round(displayPercentage)}%</span>
    </div>
</div>

<style>
    .hp-bar-root {
        display: flex;
        flex-direction: column;
        gap: 4px;
        width: 100%;
    }

    .hp-frame {
        display: flex;
        align-items: stretch;
        height: 40px;
        filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.5));
    }

    .hp-cap {
        width: 14px;
        background: linear-gradient(to bottom, #6b4f35, #3d2b1a, #2a1c12);
        border: 2px solid #8b6544;
        flex-shrink: 0;
    }

    .hp-cap-left {
        border-radius: 6px 0 0 6px;
        border-right: 1px solid #5c4033;
    }

    .hp-cap-right {
        border-radius: 0 6px 6px 0;
        border-left: 1px solid #5c4033;
    }

    .hp-track {
        flex: 1;
        position: relative;
        border-top: 2px solid #8b6544;
        border-bottom: 2px solid #8b6544;
        overflow: hidden;
        background: #0c0a08;
        box-shadow: inset 0 3px 8px rgba(0, 0, 0, 0.9);
    }

    /* Green fill — instant snap */
    .fill-bar {
        background: linear-gradient(to right, #047857, #10b981, #34d399);
        box-shadow: 0 0 16px rgba(34, 197, 94, 0.7);
        transition: width 0.15s ease-out;
    }

    /* Red damage trail — slow catch-up */
    .trail-bar {
        background: linear-gradient(to right, #7f1d1d, #b91c1c, #dc2626);
        opacity: 0.9;
        transition: width 1.2s cubic-bezier(0.4, 0, 0.2, 1);
    }

    /* Red flash on hit */
    .damage-flash {
        background: radial-gradient(ellipse at left, rgba(239, 68, 68, 0.5), transparent 70%);
        animation: flashIn 0.3s ease-out forwards;
    }

    @keyframes flashIn {
        0% { opacity: 1; }
        100% { opacity: 0; }
    }

    .hp-badge-left,
    .hp-badge-right {
        display: flex;
        align-items: baseline;
        gap: 4px;
        padding: 2px 10px;
        background: rgba(0, 0, 0, 0.6);
        backdrop-filter: blur(4px);
        border-radius: 4px;
        border: 1px solid rgba(255, 255, 255, 0.12);
        font-family: 'MedievalSharp', cursive;
        font-weight: bold;
        text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
        white-space: nowrap;
    }

    .hp-current {
        font-size: 1.1rem;
        color: #4ade80;
    }

    .hp-max {
        font-size: 0.85rem;
        color: #a8a29e;
    }

    .hp-label {
        font-size: 0.6rem;
        color: #78716c;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }

    .hp-percent {
        text-align: right;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.7rem;
        color: #78716c;
        padding-right: 2px;
    }

    .hp-percent .critical {
        color: #ef4444;
        animation: pulse 1.5s ease-in-out infinite;
    }

    @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
    }
</style>
