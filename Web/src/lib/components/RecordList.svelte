<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<ul class="record-list">
	{@render children()}
</ul>

<style>
	.record-list {
		margin: 0;
		padding: 0;
		list-style: none;
	}
</style>
