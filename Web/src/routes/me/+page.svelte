<script lang="ts">
    import type { GetPlayerResponse } from '$lib/client/Players/GetPlayer';
    import type { PageProps } from './$types';
    import { onMount } from 'svelte';
    import Description from '$lib/components/Description.svelte';
    import TimeDisplay from '$lib/components/TimeDisplay.svelte';
    import HUDFooter from '$lib/components/HUDFooter.svelte';
	import { playerHub } from '$lib/client/Players/Hub.svelte';
	import { invalidateAll } from '$app/navigation';

    let timer: TimeDisplay | undefined = undefined;
    const TOTAL_DURATION = 120; // temporary hard coded value
    let remainingTime = $state(TOTAL_DURATION);
    let timerFinished = $state(false);

    let { data }: PageProps = $props();
    const playerData: GetPlayerResponse | undefined = $derived(data.player);

    playerHub.onEvent('PlayerUpdated', (event) => {
        if (event.data == data.player?.id) {
            console.log("Refreshing data for player:", event.data);
            invalidateAll();
        }
    });

    onMount(() => {
        const mainElement = document.querySelector('main');
        if (mainElement) {
            mainElement.style.minHeight = '100vh';
        }
    });
</script>

<main>
    {#if playerData}
        <div class="character-sheet">                   
            <header class="sheet-header">
                <h1>{playerData.name}</h1>
                <h3 class="role-title">The {playerData.role?.name || 'Unassigned'}</h3>
            </header>

            <Description text={playerData.role?.description} isHidden={!playerData.role}/>

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

    <HUDFooter>
        <a href="/admin">Go to admin</a>
        <TimeDisplay 
            bind:this={timer} 
            bind:remainingTime={remainingTime} 
            initialSeconds={TOTAL_DURATION} 
            onFinished={() => timerFinished = true}
        />        
    </HUDFooter>
</main>

<style>
    .character-sheet {
        position: relative;
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