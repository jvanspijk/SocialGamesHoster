<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children, align = 'start' }: { children: Snippet; align?: 'start' | 'end' } = $props();
</script>

<div class="tag-list" class:align-end={align === 'end'}>
	{@render children()}
</div>

<style>
	.tag-list {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1);
	}

	.align-end {
		justify-content: flex-end;
	}
</style>
