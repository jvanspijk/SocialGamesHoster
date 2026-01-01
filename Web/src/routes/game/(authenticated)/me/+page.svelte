<script lang="ts">
    import type { PageProps } from './$types';
    import { onMount } from 'svelte';
    import Description from '$lib/components/Description.svelte';
    import TimeDisplay from '$lib/components/TimeDisplay.svelte';
    import HUDFooter from '$lib/components/HUDFooter.svelte';
	import { playersHub } from '$lib/client/Players/PlayersHub.svelte';
	import { invalidateAll } from '$app/navigation';
    import Spacer from '$lib/components/Spacer.svelte';
	import ChatToggle from '$lib/components/ChatToggle.svelte';
    import ChatChannel from '$lib/components/ChatChannel.svelte';

    let isChatOpen = $state(false);
    let hasNewMessages = $state(true); // Example: start with a notification

    function openChat() {
        isChatOpen = true;
        hasNewMessages = false; // Clear the dot when they open it
    }

    function closeChat() {
        isChatOpen = false;
    }

    let { data }: PageProps = $props();
    const player = $derived(data.player);
    const role = $derived(data.player.role);

    playersHub.onEvent('PlayerUpdated', (event) => {
        if (event.playerId == data.player?.id) {
            console.log("Refreshing data for player:", event.playerId);
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
    {#if player}
        <div class="character-sheet">                   
            <header class="sheet-header">
                <h1>{player.name}</h1>
                <h3 class="role-title">The {role?.name || 'Unassigned'}</h3>
            </header>

            <Description text={role?.description} isHidden={!role}/>

            {#if role?.abilities && role.abilities.length > 0}
                <section class="abilities-section">
                    <h2>Abilities</h2>
                    <ul class="abilities-list">
                        {#each role.abilities as ability}
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

    <ChatChannel 
        channelId={1} 
        channelName="Global Chat"
        readerId={data.player.id}
        isOpen={isChatOpen} 
        onClose={closeChat} 
    />

    <HUDFooter>
        <ChatToggle 
            hasNewMessage={hasNewMessages} 
            on:click={openChat} 
        />
        <Spacer/>
        <TimeDisplay 
            initialSeconds={data.timer.totalSeconds} 
            remainingTime={data.timer.remainingSeconds}
            isTimerRunning={data.timer.isRunning}
            onFinished={() => {}}
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