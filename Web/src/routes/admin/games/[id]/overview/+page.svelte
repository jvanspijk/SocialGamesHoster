<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import CirclePlay from '@lucide/svelte/icons/circle-play';
	import Flag from '@lucide/svelte/icons/flag';
	import Forward from '@lucide/svelte/icons/forward';
	import Button from '$lib/components/Button.svelte';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import RecordItem from '$lib/components/RecordItem.svelte';
	import RecordList from '$lib/components/RecordList.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import TimerControl from '$lib/features/games/components/TimerControl.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Game, TimerProjection } from '$lib/api/types';
	import { hasAssignedRole } from '$lib/features/games/roleAssignments';
	import { gameStatusLabel } from '$lib/gamePresentation';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let phaseOpen = $state(false);
	let choicesOpen = $state(false);
	let cancelConfirmOpen = $state(false);
	let pending = $state('');
	let selectedPhase = $state('');
	let startSuggestedTimer = $state(true);
	let phaseError = $state('');
	let timer = $state<TimerProjection>({
		status: 'inactive',
		totalMs: 0,
		remainingMs: 0,
		revision: 0,
		serverTime: new Date().toISOString()
	});
	let joinUrl = $state('');

	const view = $derived(gameState.admin);
	const playing = $derived(!!view && ['running', 'paused'].includes(view.game.status));
	const currentPlayers = $derived(
		view?.participants.filter((player) => !['kicked', 'left'].includes(player.status)) ?? []
	);
	const activePlayers = $derived(currentPlayers.filter((player) => player.status === 'active'));
	const eliminatedPlayers = $derived(
		currentPlayers.filter((player) => player.status === 'eliminated')
	);
	const unassignedPlayers = $derived(
		currentPlayers.filter((player) => !hasAssignedRole(player.roleKey, view?.ruleset.roles ?? []))
			.length
	);
	const validPlayerCount = $derived(
		!!view &&
			currentPlayers.length >= view.ruleset.metadata.minPlayers &&
			currentPlayers.length <= view.ruleset.metadata.maxPlayers
	);
	const readyToStart = $derived(validPlayerCount && unassignedPlayers === 0);
	const phases = $derived([...(view?.ruleset.phases ?? [])].sort((a, b) => a.order - b.order));
	const currentPhase = $derived(phases.find((phase) => phase.id === view?.game.phaseKey));
	const selectedPhaseDefinition = $derived(phases.find((phase) => phase.id === selectedPhase));
	const hasChoices = $derived(
		!!view && (view.abilityProgress.eligiblePlayerCount > 0 || view.abilityResults.length > 0)
	);
	const busy = $derived(pending !== '');

	$effect(() => {
		if (view?.timer && view.timer.revision >= timer.revision) timer = view.timer;
	});
	$effect(() => {
		if (view?.game.joiningOpen && !joinUrl) {
			void api<{ joinUrl: string }>('/setup/status')
				.then((status) => (joinUrl = status.joinUrl))
				.catch(() => {});
		}
	});

	function openPhase() {
		selectedPhase =
			phases.find((phase) => phase.id !== currentPhase?.id)?.id ?? phases[0]?.id ?? '';
		startSuggestedTimer = true;
		phaseError = '';
		phaseOpen = true;
	}

	async function gameCommand(path: string, success: string) {
		if (!view || busy) return;
		const gameId = view.game.id;
		pending = path;
		try {
			await api<Game>(`/games/${gameId}/${path}`, { method: 'POST', ...jsonBody({}) });
			await gameState.refreshAdmin(gameId);
			toasts.success(success);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game could not be updated.'));
		} finally {
			pending = '';
		}
	}

	async function changePhase() {
		if (!view || !selectedPhaseDefinition || busy || !playing) return;
		const phase = selectedPhaseDefinition;
		const gameId = view.game.id;
		pending = 'phase';
		phaseError = '';
		try {
			await api(`/games/${gameId}/phase`, { method: 'POST', ...jsonBody({ phaseKey: phase.id }) });
			phaseOpen = false;
			await gameState.refreshAdmin(gameId);
			toasts.success(`${phase.name} started.`);
			if (startSuggestedTimer && phase.suggestedDurationSeconds) {
				const durationMs = phase.suggestedDurationSeconds * 1000;
				try {
					timer = await api<TimerProjection>(`/games/${gameId}/timer/start`, {
						method: 'POST',
						...jsonBody({ durationMs })
					});
				} catch (caught) {
					toasts.error(
						`Phase changed, but the timer did not start. ${errorMessage(caught, '')}`.trim(),
						{
							actionLabel: 'Retry timer',
							persistent: true,
							action: () =>
								void api<TimerProjection>(`/games/${gameId}/timer/start`, {
									method: 'POST',
									...jsonBody({ durationMs })
								})
									.then((updated) => {
										timer = updated;
										toasts.success('Timer started.');
									})
									.catch((error) =>
										toasts.error(errorMessage(error, 'The timer could not be started.'))
									)
						}
					);
				}
			}
		} catch (caught) {
			phaseError = errorMessage(caught, 'The phase could not be changed.');
		} finally {
			pending = '';
		}
	}

	async function startCompletion() {
		if (!view || busy) return;
		pending = 'finish';
		try {
			await api(`/games/${view.game.id}/completion/start`, { method: 'POST', ...jsonBody({}) });
			await goto(resolve(`/admin/games/${view.game.id}/finish/outcomes`));
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game could not be finished.'));
		} finally {
			pending = '';
		}
	}

	async function cancelGame() {
		if (!view || busy) return;
		pending = 'cancel';
		try {
			await api(`/games/${view.game.id}/cancel`, { method: 'POST', ...jsonBody({}) });
			toasts.success('Game cancelled.');
			await goto(resolve('/admin/games'));
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game could not be cancelled.'));
		} finally {
			pending = '';
		}
	}
