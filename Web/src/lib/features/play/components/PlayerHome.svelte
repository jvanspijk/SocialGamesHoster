<script lang="ts">
	import { resolve } from '$app/paths';
	import { DoorOpen, Gamepad2, History, Settings } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import NavigationCards from '$lib/components/NavigationCards.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import { gameStatusLabel } from '$lib/gamePresentation';
	import type { Game, PlayerGameView } from '$lib/api/types';

	let {
		displayName,
		loading,
		view,
		availableLobby,
		liveGame,
		joiningLobby,
		loadError,
		join,
		retry
	}: {
		displayName: string;
		loading: boolean;
		view: PlayerGameView | null;
		availableLobby: Game | null;
		liveGame: Game | null;
		joiningLobby: boolean;
		loadError: string;
		join: () => void | Promise<void>;
		retry: () => void | Promise<void>;
	} = $props();

	const shortcuts = $derived([
		{
			label: 'Current game',
			description: view
				? `${view.game.name} · ${gameStatusLabel(view.game.status)}`
				: loading
					? 'Checking for your current game…'
					: 'You have not joined a game.',
			href: view ? resolve('/play/game') : undefined,
			disabled: !view,
			icon: Gamepad2
		},
		{
			label: 'Join game',
			description: availableLobby
				? `${availableLobby.name} is accepting players.`
				: view
					? 'You are already in the current game.'
					: liveGame
						? `${liveGame.name} is not accepting players.`
						: loading
							? 'Checking for an open lobby…'
							: 'No lobby is open right now.',
			onclick: availableLobby ? join : undefined,
			disabled: !availableLobby || joiningLobby,
			icon: DoorOpen
		},
		{
			label: 'History',
			description: 'Review completed games and achievements.',
			href: resolve('/play/history'),
			icon: History
		},
		{
			label: 'Settings',
			description: 'Manage your profile, sound, display, and account.',
			href: resolve('/play/settings'),
			icon: Settings
		}
	]);
</script>

<div class="home-page">
	<PageHeading
		eyebrow="Player"
		title="Home"
		description={`Signed in as ${displayName}. Choose where you want to go.`}
	/>

	{#if loadError}
		<div class="error-state">
			<p role="alert">{loadError}</p>
			<Button variant="secondary" onclick={retry}>Try again</Button>
		</div>
	{/if}

	<NavigationCards items={shortcuts} />
</div>

<style>
	.home-page {
		width: min(100%, 48rem);
		margin-inline: auto;
		padding: clamp(var(--space-4), 5vw, var(--space-7));
	}

	.error-state {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border: 1px solid var(--danger);
		background: color-mix(in srgb, var(--danger) 8%, var(--paper-light));
		margin-block-end: var(--space-4);
		padding: var(--space-3);
	}

	.error-state p {
		margin: 0;
	}

	@media (max-width: 35rem) {
		.error-state {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
