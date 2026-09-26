<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		Award,
		Check,
		MessageCircle,
		ShieldAlert,
		UserMinus,
		UserRoundCheck
	} from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Field from '$lib/components/Field.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import RoleVisibilityControl from '$lib/features/games/components/RoleVisibilityControl.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Participant } from '$lib/api/types';
	import { hasAssignedRole } from '$lib/features/games/roleAssignments';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let selectedId = $state('');
	let kickConfirmOpen = $state(false);
	let busy = $state(false);
	let assignments = $state<Record<string, string>>({});
	let hydratedAssignmentsFor = $state('');

	const view = $derived(gameState.admin);
	const assignmentsHydrated = $derived(view !== null && hydratedAssignmentsFor === view.game.id);
	const selected = $derived(view?.participants.find((player) => player.id === selectedId) ?? null);
	let aliasDraft = $derived(selected?.gameAlias ?? '');
	const activePlayers = $derived(
		view?.participants.filter((player) => !['kicked', 'left'].includes(player.status)) ?? []
	);
	$effect(() => {
		if (view) {
			if (hydratedAssignmentsFor !== view.game.id) {
				hydrateAssignments(view.game.id, view.participants);
				return;
			}
			for (const player of view.participants) {
				if (!(player.id in assignments)) assignments[player.id] = player.roleKey ?? '';
			}
		}
	});

	function hydrateAssignments(gameId: string, participants: Participant[]) {
		assignments = Object.fromEntries(
			participants.map((player) => [player.id, player.roleKey ?? ''])
		);
		hydratedAssignmentsFor = gameId;
	}

	function playerName(player: Participant) {
		return player.gameAlias || player.displayNameSnapshot;
	}

	function statusLabel(status: Participant['status']) {
		return (
			{ active: 'Active', eliminated: 'Eliminated', kicked: 'Kicked', left: 'Left' } as const
		)[status];
	}

	function awardedCount(player: Participant) {
		return view?.awards.filter((award) => award.profileId === player.profileId).length ?? 0;
	}

	async function saveAssignments() {
		if (!view) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/assignments`, {
				method: 'PUT',
				...jsonBody({
					assignments: activePlayers.map((player) => ({
						participantId: player.id,
						roleId: assignments[player.id]
					}))
				})
			});
			const refreshed = await gameState.refreshAdmin(view.game.id);
			hydrateAssignments(refreshed.game.id, refreshed.participants);
			toasts.success('Role assignments saved.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Role assignments could not be saved.'));
		} finally {
			busy = false;
		}
	}

	async function saveAlias() {
		if (!view || !selected) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/participants/${selected.id}`, {
				method: 'PATCH',
				...jsonBody({ gameAlias: aliasDraft })
			});
			await gameState.refreshAdmin(view.game.id);
			toasts.success('Alias saved.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The alias could not be saved.'));
		} finally {
			busy = false;
		}
	}

	async function setOutcome(outcome: Participant['outcome']) {
		if (!view || !selected) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/outcomes`, {
				method: 'PUT',
				...jsonBody({ outcomes: [{ participantId: selected.id, outcome }] })
			});
			await gameState.refreshAdmin(view.game.id);
			toasts.success('Outcome saved.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The outcome could not be saved.'));
		} finally {
			busy = false;
		}
	}

	async function toggleAchievement(achievementId: string) {
		if (!view || !selected) return;
		const existing = view.awards.find(
			(award) => award.profileId === selected.profileId && award.achievementId === achievementId
		);
		busy = true;
		try {
			if (existing) {
				await api(`/games/${view.game.id}/achievement-awards/${existing.id}`, { method: 'DELETE' });
			} else {
				await api(`/games/${view.game.id}/achievement-awards`, {
					method: 'POST',
					...jsonBody({ participantId: selected.id, achievementId, note: '' })
				});
			}
			await gameState.refreshAdmin(view.game.id);
			toasts.success(existing ? 'Achievement revoked.' : 'Achievement awarded.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The achievement could not be updated.'));
		} finally {
			busy = false;
		}
	}

	async function participantAction(action: 'eliminate' | 'reinstate' | 'kick') {
		if (!view || !selected) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/participants/${selected.id}/${action}`, {
				method: 'POST',
				...jsonBody({})
			});
			await gameState.refreshAdmin(view.game.id);
			kickConfirmOpen = false;
			toasts.success(
				action === 'eliminate'
					? 'Player eliminated.'
					: action === 'reinstate'
						? 'Player reinstated.'
						: 'Player kicked.'
			);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The player could not be updated.'));
		} finally {
			busy = false;
		}
	}

	async function messagePlayer() {
		if (!view || !selected) return;
		const room = view.rooms.find((candidate) => candidate.key === `gm:${selected.id}`);
		if (room) {
			await goto(resolve(`/admin/games/${view.game.id}/chat/${room.id}`));
		} else {
			toasts.error('The direct conversation is not available yet.');
		}
	}
