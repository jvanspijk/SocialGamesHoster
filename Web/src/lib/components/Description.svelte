<script lang="ts">
    export let text: string = ""; 
    export let isHidden: boolean = false; 
</script>

{#if !isHidden}
    <p class="description">
        {text}
    </p>
{/if}

<style>
    .description {
        background-color: #e3d8b8;
        border: 1px solid #c9b48c;
        padding: 15px;
        margin-top: 15px;
        border-radius: 3px;
        line-height: 1.5;
        font-size: 1.1em;
        text-align: left;
        color: #3e322b;
    }
</style>