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
    @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@700&family=IM+Fell+English:ital@0;1&display=swap');

    /* 2. Page & Background Styles */
    :global(body) {
        background-color: #3b332d;
        margin: 0;
        padding: 0;
        min-height: 100vh;
    }

    .page-main {
        background-color: #3b332d;
        display: flex;
        justify-content: center;
        padding: 10px 14px;
        margin-top: 10px;
        text-align: center;
    }

    .parchment-container {
        background-color: #f7e7c4;
        color: #3e322b;
        font-family: 'IM Fell English', serif;
        width: 100%;
        max-width: 700px;
        box-shadow: 
            0 4px 10px rgba(0, 0, 0, 0.4), 
            inset 0 0 50px rgba(100, 80, 50, 0.3),
            inset 0 0 10px rgba(100, 80, 50, 0.5);
        padding: 20px 40px;
        border: 8px solid #5b4a3c;
        border-radius: 4px;
        line-height: 1.6;
        min-height: 85vh;     
        
    }

    /* for mobile */
    @media (max-width: 650px) {
        .parchment-container {
            padding: 15px 10px !important;
            border-width: 5px;
        }
    }

    /* 3. Common Typography*/

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
        font-size: 1.75em;
        border-bottom: 1px solid #5b4a3c;
        padding-bottom: 5px;
        margin-top: 25px;
    }

    :global(h3) {
        font-size: 1.25em;
        text-align: center;
        color: #7a634e;
        font-style: italic;
        margin-top: -8px;
        padding-bottom: 12px;
    }

    :global(p) {
        margin-top: -2px;
        margin-bottom: 4px;
        font-size: 1.1em;
    }   

    :global(.container) {       
        display: flex;
        flex-direction: column;
        align-items: center;
        max-width: 80%;
        margin: 0 auto;
    }

    :global(.error-message) {
        font-family: 'Cinzel', serif;
        color: #a62a2a;
        text-align: center;
        font-size: 1.5em;
        padding: 50px;
    }       

    :global(.success-message) {
        font-family: 'Cinzel', serif;
        color: #387038;
        border-color: #387038;
        background-color: #e7fce7;
        text-align: center;
        font-size: 1.5em;
        padding: 50px;
    }
</style>


