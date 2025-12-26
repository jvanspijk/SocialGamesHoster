<script lang="ts">
    import type { Snippet } from 'svelte';

    type ButtonType = 'button' | 'submit' | 'reset';

    interface Props {
        children: Snippet;
        variant?: 'secondary' | 'primary' | 'danger';
        onclick?: (e: MouseEvent) => void;
        enabled?: boolean;
        type?: ButtonType;
    }

    let { 
        children, 
        variant = 'secondary', 
        onclick, 
        enabled = true,
        type = 'button' 
    }: Props = $props();
</script>

<button 
    {type} 
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
        padding: 2px 10px; 
        cursor: pointer;
        border: 1px solid #5b4a3c;
        background: #fcf5e5;
        color: #5b4a3c;
        text-transform: uppercase;
        font-size: 0.77rem;
        transition: all 0.1s ease;
        box-shadow: 1px 1px 0px #5b4a3c;
        line-height: 2.4;
        height: fit-content;
    }

    .secondary-btn:not(:disabled):active {
        transform: translate(0.5px, 0.5px);
        box-shadow: inset 1px 1px 2px rgba(0,0,0,0.1);
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