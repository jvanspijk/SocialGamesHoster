<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		Award,
		Check,
		Dices,
		Eye,
		EyeOff,
		MessageCircle,
		ShieldAlert,
		UserMinus,
		UserRoundCheck
	} from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Field from '$lib/components/Field.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import type { Participant } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let selectedId = $state('');
	let roleConfirmOpen = $state(false);
	let hideRoleConfirmOpen = $state(false);
	let kickConfirmOpen = $state(false);
	let busy = $state(false);
	let assignments = $state<Record<string, string>>({});

	const view = $derived(gameState.admin);
	const selected = $derived(view?.participants.find((player) => player.id === selectedId) ?? null);
	let aliasDraft = $derived(selected?.gameAlias ?? '');
	const activePlayers = $derived(
		view?.participants.filter((player) => !['kicked', 'left'].includes(player.status)) ?? []
	);
	const assignmentsReady = $derived(
		activePlayers.length > 0 &&
			activePlayers.every((player) => assignments[player.id] || player.roleKey)
	);

	$effect(() => {
		if (view) {
			for (const player of view.participants) {
				if (!(player.id in assignments)) assignments[player.id] = player.roleKey ?? '';
			}
		}
	});

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
			await gameState.refreshAdmin(view.game.id);
			toasts.success('Role assignments saved.');
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'Role assignments could not be saved.'
			);
		} finally {
			busy = false;
		}
	}

	async function randomize() {
		if (!view) return;
		busy = true;
		try {
			await api<{ assignments: Participant[] }>(`/games/${view.game.id}/assignments/randomize`, {
				method: 'POST',
				...jsonBody({ assignments: [] })
			});
			await gameState.refreshAdmin(view.game.id);
			const refreshed = gameState.admin;
			if (refreshed) {
				assignments = Object.fromEntries(
					refreshed.participants.map((player) => [player.id, player.roleKey ?? ''])
				);
			}
			toasts.success('Roles randomized.');
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'Roles could not be randomized.');
		} finally {
			busy = false;
		}
	}

	async function setRoleVisibility(rolesVisible: boolean) {
		if (!view) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/role-visibility`, {
				method: 'PATCH',
				...jsonBody({ rolesVisible })
			});
			await gameState.refreshAdmin(view.game.id);
			roleConfirmOpen = false;
			hideRoleConfirmOpen = false;
			toasts.success(rolesVisible ? 'Roles are available.' : 'Roles are hidden.');
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'Role visibility could not be changed.'
			);
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
			toasts.error(caught instanceof Error ? caught.message : 'The alias could not be saved.');
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
			toasts.error(caught instanceof Error ? caught.message : 'The outcome could not be saved.');
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
					...jsonBody({ profileId: selected.profileId, achievementId, note: '' })
				});
			}
			await gameState.refreshAdmin(view.game.id);
			toasts.success(existing ? 'Achievement revoked.' : 'Achievement awarded.');
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'The achievement could not be updated.'
			);
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
			toasts.error(caught instanceof Error ? caught.message : 'The player could not be updated.');
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
	<header class="page-heading">
		<div>
			<p class="eyebrow">Roster and assignments</p>
			<h1>Players</h1>
			<p>
				{activePlayers.length} active players · {activePlayers.filter((player) => player.roleKey)
					.length} roles assigned
			</p>
		</div>
		<div class="role-visibility">
			<div>
				{#if view.game.rolesVisible}<Eye size={20} />{:else}<EyeOff size={20} />{/if}
				<span>
					<strong>{view.game.rolesVisible ? 'Roles available' : 'Roles hidden'}</strong>
					<small
						>{assignmentsReady ? 'Assignments ready' : 'Assign every active player first'}</small
					>
				</span>
			</div>
			{#if view.game.rolesVisible}
				<Button variant="secondary" onclick={() => (hideRoleConfirmOpen = true)}>Hide roles</Button>
			{:else}
				<Button disabled={!assignmentsReady} onclick={() => (roleConfirmOpen = true)}
					>Make roles available</Button
				>
			{/if}
		</div>
	</header>

	<div class="assignment-actions">
		<Button variant="secondary" loading={busy} onclick={randomize}
			><Dices size={18} /> Randomize roles</Button
		>
		<Button loading={busy} disabled={!assignmentsReady} onclick={saveAssignments}
			><Check size={18} /> Save roles</Button
		>
	</div>

	<div class="player-list" role="list" aria-label="Players">
		{#each view.participants as player (player.id)}
			<article role="listitem" class:inactive={['kicked', 'left'].includes(player.status)}>
				<button class="player-open" type="button" onclick={() => (selectedId = player.id)}>
					<span class="avatar">{playerName(player).slice(0, 1).toUpperCase()}</span>
					<span class="identity">
						<strong><i>Seat {player.seatNumber}</i>{playerName(player)}</strong>
						<small
							>{player.gameAlias ? player.displayNameSnapshot : statusLabel(player.status)}</small
						>
					</span>
				</button>
				<label class="role-select">
					<span class="sr-only">Role for {playerName(player)}</span>
					<select
						bind:value={assignments[player.id]}
						disabled={view.game.status !== 'lobby' || ['kicked', 'left'].includes(player.status)}
					>
						<option value="">Unassigned</option>
						{#each view.ruleset.roles as role (role.id)}
							<option value={role.id}>{role.name}</option>
						{/each}
					</select>
				</label>
				<div class="player-facts">
					<span class:unset={player.outcome === 'unset'}
						>{player.outcome === 'unset' ? 'No outcome' : player.outcome}</span
					>
					{#if awardedCount(player) > 0}<span><Award size={15} /> Awarded</span>{/if}
				</div>
			</article>
		{/each}
	</div>
{/if}

<Dialog
	open={selected !== null}
	title={selected ? playerName(selected) : 'Player'}
	description={selected ? `Seat ${selected.seatNumber} · ${statusLabel(selected.status)}` : ''}
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
	open={roleConfirmOpen}
	title="Make roles available?"
	description="Players will be able to open and reveal their assigned roles."
	close={() => (roleConfirmOpen = false)}
>
	<p>Every active player has a role assignment.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (roleConfirmOpen = false)}>Cancel</Button>
		<Button loading={busy} onclick={() => setRoleVisibility(true)}>Make roles available</Button>
	{/snippet}
</Dialog>

<Dialog
	open={hideRoleConfirmOpen}
	title="Hide roles?"
	description="Players with the Role screen open will lose access immediately."
	close={() => (hideRoleConfirmOpen = false)}
>
	<p>Role and knowledge data will be removed from player screens.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (hideRoleConfirmOpen = false)}>Cancel</Button>
		<Button variant="danger" loading={busy} onclick={() => setRoleVisibility(false)}
			>Hide roles</Button
		>
	{/snippet}
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
	.page-heading {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-4);
		margin-block-end: var(--space-4);
	}

	.page-heading h1,
	.page-heading p {
		margin: 0;
	}

	.eyebrow {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	.role-visibility {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		border: var(--border-subtle);
		background: rgb(255 249 230 / 58%);
		padding: var(--space-2);
	}

	.role-visibility > div {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.role-visibility strong,
	.role-visibility small {
		display: block;
	}

	.role-visibility small {
		color: var(--ink-soft);
	}

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
		opacity: 0.58;
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
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.65rem;
		font-style: normal;
		margin-inline-end: var(--space-2);
		text-transform: uppercase;
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
		font-size: 0.72rem;
		padding: 0.15rem 0.4rem;
		text-transform: capitalize;
	}

	.player-facts span.unset {
		border-color: var(--ink-faint);
		color: var(--ink-soft);
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
		border: 2px solid var(--crimson-dark);
		background: rgb(166 42 42 / 10%);
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
		font-size: 0.66rem;
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

	@media (max-width: 63.99rem) {
		.page-heading {
			align-items: stretch;
			flex-direction: column;
		}

		.role-visibility {
			justify-content: space-between;
		}
	}

	@media (max-width: 47.99rem) {
		.role-visibility {
			align-items: stretch;
			flex-direction: column;
		}

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
