<script lang="ts">
	import { resolve } from '$app/paths';
	import { Clock3, Shield, Users } from '@lucide/svelte';
	import AttentionCard from '$lib/components/AttentionCard.svelte';
	import TimerDisplay from '$lib/components/TimerDisplay.svelte';
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
					{#if view.game.status === 'lobby'}
						You are in the lobby. The game master will start the game.
					{:else if view.game.status === 'paused'}
						The game is paused.
					{:else if view.game.status === 'review'}
						The game master is finishing the game.
					{:else}
						Follow the game master's instructions.
					{/if}
				</p>
			</div>
			<div class="timer-wrap">
				<Clock3 size={22} aria-hidden="true" />
				<TimerDisplay gameId={view.game.id} />
			</div>
			<div class="quick-actions">
				<!-- The disabled role state intentionally has no destination. -->
				<!-- eslint-disable svelte/no-navigation-without-resolve -->
				<a
					href={view.roleAvailable ? resolve('/play/role') : undefined}
					class:disabled={!view.roleAvailable}
					aria-disabled={!view.roleAvailable}
				>
					<Shield size={23} />
					<span
						><strong>{view.roleAvailable ? 'View role' : 'Role unavailable'}</strong><small
							>{view.roleAvailable
								? 'Open your private role screen'
								: 'The game master has not made roles available'}</small
						></span
					>
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
				<a href={resolve('/play/party')}>
					<Users size={23} />
					<span><strong>View party</strong><small>{view.party.length} players</small></span>
				</a>
			</div>
		{/if}
	</section>
{/if}

<style>
	.game-stage {
		display: grid;
		width: min(100%, 48rem);
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
		font-size: clamp(2.5rem, 12vw, 5.5rem);
		overflow-wrap: anywhere;
	}

	.eyebrow {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.16em;
		text-transform: uppercase;
	}

	.timer-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
	}

	.quick-actions {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
		text-align: start;
	}

	.quick-actions a {
		display: grid;
		min-height: 5.5rem;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		border: var(--border-subtle);
		background: rgb(255 249 230 / 62%);
		color: var(--ink);
		padding: var(--space-3);
		text-decoration: none;
	}

	.quick-actions a.disabled {
		cursor: not-allowed;
		opacity: 0.55;
	}

	.quick-actions strong,
	.quick-actions small {
		display: block;
	}

	.quick-actions small {
		color: var(--ink-soft);
	}

	.attention-stage {
		display: grid;
		place-items: center;
	}

	@media (max-width: 35rem) {
		.quick-actions {
			grid-template-columns: 1fr;
		}
	}
</style>
