<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Lock,
		Megaphone,
		MessageCircle,
		Pause,
		Play,
		Shuffle,
		Square,
		TimerReset,
		Trophy,
		Unlock,
		UsersRound,
		Volume2,
		VolumeX
	} from '@lucide/svelte';
	import { api, AppApiError, fetchBlob, jsonBody, pb } from '$lib/api/client';
	import type {
		AdminGameView,
		AppErrorBody,
		Participant,
		RealtimeEnvelope,
		Room,
		TimerProjection
	} from '$lib/api/types';
	import Button from './Button.svelte';
	import ChatPanel from './ChatPanel.svelte';
	import ErrorNotice from './ErrorNotice.svelte';
	import TimerDisplay from './TimerDisplay.svelte';

	let {
		view,
		onrefresh
	}: {
		view: AdminGameView;
		onrefresh: () => Promise<void>;
	} = $props();

	let error = $state<AppErrorBody | null>(null);
	let busy = $state('');
	let durationMinutes = $state(5);
	let announcement = $state('');
	let announcementCue = $state('');
	let cueAudience = $state('all');
	let cueTarget = $state('');
	let assignments = $state<Record<string, string>>({});
	let outcomes = $state<Record<string, Participant['outcome']>>({});
	let awardParticipant = $state('');
	let awardKey = $state('');
	let aliases = $state<Record<string, string>>({});
	let rooms = $state<Room[]>([]);
	let selectedRoom = $state<Room | null>(null);
	let soundEnabled = $state(false);
	let cueNotice = $state('');
	let audioContext: AudioContext | null = null;

	onMount(() => {
		soundEnabled = localStorage.getItem('sgh.sound-enabled') === 'true';
	});

	let roles = $derived(view.ruleset.roles as Array<{ id: string; name: string }>);
	let phases = $derived(view.ruleset.phases);
	let achievements = $derived(
		view.ruleset.achievements as Array<{
			id: string;
			name: string;
			points: number;
			hiddenUntilGameCompleted: boolean;
		}>
	);
	let audioCues = $derived(
		view.ruleset.audioCues as Array<{
			id: string;
			name: string;
			defaultAudience: 'all' | 'team' | 'player' | 'game_masters';
		}>
	);
	let teams = $derived(view.ruleset.teams as Array<{ id: string; name: string }>);

	$effect(() => {
		const nextAssignments: Record<string, string> = {};
		const nextOutcomes: Record<string, Participant['outcome']> = {};
		for (const participant of view.participants) {
			nextAssignments[participant.id] = participant.roleKey ?? '';
			nextOutcomes[participant.id] = participant.outcome;
			aliases[participant.id] ??= participant.gameAlias;
		}
		assignments = nextAssignments;
		outcomes = nextOutcomes;
		awardParticipant ||= view.participants[0]?.id ?? '';
		awardKey ||= achievements[0]?.id ?? '';
	});

	$effect(() => {
		const selected = audioCues.find((cue) => cue.id === announcementCue);
		if (selected) cueAudience = selected.defaultAudience;
		cueTarget = '';
	});

	$effect(() => {
		void loadRooms(view.game.id);
	});

	$effect(() => {
		const topic = `game:${view.game.id}:public`;
		let unsubscribe: (() => void) | undefined;
		let disposed = false;
		void pb.realtime
			.subscribe(topic, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<{ name: string; preview: string }>;
				if (event.kind === 'audio.cue') void receiveCue(event.payload);
			})
			.then((cancel) => {
				if (disposed) cancel();
				else unsubscribe = cancel;
			});
		return () => {
			disposed = true;
			unsubscribe?.();
		};
	});

	async function loadRooms(gameId = view.game.id) {
		try {
			const latest = await api<Room[]>(`/games/${gameId}/rooms`);
			rooms = latest;
			selectedRoom =
				latest.find((candidate) => candidate.id === selectedRoom?.id) ?? latest[0] ?? null;
		} catch (caught) {
			setError(caught);
		}
	}

	async function command(path: string, body: unknown = {}) {
		busy = path;
		error = null;
		try {
			await api(`/games/${view.game.id}/${path}`, { method: 'POST', ...jsonBody(body) });
			await onrefresh();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function closeLobby() {
		if (
			!window.confirm(
				'Cancel this lobby? The join code, current roster, and lobby chat will be removed. You can then delete the draft from Games.'
			)
		)
			return;
		await command('close-joining');
	}

	async function randomize() {
		await command('assignments/randomize', { assignments: [] });
	}

	async function saveAssignments() {
		busy = 'assignments';
		try {
			await api(`/games/${view.game.id}/assignments`, {
				method: 'PUT',
				...jsonBody({
					assignments: view.participants
						.filter(
							(participant) => participant.status !== 'kicked' && participant.status !== 'left'
						)
						.map((participant) => ({
							participantId: participant.id,
							roleId: assignments[participant.id]
						}))
				})
			});
			await onrefresh();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function saveOutcomes() {
		busy = 'outcomes';
		try {
			await api(`/games/${view.game.id}/outcomes`, {
				method: 'PUT',
				...jsonBody({
					outcomes: view.participants.map((participant) => ({
						participantId: participant.id,
						outcome: outcomes[participant.id]
					}))
				})
			});
			await onrefresh();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function timer(path: string, body: unknown = {}) {
		busy = `timer-${path}`;
		try {
			await api<TimerProjection>(`/games/${view.game.id}/timer/${path}`, {
				method: 'POST',
				...jsonBody(body)
			});
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function sendAnnouncement(event: SubmitEvent) {
		event.preventDefault();
		const content = announcement.trim();
		if (!content) return;
		await command('announcements', {
			content,
			cueKey: announcementCue,
			audience: announcementCue ? cueAudience : '',
			targetId: announcementCue ? cueTarget : ''
		});
		announcement = '';
	}

	async function award(event: SubmitEvent) {
		event.preventDefault();
		busy = 'award';
		try {
			await api(`/games/${view.game.id}/achievement-awards`, {
				method: 'POST',
				...jsonBody({ participantId: awardParticipant, achievementId: awardKey })
			});
			await onrefresh();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function revokeAward(awardId: string) {
		if (!window.confirm('Revoke this achievement award?')) return;
		try {
			await api(`/games/${view.game.id}/achievement-awards/${awardId}`, { method: 'DELETE' });
			await onrefresh();
		} catch (caught) {
			setError(caught);
		}
	}

	async function participantCommand(participantId: string, action: string) {
		busy = participantId + action;
		try {
			await api(`/games/${view.game.id}/participants/${participantId}/${action}`, {
				method: 'POST',
				...jsonBody({})
			});
			await onrefresh();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function saveAlias(participant: Participant) {
		busy = participant.id + 'alias';
		try {
			await api(`/games/${view.game.id}/participants/${participant.id}`, {
				method: 'PATCH',
				...jsonBody({ gameAlias: aliases[participant.id] ?? '' })
			});
			await onrefresh();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function setRoomLock(room: Room, locked: boolean) {
		busy = room.id + 'lock';
		try {
			await api(`/rooms/${room.id}/${locked ? 'lock' : 'unlock'}`, {
				method: 'POST',
				...jsonBody({})
			});
			await loadRooms();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	function toggleSound() {
		soundEnabled = !soundEnabled;
		localStorage.setItem('sgh.sound-enabled', String(soundEnabled));
		if (soundEnabled) {
			audioContext ??= new AudioContext();
			void audioContext.resume();
		}
	}

	async function receiveCue(cue: { name: string; preview: string }) {
		cueNotice = cue.name;
		window.setTimeout(() => {
			if (cueNotice === cue.name) cueNotice = '';
		}, 6000);
		if (!soundEnabled) return;
		try {
			audioContext ??= new AudioContext();
			await audioContext.resume();
			const blob = await fetchBlob(cue.preview.replace('/api/app/v1', ''));
			const buffer = await audioContext.decodeAudioData(await blob.arrayBuffer());
			const source = audioContext.createBufferSource();
			source.buffer = buffer;
			source.connect(audioContext.destination);
			source.start();
		} catch {
			cueNotice = `${cue.name} (sound could not play)`;
		}
	}

	function setError(caught: unknown) {
		error =
			caught instanceof AppApiError
				? caught.body
				: { code: 'command.failed', message: 'The command could not be completed.' };
	}
</script>

<section class="stack">
	<div class="game-title">
		<div>
			<p class="ornament">{view.ruleset.metadata.name}</p>
			<h2>{view.game.name}</h2>
			<p class="muted">{view.game.status} · Revision {view.game.revision}</p>
		</div>
		{#if view.game.joinCode}
			<div class="join-code">
				<small>Join code</small>
				<strong>{view.game.joinCode}</strong>
			</div>
		{/if}
		<Button variant="ghost" onclick={toggleSound}>
			{#if soundEnabled}<Volume2 size={17} /> Sound on{:else}<VolumeX size={17} /> Enable sound{/if}
		</Button>
	</div>

	<ErrorNotice {error} />
	{#if cueNotice}<div class="cue-notice" role="status">Sound cue: {cueNotice}</div>{/if}

	<div class="command-bar card">
		{#if view.game.status === 'draft'}
			<Button loading={busy === 'open-lobby'} onclick={() => command('open-lobby')}
				>Open lobby</Button
			>
		{:else if view.game.status === 'lobby'}
			<Button variant="danger" loading={busy === 'close-joining'} onclick={closeLobby}
				>Close lobby</Button
			>
			<Button loading={busy === 'start'} onclick={() => command('start')}
				><Play size={17} /> Start game</Button
			>
		{:else if view.game.status === 'running'}
			<Button variant="secondary" onclick={() => command('pause')}
				><Pause size={17} /> Pause game</Button
			>
			<Button variant="secondary" onclick={() => command('review')}>Begin review</Button>
		{:else if view.game.status === 'paused'}
			<Button onclick={() => command('resume')}><Play size={17} /> Resume</Button>
			<Button variant="secondary" onclick={() => command('review')}>Begin review</Button>
		{:else if view.game.status === 'review'}
			<Button variant="secondary" onclick={() => command('return-to-running')}
				>Return to game</Button
			>
			<Button variant="danger" onclick={() => command('archive', { confirmUnsetOutcomes: true })}
				>Archive</Button
			>
		{/if}
	</div>

	<div class="grid">
		<section class="card stack">
			<div class="section-title">
				<UsersRound size={22} />
				<h3>Roster and roles</h3>
			</div>
			{#each view.participants as participant (participant.id)}
				<div class="participant">
					<div class="identity">
						<strong
							>{participant.seatNumber}. {participant.gameAlias ||
								participant.displayNameSnapshot}</strong
						>
						<small>{participant.status}</small>
						<div class="alias-editor">
							<input
								aria-label={`Game alias for ${participant.displayNameSnapshot}`}
								bind:value={aliases[participant.id]}
								maxlength="32"
								placeholder="Game alias"
								disabled={view.game.status === 'archived'}
							/>
							<button
								disabled={view.game.status === 'archived'}
								onclick={() => saveAlias(participant)}>Save alias</button
							>
						</div>
					</div>
					<select
						aria-label={`Role for ${participant.displayNameSnapshot}`}
						bind:value={assignments[participant.id]}
						disabled={view.game.status !== 'lobby'}
					>
						<option value="">Unassigned</option>
						{#each roles as role (role.id)}<option value={role.id}>{role.name}</option>{/each}
					</select>
					<div class="row-actions">
						{#if participant.status === 'active' && ['running', 'paused'].includes(view.game.status)}
							<button onclick={() => participantCommand(participant.id, 'eliminate')}
								>Eliminate</button
							>
						{:else if participant.status === 'eliminated'}
							<button onclick={() => participantCommand(participant.id, 'reinstate')}
								>Reinstate</button
							>
						{/if}
						{#if participant.status !== 'kicked' && view.game.status !== 'archived'}
							<button class="danger-link" onclick={() => participantCommand(participant.id, 'kick')}
								>Kick</button
							>
						{/if}
					</div>
				</div>
			{/each}
			{#if view.game.status === 'lobby'}
				<div class="button-row">
					<Button variant="secondary" onclick={randomize}><Shuffle size={17} /> Randomize</Button>
					<Button loading={busy === 'assignments'} onclick={saveAssignments}>Save roles</Button>
				</div>
			{/if}
		</section>

		<section class="card stack">
			<div class="section-title">
				<TimerReset size={22} />
				<h3>Phase and timer</h3>
			</div>
			{#if ['running', 'paused'].includes(view.game.status)}
				<div class="phase-grid">
					{#each phases as phase (phase.id)}
						<button
							class:active={view.game.phaseKey === phase.id}
							onclick={() => command('phase', { phaseKey: phase.id })}
						>
							<strong>{phase.name}</strong>
							<small
								>{phase.suggestedDurationSeconds
									? `${phase.suggestedDurationSeconds}s suggested`
									: 'No suggested timer'}</small
							>
						</button>
					{/each}
				</div>
				<TimerDisplay gameId={view.game.id} />
				<label class="duration">
					<span>Minutes</span>
					<input type="number" min="1" max="180" bind:value={durationMinutes} />
				</label>
				<div class="button-row">
					<Button onclick={() => timer('start', { durationMs: durationMinutes * 60_000 })}
						><Play size={16} /> Start</Button
					>
					<Button variant="secondary" onclick={() => timer('pause')}
						><Pause size={16} /> Pause</Button
					>
					<Button variant="secondary" onclick={() => timer('resume')}
						><Play size={16} /> Resume</Button
					>
					<Button variant="secondary" onclick={() => timer('stop')}
						><Square size={16} /> Stop</Button
					>
					<Button variant="ghost" onclick={() => timer('adjust', { deltaMs: 60_000 })}
						>+1 min</Button
					>
				</div>
			{:else}
				<p class="muted">Phase and timer controls appear while the game is running.</p>
			{/if}
		</section>
	</div>

	{#if ['lobby', 'running', 'paused'].includes(view.game.status)}
		<form class="card announcement" onsubmit={sendAnnouncement}>
			<Megaphone size={22} aria-hidden="true" />
			<label class="sr-only" for="announcement">Announcement</label>
			<input
				id="announcement"
				bind:value={announcement}
				maxlength="1000"
				placeholder="Announcement to every player…"
			/>
			{#if audioCues.length}
				<select aria-label="Optional sound cue" bind:value={announcementCue}>
					<option value="">No sound</option>
					{#each audioCues as cue (cue.id)}<option value={cue.id}>{cue.name}</option>{/each}
				</select>
				{#if announcementCue}
					<select aria-label="Sound audience" bind:value={cueAudience}>
						<option value="all">All players</option>
						<option value="team">One team</option>
						<option value="player">One player</option>
						<option value="game_masters">Game masters</option>
					</select>
					{#if cueAudience === 'team'}
						<select aria-label="Sound target team" bind:value={cueTarget} required>
							<option value="">Choose team</option>
							{#each teams as team (team.id)}<option value={team.id}>{team.name}</option>{/each}
						</select>
					{:else if cueAudience === 'player'}
						<select aria-label="Sound target player" bind:value={cueTarget} required>
							<option value="">Choose player</option>
							{#each view.participants as participant (participant.id)}
								<option value={participant.id}>{participant.displayNameSnapshot}</option>
							{/each}
						</select>
					{/if}
				{/if}
			{/if}
			<Button type="submit">Announce</Button>
		</form>
	{/if}

	<section class="card stack">
		<div class="section-title">
			<MessageCircle size={22} />
			<h3>Rooms and moderation</h3>
		</div>
		<p class="muted small">
			Game masters can read and moderate every room. Players are told this in their room view.
		</p>
		<div class="moderation-layout">
			<aside class="moderation-rooms">
				{#each rooms as candidate (candidate.id)}
					<div class:active={selectedRoom?.id === candidate.id}>
						<button class="room-select" onclick={() => (selectedRoom = candidate)}>
							<strong>{candidate.label}</strong>
							<small>{candidate.kind.replace('_', ' ')}</small>
						</button>
						<button
							class="room-lock"
							aria-label={`${candidate.locked ? 'Unlock' : 'Lock'} ${candidate.label}`}
							onclick={() => setRoomLock(candidate, !candidate.locked)}
						>
							{#if candidate.locked}<Unlock size={16} />{:else}<Lock size={16} />{/if}
						</button>
					</div>
				{:else}
					<p class="muted">Rooms appear after a game is created.</p>
				{/each}
			</aside>
			{#if selectedRoom}
				{#key selectedRoom.id}<ChatPanel room={selectedRoom} canModerate />{/key}
			{:else}
				<div class="empty-room">Choose a room to inspect it.</div>
			{/if}
		</div>
	</section>

	{#if ['running', 'paused', 'review'].includes(view.game.status)}
		<div class="grid">
			<section class="card stack">
				<h3>Outcomes</h3>
				{#each view.participants as participant (participant.id)}
					<label class="outcome">
						<span>{participant.displayNameSnapshot}</span>
						<select bind:value={outcomes[participant.id]}>
							<option value="unset">Unset</option>
							<option value="win">Win</option>
							<option value="loss">Loss</option>
							<option value="draw">Draw</option>
						</select>
					</label>
				{/each}
				<Button loading={busy === 'outcomes'} onclick={saveOutcomes}>Save outcomes</Button>
			</section>
			{#if achievements.length}
				<form class="card stack" onsubmit={award}>
					<div class="section-title">
						<Trophy size={22} />
						<h3>Award achievement</h3>
					</div>
					<select aria-label="Player" bind:value={awardParticipant}>
						{#each view.participants as participant (participant.id)}<option value={participant.id}
								>{participant.displayNameSnapshot}</option
							>{/each}
					</select>
					<select aria-label="Achievement" bind:value={awardKey}>
						{#each achievements as achievement (achievement.id)}<option value={achievement.id}
								>{achievement.name} ({achievement.points ?? 0} points{achievement.hiddenUntilGameCompleted
									? ', hidden until game ends'
									: ''})</option
							>{/each}
					</select>
					<Button type="submit" loading={busy === 'award'}>Award</Button>
					{#if view.awards.length}
						<h4>Current awards</h4>
						{#each view.awards as existing (existing.id)}
							<div class="award-row">
								<span>
									<strong>{existing.title}</strong>
									<small
										>{view.participants.find(
											(participant) => participant.profileId === existing.profileId
										)?.displayNameSnapshot ?? 'Player'} · {existing.points ?? 0} points
										{existing.hiddenUntilGameCompleted &&
										!['review', 'archived'].includes(view.game.status)
											? ' · hidden from players'
											: ''}</small
									>
								</span>
								<Button
									variant="danger"
									disabled={view.game.status === 'archived'}
									onclick={() => revokeAward(existing.id)}>Revoke</Button
								>
							</div>
						{/each}
					{/if}
				</form>
			{/if}
		</div>
	{/if}

	{#if view.audit.length}
		<details class="card audit">
			<summary>Recent game-master activity</summary>
			<ol>
				{#each view.audit as entry (entry.id)}
					<li>
						<time>{new Date(entry.createdAt).toLocaleString()}</time>
						<strong>{entry.actorLabel}</strong>
						<span>{entry.action.replaceAll('.', ' ')}</span>
					</li>
				{/each}
			</ol>
		</details>
	{/if}
</section>

<style>
	.game-title,
	.command-bar,
	.button-row,
	.section-title,
	.announcement {
		display: flex;
		align-items: center;
		gap: 0.55rem;
	}

	.game-title {
		justify-content: space-between;
	}

	.join-code {
		display: grid;
		border: 1px solid var(--crimson);
		background: var(--paper-light);
		padding: 0.55rem 0.85rem;
		text-align: center;
	}

	.join-code small {
		font-family: var(--font-display);
		font-size: 0.58rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	.join-code strong {
		font-family: var(--font-display);
		font-size: 1.35rem;
		letter-spacing: 0.18em;
	}

	.cue-notice {
		border: 1px solid var(--gold-dark);
		background: var(--paper-light);
		padding: 0.65rem 0.8rem;
	}

	.command-bar,
	.button-row {
		flex-wrap: wrap;
	}

	.section-title h3 {
		margin: 0;
	}

	.participant {
		display: grid;
		grid-template-columns: minmax(9rem, 1fr) minmax(9rem, 0.6fr) auto;
		align-items: center;
		gap: 0.55rem;
		border-bottom: 1px dotted #a98d61;
		padding-block: 0.45rem;
	}

	.identity {
		display: grid;
	}

	.identity small {
		color: var(--ink-faint);
		text-transform: capitalize;
	}

	.alias-editor {
		display: flex;
		gap: 0.3rem;
		margin-block-start: 0.25rem;
	}

	.alias-editor input {
		min-width: 6rem;
		min-height: 36px;
	}

	.alias-editor button {
		min-height: 36px;
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		text-decoration: underline;
	}

	select,
	input {
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.55rem;
	}

	.row-actions {
		display: flex;
		gap: 0.4rem;
	}

	.row-actions button {
		min-height: 36px;
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		text-decoration: underline;
	}

	.danger-link {
		color: var(--danger) !important;
	}

	.phase-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
		gap: 0.4rem;
	}

	.phase-grid button {
		display: grid;
		min-height: 54px;
		border: 1px solid #a98d61;
		background: var(--paper-light);
		cursor: pointer;
		padding: 0.45rem;
		text-align: start;
	}

	.phase-grid button.active {
		border-color: var(--crimson);
		box-shadow: inset 3px 0 var(--crimson);
	}

	.phase-grid small {
		color: var(--ink-faint);
	}

	.duration,
	.outcome {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.6rem;
	}

	.duration input {
		width: 6rem;
	}

	.announcement {
		display: grid;
		grid-template-columns: auto minmax(12rem, 1fr) repeat(4, minmax(8rem, auto));
	}

	.announcement input {
		width: 100%;
	}

	.moderation-layout {
		display: grid;
		grid-template-columns: minmax(12rem, 0.3fr) minmax(0, 1fr);
		gap: 1rem;
	}

	.moderation-rooms {
		align-self: start;
	}

	.moderation-rooms > div {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		border-inline-start: 3px solid transparent;
	}

	.moderation-rooms > div.active {
		border-color: var(--crimson);
		background: rgb(166 42 42 / 8%);
	}

	.room-select,
	.room-lock {
		border: 0;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
	}

	.room-select {
		display: grid;
		min-height: 50px;
		padding: 0.5rem;
		text-align: start;
	}

	.room-select small {
		color: var(--ink-faint);
		text-transform: capitalize;
	}

	.room-lock {
		min-width: 44px;
	}

	.empty-room {
		display: grid;
		min-height: 20rem;
		place-items: center;
		color: var(--ink-faint);
	}

	.award-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		border-top: 1px dotted #a98d61;
		padding-block: 0.4rem;
	}

	.award-row span {
		display: grid;
	}

	.audit summary {
		cursor: pointer;
		font-family: var(--font-display);
		font-weight: 700;
	}

	.audit li {
		display: grid;
		grid-template-columns: minmax(10rem, auto) minmax(8rem, auto) 1fr;
		gap: 0.5rem;
		padding-block: 0.25rem;
	}

	.audit time {
		color: var(--ink-faint);
	}

	@media (max-width: 620px) {
		.participant {
			grid-template-columns: 1fr;
		}

		.announcement {
			grid-template-columns: 1fr;
		}

		.moderation-layout {
			grid-template-columns: 1fr;
		}

		.announcement > :global(svg) {
			display: none;
		}
	}
</style>
