<script lang="ts" generics="T extends { id: string; name: string }">

	let { items, selectedIds, onToggle } = $props();
</script>

<div class="pool">
	{#each items as item}
		{@const isActive = selectedIds.includes(item.id)}
		<button 
			type="button"
			onclick={() => onToggle(item.id)}
			class="pill {isActive ? 'active' : ''}"
		>
			{item.name}
			{#if isActive}
				<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
					<polyline points="20 6 9 17 4 12"></polyline>
				</svg>
			{/if}
		</button>
	{:else}
		<p class="empty-text">No items available.</p>
	{/each}
</div>

<style>
	.pool {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		min-height: 2.5rem;
	}

	.pill {
		display: inline-flex;
		align-items: center;
		gap: 0.375rem;
		padding: 0.375rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		border: 1px solid #e2e8f0;
		background-color: white;
		color: #475569;
	}

	.pill:hover {
		border-color: #cbd5e1;
		background-color: #f8fafc;
	}

	.pill.active {
		background-color: #4f46e5;
		border-color: #4f46e5;
		color: white;
	}

	.empty-text {
		font-size: 0.875rem;
		color: #94a3b8;
		font-style: italic;
		margin: 0.5rem 0;
	}
</style>