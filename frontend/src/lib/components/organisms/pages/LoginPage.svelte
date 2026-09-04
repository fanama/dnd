<script>
    import Button from '../../atoms/Button.svelte';
    import Input from '../../atoms/Input.svelte';
    export let pseudo = '';
    export let charName = '';
    export let charClass = 'Guerrier';
    export let onJoin = () => {};
</script>

<div id="auth" class="max-w-lg mx-auto bg-dnd-panel border-2 border-dnd-accent rounded-2xl p-8 sm:p-10 my-10 flex flex-col gap-8 shadow-2xl transform transition-all hover:scale-[1.01] text-white">
    <!-- Header Section -->
    <div class="text-center space-y-2 pb-2 border-b border-stone-700/60">
        <h3 class="text-dnd-gold font-medieval text-3xl tracking-wide">Enregistrement du Registre</h3>
        <p class="text-stone-400 text-sm font-alegreya italic">Inscrivez votre nom dans la légende avant de braver le donjon</p>
    </div>

    <!-- Form Fields -->
    <div class="space-y-6">
        <div class="space-y-2">
            <label for="pseudo-input" class="block text-xs uppercase tracking-wider text-dnd-gold font-medieval">Pseudo Joueur</label>
            <Input id="pseudo-input" bind:value={pseudo} placeholder="Ton pseudo..." class="w-full" />
        </div>

        <div class="space-y-2">
            <label for="char-name-input" class="block text-xs uppercase tracking-wider text-dnd-gold font-medieval">Nom du Personnage</label>
            <Input id="char-name-input" bind:value={charName} placeholder="Nom de ton personnage..." class="w-full" />
        </div>

        <div class="space-y-2">
            <label for="char-class-select" class="block text-xs uppercase tracking-wider text-dnd-gold font-medieval">Classe de Départ</label>
            <select id="char-class-select" bind:value={charClass} class="w-full p-3.5 bg-[#2b221a] border border-stone-600 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-dnd-gold transition-all font-alegreya text-base">
                <option value="Guerrier">🛡️ Guerrier</option>
                <option value="Magicien">🔮 Magicien</option>
            </select>
        </div>
    </div>

    <!-- Action Section -->
    <div class="pt-2">
        <Button onClick={onJoin} class="w-full py-4 text-xl font-bold shadow-xl hover:brightness-110 transition-all tracking-wider">
            Forger le Destin
        </Button>
    </div>
</div>
