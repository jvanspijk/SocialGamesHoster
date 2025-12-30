<script lang="ts">
    import MainButton from "$lib/components/MainButton.svelte";
    import TextInput from "$lib/components/TextInput.svelte";
    import { enhance } from '$app/forms';

    export let form;

    let isLoading = false;
</script>

<div class="container">
    <h1>Admin Login</h1>

    {#if form?.success === false}
        <p class="error">{form.message}</p>
    {/if}

    <form 
        method="POST" 
        action="?/login" 
        use:enhance={() => {
            isLoading = true;
            return async ({ update }) => {
                isLoading = false;
                await update();
            };
        }}
    >
        <TextInput placeholder="Enter username" name="name"/>
        <TextInput placeholder="Enter password" name="password" type="password" />
        <MainButton 
            type="submit" 
            label="Login" 
            {isLoading} 
        />
    </form>
</div>

<style>
    .container {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        width: 100%;
        text-align: center;
    }

    form {
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 100%;
    }

    .error {
        color: red;
    }
</style>