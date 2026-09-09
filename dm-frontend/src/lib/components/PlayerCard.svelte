<script>
    export let player = '';
    export let data = {};
    export let selected = false;
    export let onSelect = (pseudo) => {};
</script>

<button class="player-card" class:selected on:click={() => onSelect(player)}>
    <div class="card-top">
        <span class="card-name">{data.nom || player}</span>
        <span class="card-class">{data.classe}</span>
    </div>
    <div class="card-stats">
        <span>❤️ {data.pv}/{data.max_pv}</span>
        <span>📍 {data.lieu}</span>
    </div>
    <div class="card-bar">
        <div class="card-bar-fill" style="width: {data.max_pv ? (data.pv / data.max_pv * 100) : 100}%"></div>
    </div>
</button>

<style>
    .player-card {
        width: 100%;
        padding: 14px;
        background: rgba(26, 20, 16, 0.6);
        border: 1px solid rgba(197, 160, 89, 0.1);
        border-radius: 10px;
        cursor: pointer;
        text-align: left;
        transition: all 0.2s ease;
        display: flex;
        flex-direction: column;
        gap: 8px;
        color: inherit;
    }

    .player-card:hover {
        border-color: rgba(197, 160, 89, 0.3);
        background: rgba(26, 20, 16, 0.8);
    }

    .player-card.selected {
        border-color: #c5a059;
        background: rgba(197, 160, 89, 0.1);
    }

    .card-top {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .card-name {
        font-family: 'Cinzel', serif;
        color: #c5a059;
        font-size: 0.9rem;
        font-weight: 700;
    }

    .card-class {
        font-family: 'MedievalSharp', cursive;
        color: #7a6f5f;
        font-size: 0.7rem;
    }

    .card-stats {
        display: flex;
        justify-content: space-between;
        font-family: 'Alegreya', serif;
        font-size: 0.75rem;
        color: #7a6f5f;
    }

    .card-bar {
        height: 4px;
        background: rgba(0, 0, 0, 0.4);
        border-radius: 4px;
        overflow: hidden;
    }

    .card-bar-fill {
        height: 100%;
        background: linear-gradient(90deg, #22c55e, #16a34a);
        border-radius: 4px;
        transition: width 0.3s ease;
    }
</style>
