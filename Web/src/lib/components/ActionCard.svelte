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
        background-color: var(--color-surface-soft);
        padding: 1.5rem;
        border-radius: 3px; /* Sharper corners for old-paper feel */
        border: 2px solid var(--color-border);
        box-shadow: 2px 2px 0px rgba(91, 74, 60, 0.2);
        
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        transition: transform 0.1s ease, box-shadow 0.1s ease;
    }

    .card:hover {
        border-color: var(--color-accent); /* Deep crimson on hover */
        transform: translateY(-1px);
        box-shadow: 3px 3px 0px rgba(140, 42, 62, 0.3);
    }

    .body {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
    }

    .title {
        font-family: var(--font-body);
        font-size: 1.4rem;
        font-weight: 700;
        color: var(--color-text);
        margin: 0;
    }

    .description {
        font-family: var(--font-body);
        font-size: 1rem;
        color: var(--color-border);
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
        background-color: var(--color-surface-alt);
        color: var(--color-accent);
        padding: 0.2rem 0.6rem;
        border-radius: 2px;
        font-size: 0.75rem;
        font-weight: bold;
        text-transform: uppercase;
        letter-spacing: 0.05em;
        border: 1px solid var(--color-border-soft);
    }

    .actions {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
</style>