<script lang="ts">
	import AttentionBadge from '$lib/components/AttentionBadge.svelte';
	import type { UnreadCounts } from '../unread.svelte';

	let {
		counts,
		placement = 'corner'
	}: {
		counts: Promise<UnreadCounts> | null;
		placement?: 'corner' | 'navigation';
	} = $props();
</script>

{#await counts then loaded}
	{#if loaded && loaded.total > 0}
		<AttentionBadge count={loaded.total} {placement} />
	{/if}
{:catch}
	<!-- An optional badge must not prevent the surrounding page from rendering. -->
{/await}
