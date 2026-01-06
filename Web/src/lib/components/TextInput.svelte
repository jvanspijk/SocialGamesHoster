<script lang="ts">
    let { 
        value = $bindable(''), 
        placeholder = 'Enter text...', 
        name = '', 
        type = 'text',
        multiline = false,
        rows = 3,
        onkeydown = undefined
    } = $props();
</script>

{#if multiline}
    <textarea
        {name}
        {placeholder}
        {rows}
        bind:value={value}
        {onkeydown}
    ></textarea>
{:else}
    <input 
        {type}
        {name} 
        {placeholder} 
        bind:value={value}
        {onkeydown}
    />
{/if}

<style>
    input, textarea {
        font-family: 'IM Fell English', serif;
        font-size: 1.2em;
        color: #3e322b;
        padding: 10px 15px;
        width: 100%;
        max-width: 400px;
        background-color: #fff9e6;
        border: 2px solid #5b4a3c;
        border-radius: 3px;
        box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
        margin-top: 10px; 
        margin-bottom: 10px;
        display: block;
        box-sizing: border-box;
    }

    textarea {
        resize: vertical;
    }

    input::placeholder, textarea::placeholder {
        color: #7a634e;
        opacity: 1;
    }

    input:focus, textarea:focus {
        outline: none;
        border-color: #8c2a3e;
        box-shadow: 0 0 5px rgba(140, 42, 62, 0.5);
    }
</style>