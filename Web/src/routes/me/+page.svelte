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

    .role-description {
        font-style: italic;
        color: #4f4135;
        margin-top: 10px;
    }
    
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
</style>