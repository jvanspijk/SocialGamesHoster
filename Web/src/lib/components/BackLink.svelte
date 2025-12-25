<script lang="ts">
    import { goto } from '$app/navigation';
    
    export let href = '/';
    export let className = '';
    export let pageName = '';

    function handleNavigation(e: MouseEvent) {
        // Prevent actual anchor navigation to let Svelte handle it
        e.preventDefault();
        goto(href);
    }
</script>

<a 
    href={href} 
    class="back-link {className}" 
    on:click={handleNavigation}
>
    ← <slot>Back to {pageName}</slot>
</a>

<style>
    .back-link {
        font-family: 'Cinzel', serif;
        color: #8b7355;
        text-decoration: none;
        font-size: 0.9rem;
        transition: color 0.2s;
    }

    .back-link:hover { color: #5b4a3c; }
</style>