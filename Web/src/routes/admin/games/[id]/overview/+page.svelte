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
	let announcementAudience = $state<'all' | 'team' | 'player'>('all');
	let announcementTarget = $state('');
	let announcementImage = $state('');
	let announcementImageDescription = $state('');
	let announcementAudio = $state('');
	let announcementAudioAlternative = $state('');
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
	const imageAssets = $derived(view?.assets.filter((asset) => asset.kind === 'image') ?? []);
	const audioAssets = $derived(view?.assets.filter((asset) => asset.kind === 'audio') ?? []);

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
				...jsonBody({
					content: announcementMessage,
					audience: announcementAudience,
					targetId: announcementAudience === 'all' ? '' : announcementTarget,
					imageAssetKey: announcementImage,
					imageDescription: announcementImageDescription,
					audioAssetKey: announcementAudio,
					audioAlternative: announcementAudioAlternative
				})
			});
			announcementMessage = '';
			announcementAudience = 'all';
			announcementTarget = '';
			announcementImage = '';
			announcementImageDescription = '';
			announcementAudio = '';
			announcementAudioAlternative = '';
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
						onclick={() => gameCommand('close-joining', 'Lobby cancelled and reset.')}
					>
						Cancel &amp; reset
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
				{#if view.game.joiningOpen && view.game.status !== 'lobby'}
					<Button
						variant="ghost"
						disabled={busy}
						onclick={() => gameCommand('close-joining', 'Joining closed.')}
					>
						Close joining
					</Button>
				{/if}
			</div>

			{#if view.game.joiningOpen}
				<Panel title="Player invitation" description="Players can join while joining remains open.">
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

			{#if view.abilityProgress.eligiblePlayerCount > 0 || view.abilityResults.length > 0}
				<Panel title="Ability choices">
					{#if view.abilityProgress.eligiblePlayerCount > 0}
						{#if view.abilityProgress.locked}
							<p>
								Choices finalized for {view.abilityProgress.finalizedPlayerCount} of
								{view.abilityProgress.eligiblePlayerCount} eligible players.
							</p>
						{:else}
							<p>
								{view.abilityProgress.activatedPlayerCount} of
								{view.abilityProgress.eligiblePlayerCount} eligible players have activated an ability.
								Individual choices stay hidden until the phase locks.
							</p>
							<progress
								value={view.abilityProgress.activatedPlayerCount}
								max={Math.max(view.abilityProgress.eligiblePlayerCount, 1)}
								aria-label={`${view.abilityProgress.activatedPlayerCount} of ${view.abilityProgress.eligiblePlayerCount} eligible players activated`}
							></progress>
						{/if}
					{/if}
					{#if view.abilityResults.length > 0}
						<ul class="ability-results">
							{#each view.abilityResults as result (`${result.participantId}-${result.roundNumber}-${result.phaseKey}`)}
								<li>
									<strong>Seat {result.seatNumber} · {result.displayName}</strong>
									<span>
										Round {result.roundNumber || 1}, {view.ruleset.phases.find(
											(phase) => phase.id === result.phaseKey
										)?.name ?? result.phaseKey}:
										{result.abilities.map((ability) => ability.name).join(', ')}
									</span>
								</li>
							{/each}
						</ul>
					{/if}
				</Panel>
			{/if}
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
		<label>
			<span>Recipients</span>
			<select bind:value={announcementAudience} onchange={() => (announcementTarget = '')}>
				<option value="all">All players</option>
				<option value="team">Selected team</option>
				<option value="player">Selected player</option>
			</select>
		</label>
		{#if announcementAudience === 'team'}
			<label>
				<span>Team</span>
				<select bind:value={announcementTarget} required>
					<option value="">Choose a team</option>
					{#each view?.ruleset.teams ?? [] as team (team.id)}
						<option value={team.id}>{team.name}</option>
					{/each}
				</select>
			</label>
		{:else if announcementAudience === 'player'}
			<label>
				<span>Player</span>
				<select bind:value={announcementTarget} required>
					<option value="">Choose a player</option>
					{#each activePlayers as player (player.id)}
						<option value={player.id}
							>Seat {player.seatNumber} · {player.displayNameSnapshot}</option
						>
					{/each}
				</select>
			</label>
		{/if}
		<label>
			<span>Image (optional)</span>
			<select
				bind:value={announcementImage}
				onchange={() =>
					(announcementImageDescription =
						view?.ruleset.assetAccessibility?.[announcementImage]?.description ?? '')}
			>
				<option value="">No image</option>
				{#each imageAssets as asset (asset.id)}
					<option value={asset.assetKey}>{asset.assetKey}</option>
				{/each}
			</select>
		</label>
		{#if announcementImage}
			<Field
				label="Image description"
				name="announcement-image-description"
				bind:value={announcementImageDescription}
				multiline
				required
			/>
		{/if}
		<label>
			<span>Audio (optional)</span>
			<select
				bind:value={announcementAudio}
				onchange={() =>
					(announcementAudioAlternative =
						view?.ruleset.assetAccessibility?.[announcementAudio]?.description ?? '')}
			>
				<option value="">No audio</option>
				{#each audioAssets as asset (asset.id)}
					<option value={asset.assetKey}>{asset.assetKey}</option>
				{/each}
			</select>
		</label>
		{#if announcementAudio}
			<Field
				label="Audio alternative"
				name="announcement-audio-alternative"
				bind:value={announcementAudioAlternative}
				multiline
				required
			/>
		{/if}
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

	form > label {
		display: grid;
		gap: var(--space-1);
	}

	form > label > span {
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		text-transform: uppercase;
	}

	.ability-results {
		display: grid;
		gap: var(--space-2);
		list-style: none;
		padding: 0;
	}

	.ability-results li,
	.ability-results span {
		display: grid;
	}

	.ability-results span {
		color: var(--ink-soft);
		font-size: 0.82rem;
	}

	progress {
		width: 100%;
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
