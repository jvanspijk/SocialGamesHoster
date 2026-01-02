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
		background-color: white;
		padding: 1.5rem;
		border-radius: 0.75rem;
		border: 1px solid #e2e8f0;
		box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		transition: border-color 0.2s;
	}

	.card:hover {
		border-color: #cbd5e1;
	}

	.body {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.title {
		font-size: 1.125rem;
		font-weight: 700;
		color: #0f172a;
		margin: 0;
	}

	.description {
		font-size: 0.875rem;
		color: #64748b;
		margin: 0.25rem 0 0 0;
	}

	.tag-pool {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.tag {
		background-color: #eef2ff;
		color: #4338ca;
		padding: 0.125rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		border: 1px solid #e0e7ff;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
</style>