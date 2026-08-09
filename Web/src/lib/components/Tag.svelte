<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children, icon }: { children: Snippet; icon?: Snippet } = $props();
</script>

<span class="tag">
	{#if icon}<span class="icon" aria-hidden="true">{@render icon()}</span>{/if}
	{@render children()}
</span>

<style>
	.tag {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		border: 1px solid var(--gold-dark);
		font-size: 0.75rem;
		padding: 0.2rem 0.4rem;
	}

	.icon {
		display: inline-flex;
	}
</style>
