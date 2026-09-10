<script>
    export let variant = 'primary';
    export let onClick = () => {};
    export let className = '';
    export let disabled = false;
    export let loading = false;
    export let type = 'button';

    $: isDisabled = disabled || loading;
</script>

<button
    {type}
    class="btn btn-{variant} {className}"
    class:is-disabled={isDisabled}
    on:click={onClick}
    disabled={isDisabled}
    aria-disabled={isDisabled}
>
    {#if loading}
        <span class="spinner" aria-hidden="true"></span>
    {/if}
    <span class="btn-content" class:invisible={loading}>
        <slot />
    </span>
</button>

<style>
    .btn {
        font-family: 'MedievalSharp', cursive;
        cursor: pointer;
        padding: 10px 20px;
        border: 2px solid #7c6243;
        border-radius: 8px;
        color: #d4af37;
        font-size: 1.05rem;
        font-weight: bold;
        transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
        box-shadow: 0 4px 0px #4a3b2c;
        position: relative;
        top: 0;
        outline: none;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        min-height: 44px;
    }

    .btn:focus-visible {
        outline: 2px solid #d4af37;
        outline-offset: 2px;
    }

    .btn.primary {
        background: linear-gradient(to bottom, #5c4033, #3d2b1f);
        border-color: #7c6243;
        box-shadow: 0 4px 0px #2d2319;
    }

    .btn.danger {
        background: linear-gradient(to bottom, #a11b1b, #5f1010);
        color: #ffcdd2;
        border-color: #b71c1c;
        box-shadow: 0 4px 0px #3e0a0a;
    }

    .btn.success {
        background: linear-gradient(to bottom, #2e7d32, #1b5e20);
        color: #e8f5e9;
        border-color: #388e3c;
        box-shadow: 0 4px 0px #0f3d0f;
    }

    .btn.arcane {
        background: linear-gradient(to bottom, #4527a0, #281566);
        color: #e1bee7;
        border-color: #512da8;
        box-shadow: 0 4px 0px #1a0a4d;
    }

    .btn.warning {
        background: linear-gradient(to bottom, #d97706, #92400e);
        color: #fef3c7;
        border-color: #b45309;
        box-shadow: 0 4px 0px #5f2e05;
    }

    .btn:not(.is-disabled):hover {
        color: #fff;
        filter: brightness(1.1);
        top: -2px;
        box-shadow: 0 6px 0px rgba(0,0,0,0.5);
    }

    .btn:not(.is-disabled):active {
        top: 2px;
        box-shadow: 0 2px 0px rgba(0,0,0,0.5);
    }

    .btn.is-disabled {
        opacity: 0.5;
        cursor: not-allowed;
        filter: saturate(0.5);
        top: 0;
    }

    .spinner {
        width: 18px;
        height: 18px;
        border: 2px solid currentColor;
        border-top-color: transparent;
        border-radius: 50%;
        animation: spin 0.6s linear infinite;
        position: absolute;
    }

    .btn-content {
        display: inline-flex;
        align-items: center;
        gap: 8px;
    }

    .invisible {
        visibility: hidden;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }
</style>
