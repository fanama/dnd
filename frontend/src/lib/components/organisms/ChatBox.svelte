<script>
    export let logs = [];
    let box;
    let showScrollHint = false;

    $: if (logs && box) {
        setTimeout(() => {
            if (box) {
                const isNearBottom = box.scrollHeight - box.scrollTop - box.clientHeight < 100;
                if (isNearBottom || logs.length <= 1) {
                    box.scrollTop = box.scrollHeight;
                    showScrollHint = false;
                } else {
                    showScrollHint = true;
                }
            }
        }, 0);
    }

    function scrollToBottom() {
        if (box) {
            box.scrollTop = box.scrollHeight;
            showScrollHint = false;
        }
    }

    function handleScroll() {
        if (box) {
            const isNearBottom = box.scrollHeight - box.scrollTop - box.clientHeight < 100;
            showScrollHint = !isNearBottom;
        }
    }
</script>

<div class="chat-card dnd-section">
    <div class="chat-header">
        <div class="chat-title-row">
            <span class="chat-icon">📜</span>
            <h3 class="chat-title">Journal des Evenements</h3>
        </div>
        <span class="chat-subtitle">Archives du Monde</span>
    </div>

    <div class="chat-log-wrapper">
        <div
            id="box"
            bind:this={box}
            class="chat-log custom-scrollbar"
            on:scroll={handleScroll}
        >
            {#if logs.length === 0}
                <div class="chat-empty">
                    <span class="chat-empty-icon">📜</span>
                    <span class="chat-empty-text">Le journal est vide...</span>
                    <span class="chat-empty-hint">Les evenements apparaitront ici</span>
                </div>
            {:else}
                {#each logs as log, i}
                    <div class="log-entry">
                        <span class="log-index">{i + 1}</span>
                        <span class="log-text">{log}</span>
                    </div>
                {/each}
            {/if}
        </div>

        {#if showScrollHint}
            <button class="scroll-hint" on:click={scrollToBottom}>
                <span>↓</span> Nouveaux messages
            </button>
        {/if}
    </div>
</div>

<style>
    .chat-card {
        display: flex;
        flex-direction: column;
    }

    .chat-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        flex-wrap: wrap;
        gap: 8px;
        margin-bottom: 16px;
        padding-bottom: 12px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.15);
    }

    .chat-title-row {
        display: flex;
        align-items: center;
        gap: 10px;
    }

    .chat-icon {
        font-size: 1.3rem;
        opacity: 0.8;
    }

    .chat-title {
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 1.1rem;
        margin: 0;
    }

    .chat-subtitle {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.65rem;
        color: #5a5045;
        text-transform: uppercase;
        letter-spacing: 0.1em;
    }

    .chat-log-wrapper {
        position: relative;
    }

    .chat-log {
        height: 280px;
        overflow-y: auto;
        background: linear-gradient(135deg, #f4e4bc, #e8d4a9);
        border: 2px solid rgba(92, 64, 51, 0.3);
        border-radius: 8px;
        padding: 12px;
        box-shadow: inset 0 2px 8px rgba(0, 0, 0, 0.15);
    }

    .log-entry {
        display: flex;
        gap: 10px;
        padding: 8px 8px;
        border-bottom: 1px solid rgba(112, 66, 20, 0.1);
        transition: background 0.15s ease;
        align-items: flex-start;
    }

    .log-entry:last-child {
        border-bottom: none;
    }

    .log-entry:hover {
        background: rgba(197, 160, 89, 0.12);
        border-radius: 4px;
    }

    .log-index {
        font-family: 'Cinzel', serif;
        font-size: 0.65rem;
        color: #a67c37;
        opacity: 0.5;
        min-width: 20px;
        text-align: right;
        padding-top: 2px;
        flex-shrink: 0;
    }

    .log-text {
        font-family: 'Alegreya', serif;
        font-size: 0.9rem;
        color: #2c1e16;
        line-height: 1.5;
    }

    .chat-empty {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        gap: 8px;
    }

    .chat-empty-icon {
        font-size: 3rem;
        opacity: 0.2;
    }

    .chat-empty-text {
        font-family: 'Alegreya', serif;
        color: #8b6544;
        font-style: italic;
    }

    .chat-empty-hint {
        font-family: 'Alegreya', serif;
        color: #a67c37;
        font-size: 0.8rem;
        opacity: 0.6;
    }

    .scroll-hint {
        position: absolute;
        bottom: 8px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(0, 0, 0, 0.75);
        color: #c5a059;
        border: 1px solid rgba(197, 160, 89, 0.3);
        border-radius: 100px;
        padding: 6px 16px;
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        cursor: pointer;
        transition: all 0.2s ease;
        backdrop-filter: blur(8px);
        display: flex;
        align-items: center;
        gap: 4px;
        z-index: 10;
    }

    .scroll-hint:hover {
        background: rgba(0, 0, 0, 0.9);
        border-color: #c5a059;
    }
</style>
