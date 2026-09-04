<script>
    export let value = '';
    export let placeholder = '';
    export let label = '';
    export let error = '';
    export let disabled = false;
    export let id = undefined;

    $: hasError = error && error.length > 0;
</script>

<div class="input-wrapper">
    {#if label}
        <label for={id} class="input-label" class:has-error={hasError}>
            {label}
        </label>
    {/if}

    <div class="input-container" class:has-error={hasError} class:is-disabled={disabled}>
        <input
            {id}
            type="text"
            {placeholder}
            {disabled}
            bind:value
            class="input-field"
            aria-invalid={hasError}
            aria-describedby={hasError ? `${id}-error` : undefined}
        />
    </div>

    {#if hasError}
        <p id="{id}-error" class="input-error">{error}</p>
    {/if}
</div>

<style>
    .input-wrapper {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }

    .input-label {
        font-family: 'MedievalSharp', cursive;
        font-size: 0.8rem;
        color: #c5a059;
        text-transform: uppercase;
        letter-spacing: 0.08em;
    }

    .input-label.has-error {
        color: #ef4444;
    }

    .input-container {
        position: relative;
        border-radius: 10px;
        overflow: hidden;
    }

    .input-container.has-error .input-field {
        border-color: #ef4444;
        box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.2);
    }

    .input-container.is-disabled {
        opacity: 0.5;
        pointer-events: none;
    }

    .input-field {
        width: 100%;
        padding: 12px 16px;
        background: #2b221a;
        border: 2px solid #574f3e;
        color: #f5f0e8;
        border-radius: 10px;
        font-family: 'Alegreya', serif;
        font-size: 1rem;
        transition: all 0.2s ease;
        outline: none;
        box-sizing: border-box;
    }

    .input-field::placeholder {
        color: #7a6f5f;
        font-style: italic;
    }

    .input-field:focus {
        border-color: #d4af37;
        box-shadow: 0 0 0 3px rgba(212, 175, 55, 0.2);
        background: #322a1f;
    }

    .input-field:disabled {
        cursor: not-allowed;
    }

    .input-error {
        font-family: 'Alegreya', serif;
        font-size: 0.8rem;
        color: #ef4444;
        margin: 0;
        padding-left: 4px;
    }
</style>
