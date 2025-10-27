<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();	
</script>

<div class="page-main">
    <div class="parchment-container">        
        {@render children?.()}
    </div>
</div>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>



<style>
    /* 1. Import Fantasy-Style Google Fonts */
    @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@700&family=IM+Fell+English:ital@0;1&display=swap');

    /* 2. Page & Background Styles (Paper/Parchment Look) */
    :global(body) {
        /* This creates the dark, atmospheric background */
        background-color: #3b332d;
        margin: 0;
        padding: 0;
        min-height: 100vh;
    }

    .page-main {
        display: flex;
        justify-content: center;
        padding: 40px 20px;
    }

    .parchment-container {
        /* The main parchment look */
        background-color: #f7e7c4;
        color: #3e322b;
        font-family: 'IM Fell English', serif;
        width: 100%;
        max-width: 700px; /* Kept the width from the previous example */
        /* Shadow and border for that antique book/map look */
        box-shadow: 0 4px 10px rgba(0, 0, 0, 0.4), 
                    inset 0 0 10px rgba(100, 80, 50, 0.5);
        padding: 40px;
        border: 10px solid #5b4a3c;
        border-radius: 5px;
        line-height: 1.6;
        min-height: 80vh; /* Ensure it covers a good part of the screen */
    }

    /* 3. Common Typography (Applies to all content within .parchment-container) */

    :global(h1), :global(h2) {
        font-family: 'Cinzel', serif;
        color: #5b4a3c;
        text-transform: uppercase;
        letter-spacing: 2px;
        margin-bottom: 10px;
        text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.5);
    }

    :global(h1) {
        font-size: 2.5em;
        text-align: center;
        border-bottom: 3px double #5b4a3c;
        padding-bottom: 15px;
        margin-bottom: 20px;
    }

    :global(h2) {
        font-size: 1.5em;
        border-bottom: 1px solid #5b4a3c;
        padding-bottom: 5px;
        margin-top: 25px;
    }

    /* 4. Reusable Utility Styles (Good to keep in layout if used often) */
    :global(.role-title) {
        font-style: italic;
        text-align: center;
        margin-top: -10px;
        margin-bottom: 30px;
        color: #7a634e;
    }

    :global(.detail-card), :global(.ruleset-info) {
        padding: 15px;
        background-color: #fff9e6;
        border: 2px solid #5b4a3c;
        border-radius: 3px;
    }

    :global(.error-message) {
        font-family: 'Cinzel', serif;
        color: #a62a2a;
        text-align: center;
        font-size: 1.5em;
        padding: 50px;
    }
</style>


