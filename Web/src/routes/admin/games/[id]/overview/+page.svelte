<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		CirclePause,
		CirclePlay,
		Flag,
		Forward,
		Megaphone,
		QrCode,
		ShieldCheck,
		Users
	} from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import type { Game, TimerProjection } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let phaseOpen = $state(false);
	let announcementOpen = $state(false);
	let roleConfirmOpen = $state(false);
	let hideRoleConfirmOpen = $state(false);
	let busy = $state(false);
	let selectedPhase = $state('');
	let startSuggestedTimer = $state(true);
	let announcementMessage = $state('');
	let timer = $state<TimerProjection>({
		status: 'inactive',
		totalMs: 0,
		remainingMs: 0,
		revision: 0,
		serverTime: new Date().toISOString()
	});
	let joinUrl = $state('');

	const view = $derived(gameState.admin);
	const activePlayers = $derived(
		view?.participants.filter((player) => !['kicked', 'left'].includes(player.status)) ?? []
	);
	const assignedPlayers = $derived(activePlayers.filter((player) => player.roleKey));
	const assignmentsReady = $derived(
		activePlayers.length > 0 && assignedPlayers.length === activePlayers.length
	);
	const currentPhase = $derived(
		view?.ruleset.phases.find((phase) => phase.id === view?.game.phaseKey)
	);
	const selectedPhaseDefinition = $derived(
		view?.ruleset.phases.find((phase) => phase.id === selectedPhase)
	);

	$effect(() => {
		if (view?.timer && view.timer.revision >= timer.revision) timer = view.timer;
		selectedPhase ||= view?.ruleset.phases[0]?.id ?? '';
	});

	$effect(() => {
		if (view?.game.joiningOpen && !joinUrl) {
			void api<{ joinUrl: string }>('/setup/status').then((status) => (joinUrl = status.joinUrl));
		}
	});

	async function gameCommand(path: string, success: string) {
		if (!view) return;
		busy = true;
		try {
			await api<Game>(`/games/${view.game.id}/${path}`, { method: 'POST', ...jsonBody({}) });
			await gameState.refreshAdmin(view.game.id);
			toasts.success(success);
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The game could not be updated.');
		} finally {
			busy = false;
		}
	}

	async function changePhase() {
		if (!view || !selectedPhaseDefinition) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/phase`, {
				method: 'POST',
				...jsonBody({ phaseKey: selectedPhase })
			});
			phaseOpen = false;
			await gameState.refreshAdmin(view.game.id);
			toasts.success(`Phase changed to ${selectedPhaseDefinition.name}.`);
			if (startSuggestedTimer && selectedPhaseDefinition.suggestedDurationSeconds) {
				try {
					timer = await api<TimerProjection>(`/games/${view.game.id}/timer/start`, {
						method: 'POST',
						...jsonBody({ durationMs: selectedPhaseDefinition.suggestedDurationSeconds * 1000 })
					});
				} catch (caught) {
					toasts.error(
						`Phase changed, but the timer did not start. ${caught instanceof Error ? caught.message : ''}`.trim(),
						{
							actionLabel: 'Retry timer',
							action: () => void retrySuggestedTimer(),
							persistent: true
						}
					);
				}
			}
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The phase could not be changed.');
		} finally {
			busy = false;
		}
	}

	async function retrySuggestedTimer() {
		if (!view || !selectedPhaseDefinition?.suggestedDurationSeconds) return;
		timer = await api<TimerProjection>(`/games/${view.game.id}/timer/start`, {
			method: 'POST',
			...jsonBody({ durationMs: selectedPhaseDefinition.suggestedDurationSeconds * 1000 })
		});
		toasts.success('Timer started.');
	}

	async function sendAnnouncement(event: SubmitEvent) {
		event.preventDefault();
		if (!view) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/announcements`, {
				method: 'POST',
				...jsonBody({ content: announcementMessage, audience: 'all', targetId: '' })
			});
			announcementMessage = '';
			announcementOpen = false;
			toasts.success('Announcement sent.');
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'The announcement could not be sent.'
			);
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

	async function startCompletion() {
		if (!view) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/completion/start`, { method: 'POST', ...jsonBody({}) });
			await goto(resolve(`/admin/games/${view.game.id}/finish/outcomes`));
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'The completion flow could not be started.'
			);
		} finally {
			busy = false;
		}
	}
</script>

{#if view}
	<div class="overview-grid">
		<section class="state-column">
			<div class="state-heading">
				<div>
					<p class="eyebrow">Current game state</p>
					<h1>
						{currentPhase?.name ?? (view.game.status === 'lobby' ? 'Lobby' : 'No phase selected')}
					</h1>
					<p>
						{#if currentPhase}
							Round {view.game.roundNumber || 1} · {currentPhase.description}
						{:else if view.game.status === 'draft'}
							Open the lobby when you are ready for players to join.
						{:else}
							Choose a phase when play begins.
						{/if}
					</p>
				</div>
				<span class="lifecycle">{view.game.status}</span>
			</div>

			{#if ['running', 'paused'].includes(view.game.status)}
				<Panel variant="focal">
					<TimerControl gameId={view.game.id} {timer} onchange={(updated) => (timer = updated)} />
				</Panel>
			{/if}

			<div class="primary-actions">
				{#if view.game.status === 'draft'}
					<Button loading={busy} onclick={() => gameCommand('open-lobby', 'Lobby opened.')}>
						<Users size={19} /> Open lobby
					</Button>
				{:else if view.game.status === 'lobby'}
					<Button loading={busy} onclick={() => gameCommand('start', 'Game started.')}>
						<CirclePlay size={19} /> Start game
					</Button>
					<Button
						variant="ghost"
						disabled={busy}
						onclick={() => gameCommand('close-joining', 'Lobby closed.')}
					>
						Close lobby
					</Button>
				{:else if view.game.status === 'running'}
					<Button loading={busy} onclick={() => gameCommand('pause', 'Game paused.')}>
						<CirclePause size={19} /> Pause game
					</Button>
					<Button variant="secondary" onclick={() => (phaseOpen = true)}
						><Forward size={19} /> Change phase</Button
					>
					<Button variant="secondary" onclick={() => (announcementOpen = true)}
						><Megaphone size={19} /> Announce</Button
					>
					<Button variant="ghost" onclick={startCompletion}><Flag size={19} /> End game</Button>
				{:else if view.game.status === 'paused'}
					<Button loading={busy} onclick={() => gameCommand('resume', 'Game resumed.')}>
						<CirclePlay size={19} /> Resume game
					</Button>
					<Button variant="secondary" onclick={() => (phaseOpen = true)}
						><Forward size={19} /> Change phase</Button
					>
					<Button variant="secondary" onclick={() => (announcementOpen = true)}
						><Megaphone size={19} /> Announce</Button
					>
					<Button variant="ghost" onclick={startCompletion}><Flag size={19} /> End game</Button>
				{:else if view.game.status === 'review'}
					<a class="primary-link" href={resolve(`/admin/games/${view.game.id}/finish/outcomes`)}>
						Continue finishing
					</a>
				{/if}
			</div>

			{#if view.game.joiningOpen}
				<Panel
					title="Lobby invitation"
					description="Use this after the current game controls are handled."
				>
					<div class="invitation">
						<img src="/api/app/v1/setup/join-qr" alt="QR code for the player join page" />
						<div>
							<QrCode size={24} aria-hidden="true" />
							<p>Players scan the code while connected to the same private network.</p>
							{#if joinUrl}<code>{joinUrl}</code>{/if}
						</div>
					</div>
				</Panel>
			{/if}
		</section>

		<aside>
			<Panel title="Readiness">
				<ul class="readiness">
					<li>
						<Users size={20} />
						<div>
							<strong>{activePlayers.length} players</strong><span
								>{assignedPlayers.length} assigned roles</span
							>
						</div>
						<a href={resolve(`/admin/games/${view.game.id}/players`)}>Open</a>
					</li>
					<li>
						<ShieldCheck size={20} />
						<div>
							<strong>{view.game.rolesVisible ? 'Roles available' : 'Roles hidden'}</strong>
							<span
								>{assignmentsReady
									? 'Assignments are ready'
									: 'Assign every active player first'}</span
							>
						</div>
						{#if view.game.rolesVisible}
							<button type="button" onclick={() => (hideRoleConfirmOpen = true)}>Hide</button>
						{:else}
							<button
								type="button"
								disabled={!assignmentsReady}
								onclick={() => (roleConfirmOpen = true)}>Make available</button
							>
						{/if}
					</li>
				</ul>
			</Panel>

			<Panel title="Announcement">
				<p>Send an important update separately from chat.</p>
				<Button variant="secondary" onclick={() => (announcementOpen = true)}>
					<Megaphone size={18} /> New announcement
				</Button>
			</Panel>
		</aside>
	</div>
{/if}

<Dialog
	open={phaseOpen}
	title="Change phase"
	description="Choose the next phase and optionally start its suggested timer."
	close={() => (phaseOpen = false)}
>
	<div class="dialog-stack">
		<label>
			<span>Phase</span>
			<select bind:value={selectedPhase}>
				{#each view?.ruleset.phases ?? [] as phase (phase.id)}
					<option value={phase.id}>{phase.name}</option>
				{/each}
			</select>
		</label>
		{#if selectedPhaseDefinition?.description}<p>{selectedPhaseDefinition.description}</p>{/if}
		{#if selectedPhaseDefinition?.suggestedDurationSeconds}
			<label class="check">
				<input type="checkbox" bind:checked={startSuggestedTimer} />
				Start the {Math.round(selectedPhaseDefinition.suggestedDurationSeconds / 60)} minute suggested
				timer
			</label>
		{/if}
	</div>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (phaseOpen = false)}>Cancel</Button>
		<Button loading={busy} onclick={changePhase}>Change phase</Button>
	{/snippet}
</Dialog>

<Dialog open={announcementOpen} title="New announcement" close={() => (announcementOpen = false)}>
	<form id="announcement-form" onsubmit={sendAnnouncement}>
		<Field
			label="Announcement message"
			name="announcement"
			bind:value={announcementMessage}
			multiline
			required
		/>
	</form>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (announcementOpen = false)}>Cancel</Button>
		<Button
			loading={busy}
			disabled={!announcementMessage.trim()}
			onclick={() =>
				(document.getElementById('announcement-form') as HTMLFormElement)?.requestSubmit()}
		>
			Send announcement
		</Button>
	{/snippet}
</Dialog>

<Dialog
	open={roleConfirmOpen}
	title="Make roles available?"
	description="Players will be able to open and reveal their assigned roles."
	close={() => (roleConfirmOpen = false)}
>
	<p>Each player still chooses when to reveal their private role screen.</p>
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

<style>
	.overview-grid {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(18rem, 0.36fr);
		gap: var(--space-6);
	}

	.state-column {
		display: grid;
		align-content: start;
		gap: var(--space-5);
	}

	.state-heading {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
	}

	.state-heading h1,
	.state-heading p {
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

	.lifecycle {
		border: 1px solid var(--gold-dark);
		background: var(--ink);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		padding: var(--space-1) var(--space-2);
		text-transform: uppercase;
	}

	.primary-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.primary-link {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		border: 1px solid var(--crimson-dark);
		background: var(--crimson);
		color: var(--paper-light);
		font-family: var(--font-display);
		font-weight: 700;
		padding: var(--space-2) var(--space-4);
		text-decoration: none;
	}

	aside {
		display: grid;
		align-content: start;
		gap: var(--space-5);
	}

	.readiness {
		display: grid;
		gap: 0;
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.readiness li {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-2);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.readiness li:last-child {
		border: 0;
	}

	.readiness strong,
	.readiness span {
		display: block;
	}

	.readiness span {
		color: var(--ink-soft);
		font-size: 0.82rem;
	}

	.readiness a,
	.readiness button {
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
	}

	.readiness button:disabled {
		color: var(--disabled);
		cursor: not-allowed;
	}

	.invitation {
		display: grid;
		grid-template-columns: 8rem 1fr;
		align-items: center;
		gap: var(--space-4);
	}

	.invitation img {
		width: 100%;
		border: var(--border-subtle);
		background: white;
		padding: var(--space-1);
	}

	code {
		overflow-wrap: anywhere;
	}

	.dialog-stack,
	form {
		display: grid;
		gap: var(--space-3);
	}

	.dialog-stack > label:not(.check) {
		display: grid;
		gap: var(--space-1);
	}

	.dialog-stack > label > span {
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		text-transform: uppercase;
	}

	select {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-2);
	}

	.check {
		display: flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
	}

	@media (max-width: 63.99rem) {
		.overview-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 47.99rem) {
		.overview-grid {
			gap: var(--space-4);
		}

		.state-column {
			gap: var(--space-4);
		}

		.state-heading h1 {
			font-size: 1.65rem;
		}

		.primary-actions {
			display: grid;
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.primary-actions :global(button:first-child),
		.primary-actions .primary-link {
			grid-column: 1 / -1;
		}

		.invitation {
			grid-template-columns: 6rem 1fr;
		}
	}
</style>
