<script>
    export let current = 100;
    export let max = 100;

    $: safeMax = max > 0 ? max : 100;
    $: numericCurrent = Math.min(Math.max(Number(current), 0), safeMax);
    $: percentage = Math.min(Math.max((numericCurrent / safeMax) * 100, 0), 100);

    // Dynamic bar styling based on remaining health percentage
    $: colorStyle = percentage > 50 
        ? 'from-emerald-700 via-emerald-500 to-green-400 shadow-[0_0_12px_rgba(34,197,94,0.6)]' 
        : percentage > 25 
            ? 'from-amber-700 via-amber-500 to-yellow-400 shadow-[0_0_12px_rgba(245,158,11,0.6)]' 
            : 'from-rose-800 via-red-600 to-red-400 shadow-[0_0_14px_rgba(239,68,68,0.8)] animate-pulse';

    $: isCritical = percentage <= 25 && percentage > 0;
</script>

<div class="flex flex-col gap-1 w-full select-none">
    <!-- Outer Decorative Wooden/Metallic Frame -->
    <div class="relative w-full h-8 p-1 bg-gradient-to-b from-[#4a3525] via-[#2d1e13] to-[#1a100a] border-2 border-[#8b6544] rounded-lg shadow-xl ring-1 ring-black/80">
        
        <!-- Track Container with Inner Shadow Depth -->
        <div class="relative w-full h-full bg-stone-950 rounded overflow-hidden shadow-[inset_0_3px_6px_rgba(0,0,0,0.9)] border border-black/50">
            
            <!-- Delayed Bleed / Damage Under-Bar -->
            <div 
                class="absolute inset-y-0 left-0 bg-red-950/80 transition-all duration-700 ease-out"
                style="width: {percentage}%;"
            ></div>

            <!-- Active HP Fill Bar -->
            <div 
                class="h-full transition-all duration-500 ease-out relative bg-gradient-to-r {colorStyle}"
                style="width: {percentage}%;"
            >
                <!-- Top Gloss / Bevel Highlight -->
                <div class="absolute inset-x-0 top-0 h-[40%] bg-gradient-to-b from-white/35 to-transparent"></div>
                
                <!-- Bottom Ambient Shadow -->
                <div class="absolute inset-x-0 bottom-0 h-[30%] bg-black/25"></div>
            </div>

            <!-- 10% Segment Grid Divider Overlay -->
            <div class="absolute inset-0 flex justify-between pointer-events-none opacity-25">
                {#each Array(9) as _}
                    <div class="h-full w-[1px] bg-black shadow-[1px_0_0_rgba(255,255,255,0.1)]"></div>
                {/each}
            </div>

            <!-- Low Health Warning Flash Overlay -->
            {#if isCritical}
                <div class="absolute inset-0 bg-red-600/15 pointer-events-none animate-ping opacity-75"></div>
            {/if}

            <!-- Text Overlay with High-Contrast Dropshadow -->
            <div class="absolute inset-0 flex items-center justify-center font-medieval text-sm font-bold tracking-wider text-white uppercase drop-shadow-[0_2px_2px_rgba(0,0,0,1)] z-10">
                <span class="px-3 py-1 rounded bg-black/50 backdrop-blur-sm border border-white/20">
                    <span class="text-emerald-400">{numericCurrent}</span> <span class="text-stone-400 mx-1">/</span> <span class="text-stone-300">{safeMax}</span> <span class="ml-1 text-[10px] opacity-80">PV</span>
                </span>
            </div>
        </div>
    </div>
</div>
