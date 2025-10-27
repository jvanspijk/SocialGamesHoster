<script lang="ts">
    import type { AbilityInfo } from '$lib/client'
    import type { GetPlayerResponse } from '$lib/client';
    import type { PageProps } from './$types';
    import { onMount } from 'svelte';

    let { data }: PageProps = $props();
    const playerData: GetPlayerResponse | undefined = data.player;

    onMount(() => {
        const mainElement = document.querySelector('main');
        if (mainElement) {
            mainElement.style.minHeight = '100vh';
        }
    });

    function getAbilityList(abilities: AbilityInfo[] | undefined): string {
        if (!abilities || abilities.length === 0) return 'No abilities listed.';
        return abilities.map(a => a.name).join(', ');
    }
</script>

<main>
    {#if playerData}
        <div class="character-sheet">
            <header class="sheet-header">
                <h1>{playerData.name}</h1>
                <p class="role-title">The {playerData.role?.name || 'Unassigned'}</p>
            </header>

            <section class="details-section">
                <div class="detail-card player-info">
                    <h2>Player Details</h2>
                    <p><strong>Your name:</strong> {playerData.name}</p>                   
                </div>
                
                {#if playerData.role}
                    <div class="detail-card role-info">
                        <h2>Role: {playerData.role.name}</h2>
                        <p class="role-description">{playerData.role.description || 'No description provided.'}</p>
                    </div>
                {/if}
            </section>

            {#if playerData.role?.abilities && playerData.role.abilities.length > 0}
                <section class="abilities-section">
                    <h2>Abilities</h2>
                    <ul class="abilities-list">
                        {#each playerData.role.abilities as ability}
                            <li class="ability-item">
                                <strong>{ability.name}:</strong> {ability.description || 'A mysterious power...'}
                            </li>
                        {/each}
                    </ul>
                </section>
            {/if}
        </div>
    {:else}
        <p class="error-message">Player data not found.</p>
    {/if}
</main>

<style>
    /* 1. Import Fantasy-Style Google Fonts */
    @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@700&family=IM+Fell+English:ital@0;1&display=swap');

    /* 2. Page & Background Styles (Paper/Parchment Look) */
    :global(body) {
        background-color: #3b332d;
        margin: 0;
        padding: 0;
        min-height: 100vh;
    }

    main {
        display: flex;
        justify-content: center;
        padding: 40px 20px;
    }

    .character-sheet {
        background-color: #f7e7c4;
        color: #3e322b;
        font-family: 'IM Fell English', serif;
        width: 100%;
        max-width: 900px;
        box-shadow: 0 4px 10px rgba(0, 0, 0, 0.4), 
                    inset 0 0 10px rgba(100, 80, 50, 0.5);
        padding: 40px;
        border: 10px solid #5b4a3c;
        border-radius: 5px;
        line-height: 1.6;
    }

    /* 3. Typography and Headings */

    h1, h2 {
        font-family: 'Cinzel', serif;
        color: #5b4a3c;
        text-transform: uppercase;
        letter-spacing: 2px;
        margin-bottom: 10px;
        text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.5);
    }

    h1 {
        font-size: 2.5em;
        text-align: center;
        border-bottom: 3px double #5b4a3c;
        padding-bottom: 15px;
        margin-bottom: 20px;
    }

    h2 {
        font-size: 1.5em;
        border-bottom: 1px solid #5b4a3c;
        padding-bottom: 5px;
        margin-top: 25px;
    }

    .role-title {
        font-style: italic;
        text-align: center;
        margin-top: -10px;
        margin-bottom: 30px;
        color: #7a634e;
    }

    /* 4. Section Layout and Cards */

    .details-section {
        display: grid;
        grid-template-columns: 1fr;
        gap: 20px;
        margin-bottom: 30px;
    }
    
    @media (min-width: 600px) {
        .details-section {
            grid-template-columns: 1fr 1.5fr;
        }
    }

    .detail-card {
        padding: 15px;
        background-color: #fff9e6;
        border: 2px solid #5b4a3c;
        border-radius: 3px;
    }
    
    .role-description {
        font-style: italic;
        color: #4f4135;
        margin-top: 10px;
    }
    
    /* 5. Abilities List */
    
    .abilities-section {
        margin-top: 30px;
    }

    .abilities-list {
        list-style: none;
        padding: 0;
        margin-top: 15px;
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 15px;
    }

    .ability-item {
        background-color: #f0e0b8;
        padding: 10px;
        border-left: 5px solid #a62a2a;
        box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.1);
    }
    
    .ability-item strong {
        font-family: 'Cinzel', serif;
        color: #4f4135;
    }
    
    .error-message {
        font-family: 'Cinzel', serif;
        color: #a62a2a;
        text-align: center;
        font-size: 1.5em;
        padding: 50px;
    }
</style>