</script>

{#if view}
	<PageHeading
		eyebrow="Roster and assignments"
		title="Players"
		description={`${activePlayers.length} active players · ${activePlayers.filter((player) => hasAssignedRole(player.roleKey, view.ruleset.roles)).length} roles assigned`}
		variant="compact"
	>
		{#snippet actions()}
			<RoleVisibilityControl
				gameId={view.game.id}
				rolesVisible={view.game.rolesVisible}
				presentation="heading"
			/>
		{/snippet}
	</PageHeading>

	{#if assignmentsHydrated}
		<div class="assignment-actions">
			<Button loading={busy} onclick={saveAssignments}><Check size={18} /> Save roles</Button>
		</div>

		<div class="player-list" role="list" aria-label="Players">
			{#each view.participants as player (player.id)}
				<article role="listitem" class:inactive={['kicked', 'left'].includes(player.status)}>
					<button class="player-open" type="button" onclick={() => (selectedId = player.id)}>
						<span class="avatar">{playerName(player).slice(0, 1).toUpperCase()}</span>
						<span class="identity">
							<strong><i>Player {player.playerNumber}</i>{playerName(player)}</strong>
							<small
								>{player.gameAlias ? player.displayNameSnapshot : statusLabel(player.status)}</small
							>
						</span>
					</button>
					<label class="role-select">
						<span class="sr-only">Role for {playerName(player)}</span>
						{#if view.game.status === 'lobby'}
							<select
								bind:value={assignments[player.id]}
								disabled={['kicked', 'left'].includes(player.status)}
							>
								<option value="">Unassigned</option>
								{#each view.ruleset.roles as role (role.id)}
									<option value={role.id}>{role.name}</option>
								{/each}
							</select>
						{:else}
							<select value={player.roleKey ?? ''} disabled>
								<option value="">Unassigned</option>
								{#each view.ruleset.roles as role (role.id)}
									<option value={role.id}>{role.name}</option>
								{/each}
							</select>
						{/if}
					</label>
					<div class="player-facts">
						{#if player.outcome !== 'unset'}
							<span>{player.outcome}</span>
						{/if}
						{#if awardedCount(player) > 0}<span><Award size={15} /> Awarded</span>{/if}
					</div>
				</article>
			{/each}
		</div>
	{/if}
{/if}

<Dialog
	open={selected !== null}
	title={selected ? playerName(selected) : 'Player'}
	description={selected ? `Player ${selected.playerNumber} · ${statusLabel(selected.status)}` : ''}
	close={() => (selectedId = '')}
>
	{#if selected && view}
		<div class="player-detail">
			<Button variant="secondary" onclick={messagePlayer}
				><MessageCircle size={18} /> Message player</Button
			>

			<div class="detail-section">
				<h3>Alias</h3>
				<Field
					label="Game alias"
					name="game-alias"
					bind:value={aliasDraft}
					help={`Profile name: ${selected.displayNameSnapshot}`}
				/>
				<Button variant="secondary" loading={busy} onclick={saveAlias}>Save alias</Button>
			</div>

			<div class="detail-section">
				<h3>Outcome</h3>
				<div class="choice-grid" aria-label="Player outcome">
					{#each ['win', 'loss', 'draw', 'unset'] as outcome (outcome)}
						<button
							type="button"
							class:active={selected.outcome === outcome}
							aria-pressed={selected.outcome === outcome}
							onclick={() => setOutcome(outcome as Participant['outcome'])}
						>
							{outcome === 'unset' ? 'Unset' : outcome}
						</button>
					{/each}
				</div>
			</div>

			{#if view.ruleset.achievements.length > 0}
				<div class="detail-section">
					<h3>Achievements</h3>
					<div class="achievement-list">
						{#each view.ruleset.achievements as achievement (achievement.id)}
							{@const awarded = view.awards.some(
								(award) =>
									award.profileId === selected.profileId && award.achievementId === achievement.id
							)}
							<button type="button" class:awarded onclick={() => toggleAchievement(achievement.id)}>
								<Award size={18} />
								<span
									><strong>{achievement.name}</strong><small>{achievement.description}</small></span
								>
								<i>{awarded ? 'Revoke' : 'Award'}</i>
							</button>
						{/each}
					</div>
				</div>
			{/if}

			<div class="detail-section danger-zone">
				<h3>Player status</h3>
				<div class="status-actions">
					{#if selected.status === 'active' && ['running', 'paused'].includes(view.game.status)}
						<Button variant="secondary" onclick={() => participantAction('eliminate')}>
							<UserMinus size={18} /> Eliminate
						</Button>
					{:else if selected.status === 'eliminated'}
						<Button variant="secondary" onclick={() => participantAction('reinstate')}>
							<UserRoundCheck size={18} /> Reinstate
						</Button>
					{/if}
					{#if !['kicked', 'left'].includes(selected.status)}
						<Button variant="danger" onclick={() => (kickConfirmOpen = true)}>
							<ShieldAlert size={18} /> Kick player
						</Button>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</Dialog>

<Dialog
	open={kickConfirmOpen}
	title="Kick player?"
	description={selected ? `${playerName(selected)} will lose access to this game.` : ''}
	close={() => (kickConfirmOpen = false)}
>
	<p>The player remains in game history but cannot rejoin this game.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (kickConfirmOpen = false)}>Cancel</Button>
		<Button variant="danger" loading={busy} onclick={() => participantAction('kick')}
			>Kick player</Button
		>
	{/snippet}
</Dialog>

<style>
	.assignment-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		border-block-end: var(--border-strong);
		padding-block-end: var(--space-3);
	}

	.player-list article {
		display: grid;
		grid-template-columns: minmax(12rem, 1fr) minmax(10rem, 0.65fr) minmax(8rem, auto);
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.player-list article.inactive {
		color: var(--ink-soft);
	}

	.player-open {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		border: 0;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		text-align: start;
	}

	.player-list article.inactive .player-open {
		color: var(--ink-soft);
	}

	.avatar {
		display: grid;
		width: 3rem;
		height: 3rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.identity strong,
	.identity small {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.identity i {
		display: inline-block;
		color: var(--action-dark);
		font-family: var(--font-display);
		font-size: var(--font-size-xs);
		font-style: normal;
		margin-inline-end: var(--space-2);
	}

	.identity small {
		color: var(--ink-soft);
	}

	.role-select select {
		width: 100%;
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		padding: var(--space-2);
	}

	.player-facts {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-1);
	}

	.player-facts span {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		border: 1px solid var(--gold-dark);
		font-size: var(--font-size-sm);
		padding: 0.15rem 0.4rem;
		text-transform: capitalize;
	}

	.player-detail {
		display: grid;
		gap: var(--space-5);
	}

	.detail-section {
		display: grid;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		padding-block-start: var(--space-3);
	}

	.detail-section h3 {
		margin: 0;
	}

	.choice-grid {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: var(--space-1);
	}

	.choice-grid button {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		cursor: pointer;
		text-transform: capitalize;
	}

	.choice-grid button.active {
		border: 2px solid var(--action-dark);
		background: color-mix(in srgb, var(--action) 10%, transparent);
		font-weight: 700;
	}

	.achievement-list {
		display: grid;
	}

	.achievement-list button {
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: var(--space-2);
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: var(--space-2) 0;
		text-align: start;
	}

	.achievement-list button.awarded {
		color: var(--success);
	}

	.achievement-list strong,
	.achievement-list small {
		display: block;
	}

	.achievement-list small {
		color: var(--ink-soft);
	}

	.achievement-list i {
		font-family: var(--font-display);
		font-size: var(--font-size-xs);
		font-style: normal;
		font-weight: 700;
	}

	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.danger-zone {
		border-block-start-color: var(--danger);
	}

	@media (max-width: 47.99rem) {
		.assignment-actions {
			display: grid;
			grid-template-columns: 1fr 1fr;
		}

		.player-list article {
			grid-template-columns: minmax(0, 1fr) auto;
		}

		.role-select {
			grid-column: 1 / -1;
			grid-row: 2;
		}

		.player-facts {
			grid-column: 2;
			grid-row: 1;
			align-self: center;
			flex-direction: column;
		}

		.choice-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
