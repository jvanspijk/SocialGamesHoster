<script lang="ts">
	import { resolve } from '$app/paths';
	import Clock3 from '@lucide/svelte/icons/clock-3';
	import Shield from '@lucide/svelte/icons/shield';
	import Users from '@lucide/svelte/icons/users';
	import NavigationCards from '$lib/components/NavigationCards.svelte';
	import AttentionCard from '$lib/features/play/components/AttentionCard.svelte';
	import TimerDisplay from '$lib/features/games/components/TimerDisplay.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let acknowledging = $state(false);
	const view = $derived(gameState.player);
	const phase = $derived(
		view?.game.phaseKey || (view?.game.status === 'lobby' ? 'Lobby' : 'Waiting for phase')
	);

	async function acknowledge() {
		if (!view || view.attentionItems.length === 0) return;
		acknowledging = true;
		try {
			await api(`/games/${view.game.id}/announcements/${view.attentionItems[0].id}/acknowledge`, {
				method: 'POST'
			});
			await gameState.refreshPlayer();
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The announcement could not be acknowledged.'));
		} finally {
			acknowledging = false;
		}
	}
</script>

{#if view}
	<section class="game-stage">
		{#if view.attentionItems.length > 0}
			<div class="attention-stage">
				<AttentionCard
					item={view.attentionItems[0]}
					position={1}
					total={view.attentionItems.length}
					{acknowledge}
					busy={acknowledging}
				/>
			</div>
		{:else}
			<div class="phase-copy">
				<p class="eyebrow">Current phase</p>
				<h1>{phase}</h1>
				<p>
					{#if view.game.status === 'lobby'}You are in the lobby. The Game Master will start the
						game.
					{:else if view.game.status === 'paused'}The game is paused.
					{:else if view.game.status === 'review'}The Game Master is finishing the game.
					{:else}Follow the Game Master's instructions.{/if}
				</p>
			</div>
			<div class="timer-wrap">
				<Clock3 size={22} aria-hidden="true" /><TimerDisplay
					gameId={view.game.id}
					revision={view.game.revision}
				/>
			</div>
			<NavigationCards
				label="Game actions"
				items={[
					{
						label: view.roleAvailable ? 'View role' : 'Role unavailable',
						description: view.roleAvailable ? 'Open your private role screen' : 'Roles are hidden.',
						href: resolve('/play/role'),
						disabled: !view.roleAvailable,
						icon: Shield
					},
					{
						label: 'View party',
						description: `${view.party.length} players`,
						href: resolve('/play/party'),
						icon: Users
					}
				]}
			/>
		{/if}
	</section>
{/if}

<style>
	.game-stage {
		display: grid;
		width: 100%;
		min-height: calc(100dvh - 7.75rem);
		align-content: center;
		gap: var(--space-5);
		margin-inline: auto;
		padding: clamp(var(--space-4), 6vw, var(--space-7));
		text-align: center;
	}
	.phase-copy h1,
	.phase-copy p {
		margin: 0;
	}
	.phase-copy h1 {
		font-size: clamp(var(--font-size-2xl), 12vw, 5.5rem);
		overflow-wrap: anywhere;
	}
	.timer-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
	}
	.attention-stage {
		display: grid;
		place-items: center;
	}
</style>
