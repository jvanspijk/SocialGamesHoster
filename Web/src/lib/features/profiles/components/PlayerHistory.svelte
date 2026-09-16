<script lang="ts">
	import { resolve } from '$app/paths';
	import { ArrowLeft, Settings, UserRound } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import type { PlayerHistoryView } from '$lib/api/types';

	let {
		history,
		loading,
		loadError,
		retry
	}: {
		history: PlayerHistoryView | null;
		loading: boolean;
		loadError: string;
		retry: () => void | Promise<void>;
	} = $props();
</script>

<div class="account-page">
	<PageHeading
		eyebrow="Player account"
		title="History"
		description="Review your completed games and achievements."
		variant="flush"
	>
		{#snippet actions()}
			<nav aria-label="Account pages">
				<a href={resolve('/play')}><ArrowLeft size={18} /> Player home</a>
				<a href={resolve('/play/profile')}><UserRound size={18} /> Profile</a>
				<a href={resolve('/play/settings')}><Settings size={18} /> Settings</a>
			</nav>
		{/snippet}
	</PageHeading>

	{#if loading}
		<p role="status">Loading history…</p>
	{:else if loadError}
		<Panel title="History unavailable" variant="focal">
			<div class="load-error">
				<p role="alert">{loadError}</p>
				<Button variant="secondary" onclick={retry}>Try again</Button>
			</div>
		</Panel>
	{:else if history}
		<Panel
			title="Game history"
			description={`${history.statistics.achievementPoints} achievement points across ${history.statistics.achievementCount} achievements`}
		>
			{#if history.games.length === 0}
				<p>No completed games yet.</p>
			{:else}
				<div class="history-list">
					{#each history.games as game (game.id)}
						<article>
							<div>
								<h3>{game.name}</h3>
								<p>{game.rulesetName} · {game.roleName || 'No role'}</p>
								{#if game.achievements.length > 0}
									<small
										>{game.achievements.map((achievement) => achievement.title).join(', ')}</small
									>
								{/if}
							</div>
							<strong>{game.outcome}</strong>
						</article>
					{/each}
				</div>
			{/if}
		</Panel>
	{/if}
</div>

<style>
	.account-page {
		display: grid;
		width: min(100%, 48rem);
		gap: var(--space-5);
		margin-inline: auto;
		padding: clamp(var(--space-4), 5vw, var(--space-6));
	}

	nav {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	nav a {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	.history-list article {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.history-list h3,
	.history-list p,
	.load-error p {
		margin: 0;
	}

	.history-list p,
	.history-list small {
		display: block;
		color: var(--ink-soft);
	}

	.history-list > article > strong {
		text-transform: capitalize;
	}

	.load-error {
		display: grid;
		justify-items: start;
		gap: var(--space-3);
	}
</style>
