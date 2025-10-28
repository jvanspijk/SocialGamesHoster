<script lang="ts">
    import type { AbilityInfo } from '$lib/client'
    import type { GetPlayerResponse } from '$lib/client';
    import type { PageProps } from './$types';
    import { onMount } from 'svelte';
    import DetailCard from '$lib/components/DetailCard.svelte';

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
                <h3 class="role-title">The {playerData.role?.name || 'Unassigned'}</h3>
            </header>

            <section class="details-section">
                {#if playerData.role}
                    <DetailCard title={playerData.role.name} content={playerData.role.description}/>                    
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
        margin-top: 10px;
    }
    
    @media (min-width: 600px) {
        .details-section {
            grid-template-columns: 1fr 1.5fr;
        }
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