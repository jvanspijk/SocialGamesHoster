<script lang="ts">
	import PageHeading from '$lib/components/PageHeading.svelte';
	import AnnouncementPanel from '$lib/features/games/components/AnnouncementPanel.svelte';
	import { gameState } from '$lib/state/game.svelte';

	const view = $derived(gameState.admin);
	const live = $derived(!!view && ['lobby', 'running', 'paused'].includes(view.game.status));
</script>

{#if view}
	<div class="announcements-page">
		<PageHeading title="Announcements" variant="flush" />
		{#if live}
			<AnnouncementPanel {view} />
		{:else}
			<p>Announcements can be sent while the game is in the lobby, running or paused.</p>
		{/if}
	</div>
{/if}

<style>
	.announcements-page {
		max-width: 48rem;
		margin-inline: auto;
		display: grid;
		gap: var(--space-4);
	}
</style>
