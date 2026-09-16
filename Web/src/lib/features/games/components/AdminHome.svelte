<script lang="ts">
	import { resolve } from '$app/paths';
	import { Gamepad2, ScrollText, Settings, Swords, UsersRound } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import NavigationCards from '$lib/components/NavigationCards.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { gameStatusLabel } from '$lib/gamePresentation';
	import type { Game } from '$lib/api/types';

	let {
		games,
		loading,
		loadError,
		retry
	}: {
		games: Game[];
		loading: boolean;
		loadError: string;
		retry: () => void | Promise<void>;
	} = $props();

	const activeGame = $derived(
		games.find((game) => ['lobby', 'running', 'paused', 'review'].includes(game.status)) ?? null
	);
	const activeHref = $derived(
		activeGame?.status === 'review'
			? resolve(`/admin/games/${activeGame.id}/finish/outcomes`)
			: activeGame
				? resolve(`/admin/games/${activeGame.id}/overview`)
				: ''
	);
	const shortcuts = [
		{
			label: 'Games',
			description: 'Create, resume, and review game sessions.',
			href: resolve('/admin/games'),
			icon: Gamepad2
		},
		{
			label: 'Rulesets',
			description: 'Manage reusable game rules and content.',
			href: resolve('/admin/rulesets'),
			icon: ScrollText
		},
		{
			label: 'Profiles',
			description: 'Approve requests and manage player profiles.',
			href: resolve('/admin/approvals'),
			icon: UsersRound
		},
		{
			label: 'Settings',
			description: 'Configure this host and its display.',
			href: resolve('/admin/settings/network'),
			icon: Settings
		}
	];
</script>

<PageHeading
	eyebrow="Management"
	title="Home"
	description="Open the current game or choose an area to manage."
/>

<div class="home-content">
	{#if loading}
		<Panel variant="focal"><p role="status">Loading home…</p></Panel>
	{:else if loadError}
		<Panel title="Home unavailable" variant="focal">
			<div class="load-error">
				<p role="alert">{loadError}</p>
				<Button variant="secondary" onclick={retry}>Try again</Button>
			</div>
		</Panel>
	{:else if activeGame}
		<Panel title="Active game" variant="focal">
			<div class="active-game">
				<div class="seal" aria-hidden="true"><Swords size={29} /></div>
				<div>
					<StatusBadge label={gameStatusLabel(activeGame.status)} />
					<h2>{activeGame.name}</h2>
					<p>{activeGame.playerCount ?? 0} players</p>
				</div>
				<!-- The destination is derived from the active game's status. -->
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a class="primary-link" href={activeHref}
					>{activeGame.status === 'review' ? 'Continue finishing' : 'Open game'}</a
				>
			</div>
		</Panel>
	{:else}
		<Panel title="No active game">
			<p class="empty">Create or open a game from Games when your group is ready.</p>
		</Panel>
	{/if}

	<NavigationCards items={shortcuts} />
</div>

<style>
	.home-content {
		display: grid;
		gap: var(--space-5);
	}

	.active-game {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-4);
	}

	.seal {
		display: grid;
		width: 3.5rem;
		height: 3.5rem;
		place-items: center;
		border: 2px double var(--gold);
		background: var(--ink);
		color: var(--gold-light);
	}

	h2,
	p {
		margin: 0;
	}

	.active-game p,
	.empty {
		color: var(--ink-soft);
	}

	.primary-link {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		border: 1px solid var(--crimson-dark);
		background: var(--crimson);
		box-shadow: 0 3px 0 var(--crimson-dark);
		color: var(--paper-light);
		font-family: var(--font-display);
		font-size: 0.76rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		padding: var(--space-2) var(--space-4);
		text-decoration: none;
		text-transform: uppercase;
	}

	.load-error {
		display: grid;
		justify-items: start;
		gap: var(--space-3);
	}

	@media (max-width: 35rem) {
		.active-game {
			grid-template-columns: auto minmax(0, 1fr);
		}

		.active-game .primary-link {
			grid-column: 1 / -1;
			width: 100%;
		}
	}
</style>
