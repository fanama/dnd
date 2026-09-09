<script>
    import { afterUpdate } from 'svelte';

    export let logs = [];

    let chatEl;
    let autoScroll = true;

    afterUpdate(() => {
        if (autoScroll && chatEl) {
            chatEl.scrollTop = chatEl.scrollHeight;
        }
    });

    function handleScroll() {
        if (!chatEl) return;
        const { scrollTop, scrollHeight, clientHeight } = chatEl;
        autoScroll = scrollHeight - scrollTop - clientHeight < 40;
    }
</script>

<div class="dnd-section chat-box">
    <h3 class="section-title"><span>📜</span> Journal du Donjon</h3>
    <div class="chat-log custom-scrollbar" bind:this={chatEl} on:scroll={handleScroll}>
        {#each logs as msg, i}
            <div class="chat-entry">
                <span class="chat-num">{i + 1}</span>
                <span class="chat-msg">{msg}</span>
            </div>
        {:else}
            <div class="empty-chat">
                <span class="empty-icon">📜</span>
                <span class="empty-text">Le journal est vide...</span>
            </div>
        {/each}
    </div>
</div>

<style>
    .chat-box {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }

    .section-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: 'MedievalSharp', cursive;
        color: #c5a059;
        font-size: 0.95rem;
        margin: 0;
        padding-bottom: 8px;
        border-bottom: 1px solid rgba(197, 160, 89, 0.12);
    }

    .chat-log {
        display: flex;
        flex-direction: column;
        gap: 4px;
        max-height: 300px;
        overflow-y: auto;
    }

    .chat-entry {
        display: flex;
        gap: 8px;
        padding: 4px 8px;
        border-radius: 4px;
        transition: background 0.15s ease;
    }

    .chat-entry:hover {
        background: rgba(197, 160, 89, 0.05);
    }

    .chat-num {
        font-family: 'Cinzel', serif;
        color: #5a5045;
        font-size: 0.65rem;
        min-width: 24px;
        text-align: right;
        flex-shrink: 0;
        padding-top: 2px;
    }

    .chat-msg {
        font-family: 'Alegreya', serif;
        color: #c5c0b4;
        font-size: 0.8rem;
        line-height: 1.4;
    }

    .empty-chat {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        padding: 28px 16px;
    }

    .empty-icon { font-size: 2rem; opacity: 0.2; }

    .empty-text {
        font-family: 'Alegreya', serif;
        color: #5a5045;
        font-style: italic;
        font-size: 0.85rem;
    }
</style>
