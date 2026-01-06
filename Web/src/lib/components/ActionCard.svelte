<script lang="ts">
	import type { Snippet } from 'svelte';

	export let title;
    export let description: string | null = null;
    export let tags: string[] = [];
    export let actions: Snippet | null = null;
</script>

<div class="card">
	<div class="body">
		<header>
			<h4 class="title">{title}</h4>
			{#if description}
				<p class="description">{description}</p>
			{/if}
		</header>

		{#if tags.length > 0}
			<div class="tag-pool">
				{#each tags as tag}
					<span class="tag">{tag}</span>
				{/each}
			</div>
		{/if}
	</div>

	{#if actions}
		<div class="actions">
			{@render actions()}
		</div>
	{/if}
</div>

<style>
    .card {
        /* Parchment Theme */
        background-color: #fff9e6;
        padding: 1.5rem;
        border-radius: 3px; /* Sharper corners for old-paper feel */
        border: 2px solid #5b4a3c;
        box-shadow: 2px 2px 0px rgba(91, 74, 60, 0.2);
        
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        transition: transform 0.1s ease, box-shadow 0.1s ease;
    }

    .card:hover {
        border-color: #8c2a3e; /* Deep crimson on hover */
        transform: translateY(-1px);
        box-shadow: 3px 3px 0px rgba(140, 42, 62, 0.3);
    }

    .body {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .title {
        font-family: 'IM Fell English', serif;
        font-size: 1.4rem;
        font-weight: 700;
        color: #3e322b;
        margin: 0;
    }

    .description {
        font-family: 'IM Fell English', serif;
        font-size: 1rem;
        color: #5b4a3c;
        line-height: 1.4;
        margin: 0.25rem 0 0 0;
        font-style: italic;
    }

    .tag-pool {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }

    .tag {
        background-color: #f1e9d2;
        color: #8c2a3e;
        padding: 0.2rem 0.6rem;
        border-radius: 2px;
        font-size: 0.75rem;
        font-weight: bold;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        border: 1px solid #dcd3b6;
    }

    .actions {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
</style>