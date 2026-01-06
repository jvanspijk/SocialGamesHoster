<script lang="ts" generics="T extends { id: string; name: string }">

	let { items, selectedIds, onToggle } = $props();
</script>

<div class="pool">
	{#each items as item}
		{@const isActive = selectedIds.includes(item.id)}
		<button 
			type="button"
			onclick={() => onToggle(item.id)}
			class="tag {isActive ? 'active' : ''}"
		>
			{item.name}
			{#if isActive}
				<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4" stroke-linecap="round" stroke-linejoin="round">
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
        gap: 0.75rem;
        min-height: 2.5rem;
        padding: 0.5rem 0;
    }

    .tag {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.4rem 0.8rem;
        
        font-family: 'IM Fell English', serif;
        font-size: 1rem;
        font-weight: 500;
        cursor: pointer;
        transition: all 0.15s ease-in-out;
        
        background-color: #f1e9d2;
        color: #5b4a3c;
        border: 1px solid #dcd3b6;
        border-radius: 16px;
        box-shadow: 1px 1px 2px rgba(0, 0, 0, 0.05);
    }

    .tag:hover {
        border-color: #5b4a3c;
        background-color: #e8e0c5;
        transform: translateY(-1px);
    }

    .tag.active {
        background-color: #8c2a3e;
        border-color: #5b1a29;
        color: #fff9e6;
        box-shadow: 2px 2px 4px rgba(140, 42, 62, 0.3);
    }

    .tag.active:hover {
        background-color: #a6324a;
    }

    .empty-text {
        font-family: 'IM Fell English', serif;
        font-size: 0.95rem;
        color: #fff9e6;
        font-style: italic;
        margin: 0.5rem 0;
    }
</style>