<script>
    import Button from '../atoms/Button.svelte';
    export let item = '';
    export let index = 0;
    export let onLoot = (idx) => {};
</script>

<div class="loot-item">
    <span>💎 {item.nom}</span>
    <Button variant="success" onClick={() => onLoot(index)}>Ramasser</Button>
</div>

<style>
    .loot-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: #3a2f28;
        padding: 8px 12px;
        border-radius: 4px;
        margin-bottom: 8px;
        border: 1px solid #d4af37;
    }
</style>