</script>

{#if view}
	<div class="actions-page">
		<div class="actions-workspace" class:with-finish={playing}>
			<div class="actions-content">
				{#if playing}
					<Panel variant="focal">
						<div class="live-status">
							<div>
								<strong>{gameStatusLabel(view.game.status)}</strong>
								<p>
									{activePlayers.length} in play · {eliminatedPlayers.length} eliminated{view.game
										.roundNumber
										? ` · Round ${view.game.roundNumber}`
										: ''}
								</p>
							</div>
						</div>
						<div class="game-controls" class:with-context={phases.length > 0 || hasChoices}>
							{#if phases.length > 0 || hasChoices}
								<div class="phase-content">
									{#if phases.length}
										<h2 class="phase-name">{currentPhase?.name ?? 'No phase selected'}</h2>
										{#if currentPhase?.description}<p>{currentPhase.description}</p>{/if}
										<div class="button-row">
											<Button variant="secondary" disabled={busy} onclick={openPhase}
												><Forward size={18} /> Change phase</Button
											>
										</div>
									{/if}
									{#if hasChoices}
										<div class="choice-progress">
											<div>
												<strong>Ability choices</strong>
												{#if view.abilityProgress.eligiblePlayerCount > 0}
													<p>
														{view.abilityProgress.locked
															? view.abilityProgress.finalizedPlayerCount
															: view.abilityProgress.activatedPlayerCount} / {view.abilityProgress
															.eligiblePlayerCount} players · {view.abilityProgress.locked
															? 'Finalized'
															: 'Submitted'}
													</p>
												{/if}
											</div>
											<Button variant="ghost" onclick={() => (choicesOpen = true)}
												>View choices</Button
											>
										</div>
									{/if}
								</div>
							{/if}
							<div class="timer-card">
								<TimerControl
									gameId={view.game.id}
									{timer}
									onchange={(updated) => (timer = updated)}
								/>
							</div>
						</div>
					</Panel>
				{:else if ['draft', 'lobby'].includes(view.game.status)}
					<Panel
						title={view.game.status === 'draft' ? 'Open the lobby' : 'Prepare to play'}
						variant="focal"
					>
						<div class="preparation">
							<div class="stack">
								<p>
									{currentPlayers.length} players joined · {currentPlayers.length -
										unassignedPlayers} roles assigned
								</p>
								{#if view.game.status === 'draft'}
									<p>Open the lobby when you are ready for players to join.</p>
									<Button
										loading={pending === 'open-lobby'}
										disabled={busy}
										onclick={() => gameCommand('open-lobby', 'Lobby opened.')}>Open lobby</Button
									>
								{:else}
									{#if !validPlayerCount}<p>
											Needs {view.ruleset.metadata.minPlayers}–{view.ruleset.metadata.maxPlayers} players.
										</p>{/if}
									{#if unassignedPlayers}<p>
											Assign a role to {unassignedPlayers === 1
												? 'the remaining player'
												: `the remaining ${unassignedPlayers} players`} before starting.
										</p>{/if}
									<div class="button-row">
										<Button
											disabled={busy || !readyToStart}
											loading={pending === 'start'}
											onclick={() => gameCommand('start', 'Game started.')}
											><CirclePlay size={18} /> Start game</Button
										>
										<a href={resolve(`/admin/games/${view.game.id}/players`)}>Assign roles</a>
									</div>
								{/if}
							</div>
							{#if view.game.joiningOpen}
								<div class="invitation">
									<img src="/api/app/v1/setup/join-qr" alt="QR code for the player join page" />
									<div>
										<strong>Scan to join</strong>
										<p>Connect to the same Wi-Fi network.</p>
										{#if joinUrl}<code>{joinUrl}</code>{/if}
									</div>
								</div>
							{/if}
						</div>
					</Panel>
				{:else}
					<Panel
						title={view.game.status === 'review' ? 'Review the game' : 'Game archived'}
						variant="focal"
					>
						<div class="stack">
							<p>
								{view.game.status === 'review'
									? 'Check outcomes and achievements before archiving.'
									: 'The game and its history are read-only.'}
							</p>
							<a
								href={resolve(
									view.game.status === 'review'
										? `/admin/games/${view.game.id}/finish/outcomes`
										: `/admin/games/${view.game.id}/summary`
								)}>{view.game.status === 'review' ? 'Continue finishing' : 'View summary'}</a
							>
						</div>
					</Panel>
				{/if}

				{#if view.game.status === 'lobby'}
					<Panel
						title="Cancel game"
						description="Permanently removes this lobby, its players and chat."
					>
						<Button variant="ghost" disabled={busy} onclick={() => (cancelConfirmOpen = true)}
							>Cancel game</Button
						>
					</Panel>
				{/if}
			</div>
			{#if playing}
				<div class="finish-control">
					<p></p>
					<Button
						variant="ghost"
						disabled={busy}
						loading={pending === 'finish'}
						onclick={startCompletion}><Flag size={18} /> Finish game</Button
					>
				</div>
			{/if}
		</div>
	</div>

	<Sheet
		open={phaseOpen}
		title="Change phase"
		close={() => {
			phaseOpen = false;
		}}
	>
		<div class="stack">
			<ErrorNotice message={phaseError} />
			<SelectField
				label="Phase"
				name="game-phase"
				bind:value={selectedPhase}
				disabled={busy}
				options={phases.map((phase) => ({ value: phase.id, label: phase.name }))}
			/>
			{#if selectedPhaseDefinition?.description}<p>{selectedPhaseDefinition.description}</p>{/if}
			{#if selectedPhase === currentPhase?.id}<p>This restarts the current phase.</p>{/if}
			{#if selectedPhaseDefinition?.startsRound}<p>
					Starts round {view.game.roundNumber + 1}.
				</p>{/if}
			{#if view.abilityProgress.eligiblePlayerCount > 0 && !view.abilityProgress.locked}<p>
					Current ability choices will be finalized.
				</p>{/if}
			{#if selectedPhaseDefinition?.suggestedDurationSeconds}
				<CheckboxField
					name="phase-timer"
					label={`Start the suggested timer (${selectedPhaseDefinition.suggestedDurationSeconds} seconds)`}
					bind:checked={startSuggestedTimer}
					disabled={busy}
					description={timer.status !== 'inactive' ? 'Replaces the current timer.' : ''}
				/>
			{/if}
		</div>
		{#snippet actions()}
			<Button variant="ghost" disabled={busy} onclick={() => (phaseOpen = false)}>Cancel</Button>
			<Button
				loading={pending === 'phase'}
				disabled={busy || !selectedPhaseDefinition || !playing}
				onclick={changePhase}
				>{selectedPhase === currentPhase?.id ? 'Restart' : 'Start'}
				{selectedPhaseDefinition?.name ?? 'phase'}</Button
			>
		{/snippet}
	</Sheet>

	<Sheet open={choicesOpen} title="Ability choices" close={() => (choicesOpen = false)}>
		<div class="stack">
			{#if view.abilityProgress.eligiblePlayerCount > 0 && !view.abilityProgress.locked}<p>
					Individual choices stay hidden until the phase ends or the timer runs out.
				</p>{/if}
			<RecordList
				>{#each view.abilityResults as result (`${result.participantId}-${result.roundNumber}-${result.phaseKey}`)}
					<RecordItem
						><strong>Player {result.playerNumber} · {result.displayName}</strong>
						<p>
							Round {result.roundNumber || 1} · {phases.find(
								(phase) => phase.id === result.phaseKey
							)?.name ?? result.phaseKey}: {result.abilities
								.map((ability) => ability.name)
								.join(', ')}
						</p></RecordItem
					>
				{/each}</RecordList
			>
		</div>
	</Sheet>
{/if}

<Dialog
	open={cancelConfirmOpen}
	title="Cancel game?"
	description={view ? `"${view.game.name}" will be permanently removed.` : ''}
	close={() => {
		cancelConfirmOpen = false;
	}}
>
	<p>The player roster and lobby chat will also be deleted. This action cannot be undone.</p>
	{#snippet actions()}
		<Button variant="ghost" disabled={busy} onclick={() => (cancelConfirmOpen = false)}
			>Keep game</Button
		>
		<Button variant="danger" loading={pending === 'cancel'} onclick={cancelGame}>Cancel game</Button
		>
	{/snippet}
</Dialog>

<style>
	.actions-page {
		max-width: 76rem;
		margin-inline: auto;
		display: grid;
		gap: var(--space-4);
	}
	.actions-workspace,
	.actions-content {
		display: grid;
		gap: var(--space-4);
		min-width: 0;
		align-content: start;
	}
	.actions-content {
		container: actions-content / inline-size;
	}
	.game-controls {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: minmax(0, 1fr);
	}
	.live-status,
	.finish-control {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-3);
	}
	.live-status {
		padding-block-end: var(--space-3);
		margin-block-end: var(--space-3);
		border-block-end: var(--border-subtle);
	}
	.finish-control {
		border-block-start: var(--border-subtle);
		padding-block-start: var(--space-3);
	}
	.phase-content,
	.stack {
		display: grid;
		align-content: start;
		gap: var(--space-3);
		min-width: 0;
	}
	.phase-name {
		margin: 0;
		font-size: var(--font-size-xl);
		overflow-wrap: anywhere;
	}
	p {
		margin: 0;
		color: var(--text-secondary);
		overflow-wrap: anywhere;
	}
	a {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		color: var(--action-dark);
		font-weight: 700;
	}
	.stack > a {
		justify-self: start;
	}
	.timer-card {
		container: timer-card / inline-size;
		width: min(100%, 28rem);
	}
	.button-row {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-3);
	}
	.choice-progress {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		padding-block-start: var(--space-3);
	}
	.preparation {
		display: grid;
		gap: var(--space-5);
	}
	.invitation {
		display: grid;
		grid-template-columns: 6rem minmax(0, 1fr);
		gap: var(--space-3);
		align-items: center;
	}
	.invitation img {
		width: 100%;
	}
	code {
		overflow-wrap: anywhere;
	}
	@media (min-width: 48rem) {
		.preparation {
			grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
			align-items: center;
		}
	}
	@media (min-width: 64rem) {
		.with-finish > .actions-content {
			padding-block-end: var(--space-4);
		}
		.finish-control {
			grid-column: 1;
		}
	}
	@container actions-content (min-width: 48rem) {
		.game-controls.with-context {
			grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
			align-items: start;
		}
	}
</style>
