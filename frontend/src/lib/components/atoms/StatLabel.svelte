<script>
    export let label = "Stat";
    export let value = 0;
</script>

<div class="flex justify-between py-1 px-2 border-b border-white/10 bg-white/5 rounded-sm transition-colors hover:bg-white/10 group">
    <span class="text-white italic text-sm group-hover:text-white transition-colors">{label}</span>
    <span class="text-dnd-gold font-bold text-sm group-hover:scale-110 transition-transform">{value}</span>
</div>
