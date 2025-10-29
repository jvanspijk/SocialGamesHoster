<script lang="ts">
    import type { GetPlayerResponse } from '$lib/client';
    import type { PageProps } from './$types';
    import { onMount } from 'svelte';
    import Description from '$lib/components/Description.svelte';
    import TimeDisplay from '$lib/components/TimeDisplay.svelte';

    let timer: TimeDisplay | undefined = undefined;
    const TOTAL_DURATION = 60;
    let remainingTime = $state(TOTAL_DURATION);
    let timerFinished = $state(false);

    let { data }: PageProps = $props();
    const playerData: GetPlayerResponse | undefined = data.player;

    onMount(() => {
        const mainElement = document.querySelector('main');
        if (mainElement) {
            mainElement.style.minHeight = '100vh';
        }
    });
</script>

<main>
    <div class="timer">
        <TimeDisplay 
            bind:this={timer} 
            bind:remainingTime={remainingTime} 
            initialSeconds={TOTAL_DURATION} 
            onFinished={() => timerFinished = true}
        />
    </div>

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
</main>

<style>
    .timer {
        position: fixed;        
        top: 20px; 
        right: 20px;        
        z-index: 1000;       
        transform: scale(0.7);         
        transform-origin: top right;
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