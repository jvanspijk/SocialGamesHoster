<script lang="ts">
    import TextInput from '$lib/components/TextInput.svelte';
    import MainButton from '$lib/components/MainButton.svelte';

    let playerName: string = $state("");
    
    let isLoading: boolean = $state(false);
    let message: { text: string; type: 'success' | 'error' | '' } = $state({ text: '', type: '' });
</script>

<form method="POST" action="?/login" class="container"> 
    <h1>Join game</h1>
    <p>Enter your name to join.</p>

    {#if message.text}
        <p class={`message ${message.type}`}>
            {message.text}
        </p>
    {/if}

    <TextInput
        name="playerName"
        placeholder="Enter your name..."
        bind:value={playerName}
    />
            
    <MainButton
        label="Login"
        type="submit"
        disabled={isLoading || playerName.trim() === ""} 
        isLoading={isLoading}
    />
</form>