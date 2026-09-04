<script>
    export let value = '';
    export let placeholder = '';
</script>

<input 
    type="text" 
    {placeholder} 
    bind:value 
    class="w-full p-3 bg-[#2b221a] border-2 border-stone-600 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-dnd-gold focus:border-transparent transition-all font-alegreya placeholder:text-stone-500 shadow-inner"
/>

