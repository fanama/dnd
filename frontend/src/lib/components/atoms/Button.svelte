<script>
    export let variant = 'primary';
    export let onClick = () => {};
    export let className = '';
</script>

<button class="{variant} {className}" on:click={onClick}>
    <slot />
</button>

<style>
    button {
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
    }

    .primary {
        background: linear-gradient(to bottom, #5c4033, #3d2b1f);
        border-color: #7c6243;
        box-shadow: 0 4px 0px #2d2319;
    }

    .danger {
        background: linear-gradient(to bottom, #a11b1b, #5f1010);
        color: #ffcdd2;
        border-color: #b71c1c;
        box-shadow: 0 4px 0px #3e0a0a;
    }

    .success {
        background: linear-gradient(to bottom, #2e7d32, #1b5e20);
        color: #e8f5e9;
        border-color: #388e3c;
        box-shadow: 0 4px 0px #0f3d0f;
    }

    .arcane {
        background: linear-gradient(to bottom, #4527a0, #281566);
        color: #e1bee7;
        border-color: #512da8;
        box-shadow: 0 4px 0px #1a0a4d;
    }

    button:hover {
        color: #fff;
        filter: brightness(1.1);
        top: -2px;
        box-shadow: 0 6px 0px rgba(0,0,0,0.5);
    }

    button:active {
        top: 2px;
        box-shadow: 0 2px 0px rgba(0,0,0,0.5);
    }
</style>
