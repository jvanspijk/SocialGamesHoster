<script lang="ts">
    export let value: string = '';
    export let placeholder: string = 'Enter text...';
    export let name: string = '';
    export let type: 'text' | 'password' = 'text';
</script>

<input 
    {type}
    {name} 
    {placeholder} 
    bind:value
    on:keydown
/>

<style>
    input {
        font-family: 'IM Fell English', serif;
        font-size: 1.2em;
        color: #3e322b;
        
        padding: 10px 15px;
        width: 100%;
        max-width: 400px;
        
        background-color: #fff9e6;
        border: 2px solid #5b4a3c;
        border-radius: 3px;
        box-shadow: 
            inset 0 1px 3px rgba(0, 0, 0, 0.2);
        
        margin-top: 10px; 
        margin-bottom: 10px;
    }

    input::placeholder {
        color: #7a634e;
        opacity: 1;
    }

    input:focus {
        outline: none;
        border-color: #8c2a3e;
        box-shadow: 0 0 5px rgba(140, 42, 62, 0.5);
    }
</style>