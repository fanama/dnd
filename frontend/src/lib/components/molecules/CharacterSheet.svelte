<script>
    import HPBar from '../atoms/HPBar.svelte';
    import StatLabel from '../atoms/StatLabel.svelte';
    
    export let stats = {};
</script>

<div class="bg-[#f4e4bc] border-[12px] border-double border-[#5c4033] rounded-sm p-6 text-black shadow-2xl relative overflow-hidden group transition-all">
    <div class="absolute top-0 right-0 opacity-10 pointer-events-none rotate-12 translate-x-4 -translate-y-4">
        <span class="text-9xl font-serif">📜</span>
    </div>

    <div class="relative z-10">
        <div class="flex flex-col md:flex-row gap-6 mb-8 border-b-2 border-[#5c4033]/30 pb-6">
            <div class="flex-1 text-center md:text-left">
                <label class="block text-[0.6rem] uppercase font-bold text-black font-medieval">Nom du Personnage</label>
                <h2 class="font-cinzel text-3xl font-bold m-0 tracking-wider text-black">{stats.nom || 'Inconnu'}</h2>
            </div>
            <div class="flex-1 text-center md:text-right">
                <label class="block text-[0.6rem] uppercase font-bold text-black font-medieval">Classe & Origine</label>
                <span class="font-medieval text-xl italic text-black">{stats.classe || 'Sans Classe'}</span>
            </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div class="md:col-span-2 grid grid-cols-2 sm:grid-cols-3 gap-4">
                {#each [
                    {label: 'Force', val: stats.stats?.force},
                    {label: 'Constitution', val: stats.stats?.constitution},
                    {label: 'Vitesse', val: stats.stats?.vitesse},
                    {label: 'Savoir', val: stats.stats?.savoir},
                    {label: 'Instinct', val: stats.stats?.instinct},
                    {label: 'Charisme', val: stats.stats?.charisme}
                ] as stat}
                    <div class="flex flex-col items-center p-2 border-2 border-black/20 rounded-lg bg-[#e8d4a9] shadow-inner">
                        <span class="text-[0.7rem] uppercase font-bold text-black font-medieval mb-1">{stat.label}</span>
                        <span class="text-2xl font-bold font-cinzel text-black">{stat.val || 0}</span>
                    </div>
                {/each}
            </div>

            <div class="flex flex-col gap-4 bg-[#dcc495] p-4 rounded-lg border-2 border-black/20 shadow-inner">
                <div class="text-center">
                    <label class="block text-[0.6rem] uppercase font-bold text-black font-medieval mb-2">Lieu Actuel</label>
                    <span class="font-medieval text-lg font-bold text-black">{stats.lieu}</span>
                </div>
                <div class="pt-4 border-t border-black/20">
                    <div class="flex justify-between items-center text-[0.7rem] font-bold text-black uppercase mb-1">
                        <span>Points de Vie</span>
                        <span>{stats.pv} / 100</span>
                    </div>
                    <HPBar current={stats.pv} max={100} />
                </div>
            </div>
        </div>
    </div>
</div>
