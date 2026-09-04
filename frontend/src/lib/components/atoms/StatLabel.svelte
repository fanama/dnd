<script>
    export let label = "Stat";
    export let value = 0;
    export let icon = '';
</script>

<div class="stat-label">
    <span class="stat-text">
        {#if icon}<span class="stat-icon">{icon}</span>{/if}
        {label}
    </span>
    <span class="stat-value">{value}</span>
</div>

<style>
    .stat-label {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 6px 10px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.06);
        background: rgba(255, 255, 255, 0.03);
        border-radius: 4px;
        transition: background 0.15s ease;
    }

    .stat-label:hover {
        background: rgba(255, 255, 255, 0.07);
    }

    .stat-text {
        font-family: 'Alegreya', serif;
        color: #c0b8a8;
        font-size: 0.85rem;
        display: flex;
        align-items: center;
        gap: 6px;
    }

    .stat-icon {
        font-size: 0.9rem;
    }

    .stat-value {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.9rem;
        font-weight: bold;
    }
</style>
