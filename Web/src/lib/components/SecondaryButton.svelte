<script lang="ts">
    let { children, variant = 'secondary', onclick, enabled = true } = $props();
</script>

<button 
    class="secondary-btn {variant}" 
    {onclick} 
    disabled={!enabled}
    aria-disabled={!enabled}
>
    {@render children()}
</button>

<style>
    .secondary-btn {
        font-family: 'Cinzel', serif;
        font-weight: bold;
        padding: 4px 12px;
        cursor: pointer;
        border: 2px solid #5b4a3c;
        background: #fcf5e5;
        color: #5b4a3c;
        text-transform: uppercase;
        font-size: 0.8rem;
        transition: all 0.1s ease;
        box-shadow: 2px 2px 0px #5b4a3c;
    }

    .secondary-btn:not(:disabled):active {
        transform: translate(1px, 1px);
        box-shadow: 0px 0px 0px;
    }

    .secondary-btn:not(:disabled).primary:hover {
        background: #5b4a3c;
        color: #f7e7c4;
    }

    .secondary-btn:not(:disabled).danger:hover {
        background: #a62a2a;
        color: white;
    }

    .danger {
        color: #a62a2a;
        border-color: #a62a2a;
    }

    .secondary-btn:disabled {
        cursor: not-allowed;
        filter: grayscale(0.8);
        opacity: 0.5;
        box-shadow: 1px 1px 0px #5b4a3c;
        border-style: dashed;
    }
</style>