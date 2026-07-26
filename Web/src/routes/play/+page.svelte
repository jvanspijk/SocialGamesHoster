<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { BookOpen, MessageCircle, Scroll, UserRound, Volume2, VolumeX } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import ChatPanel from '$lib/components/ChatPanel.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import ProtectedMedia from '$lib/components/ProtectedMedia.svelte';
	import RoleCard from '$lib/components/RoleCard.svelte';
	import TimerDisplay from '$lib/components/TimerDisplay.svelte';
	import { api, AppApiError, fetchBlob, jsonBody, pb } from '$lib/api/client';
	import type {
		AppErrorBody,
		ChatMessage,
		Game,
		PlayerGameView,
		Profile,
		RealtimeEnvelope,
		Room
	} from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';

	type Tab = 'game' | 'chat' | 'profile' | 'history';
	type AchievementView = {
		id: string;
		title: string;
		description: string;
		points: number;
		awardedAt: string;
	};
	type HistoryGame = {
		id: string;
		name: string;
		rulesetName: string;
		roleName: string;
		outcome: string;
		endedAt?: string;
		achievements: AchievementView[];
	};
	type History = {
		profile: Profile;
		games: HistoryGame[];
		statistics: { achievementCount: number; achievementPoints: number };
	};
	type PartyProfile = Profile & {
		statistics: Record<string, number>;
		achievements: AchievementView[];
		winRate?: number;
	};

	let tab = $state<Tab>('game');
	let profile = $state<Profile | null>(null);
	let profileForm = $state({ displayName: '', bio: '', accent: '' });
	let history = $state<History | null>(null);
	let room = $state<Room | null>(null);
	let error = $state<AppErrorBody | null>(null);
	let busy = $state(true);
	let soundEnabled = $state(false);
	let avatarFile = $state<File | null>(null);
	let unread = $state<Record<string, number>>({});
	let roomSubscriptions: Array<() => void> = [];
	let pageSubscriptions: Array<() => void> = [];
	let roomSubscriptionGeneration = 0;
	let audioContext: AudioContext | null = null;
	let cueNotice = $state('');
	let partyProfile = $state<PartyProfile | null>(null);
	let dmTarget = $state('');

	onMount(() => {
		soundEnabled = localStorage.getItem('sgh.sound-enabled') === 'true';
		try {
			unread = JSON.parse(localStorage.getItem('sgh.room-unread') ?? '{}') as Record<
				string,
				number
			>;
		} catch {
			unread = {};
		}
		void initialize();
		return () => {
			for (const unsubscribe of roomSubscriptions) unsubscribe();
			for (const unsubscribe of pageSubscriptions) unsubscribe();
		};
	});

	$effect(() => {
		const rooms = gameState.player?.rooms ?? [];
		const signature = rooms.map((candidate) => candidate.id).join(',');
		if (signature || gameState.player) void subscribeRooms(rooms);
	});

	$effect(() => {
		const rooms = gameState.player?.rooms ?? [];
		if (room) room = rooms.find((candidate) => candidate.id === room?.id) ?? rooms[0] ?? null;
	});

	async function initialize() {
		if (!auth.isPlayer) {
			await goto(resolve('/'));
			return;
		}
		try {
			profile = await api<Profile>('/profiles/me');
			profileForm = {
				displayName: profile.displayName,
				bio: profile.bio,
				accent: profile.accent || 'crimson'
			};
			let view: PlayerGameView;
			try {
				view = await gameState.refreshPlayer();
			} catch {
				const live = await api<Game>('/games/live');
				if (live.status === 'lobby' && live.joiningOpen) {
					await api(`/games/${live.id}/join`, { method: 'POST', ...jsonBody({}) });
					view = await gameState.refreshPlayer();
				} else {
					throw new Error('not joined');
				}
			}
			room = view.rooms[0] ?? null;
			pageSubscriptions.push(
				await gameState.subscribe(`game:${view.game.id}:public`, () => gameState.refreshPlayer())
			);
			pageSubscriptions.push(
				await gameState.subscribe(`participant:${view.participant.id}:private`, () =>
					gameState.refreshPlayer()
				)
			);
			const audioUnsubscribe = await pb.realtime.subscribe(`game:${view.game.id}:public`, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<{
					name: string;
					preview: string;
				}>;
				if (event.kind === 'audio.cue') void receiveCue(event.payload);
			});
			pageSubscriptions.push(audioUnsubscribe);
		} catch (caught) {
			error =
				caught instanceof AppApiError
					? caught.body
					: { code: 'game.unavailable', message: 'No game is available for this profile yet.' };
		} finally {
			busy = false;
		}
	}

	async function subscribeRooms(rooms: Room[]) {
		const generation = ++roomSubscriptionGeneration;
		for (const unsubscribe of roomSubscriptions) unsubscribe();
		roomSubscriptions = [];
		const nextSubscriptions: Array<() => void> = [];
		for (const candidate of rooms) {
			const unsubscribe = await pb.realtime.subscribe(`room:${candidate.id}`, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<ChatMessage>;
				if (
					(event.kind === 'chat.message_created' || event.kind === 'chat.announcement') &&
					room?.id !== candidate.id
				) {
					unread = { ...unread, [candidate.id]: (unread[candidate.id] ?? 0) + 1 };
					localStorage.setItem('sgh.room-unread', JSON.stringify(unread));
				}
			});
			if (generation !== roomSubscriptionGeneration) {
				unsubscribe();
				for (const cancel of nextSubscriptions) cancel();
				return;
			}
			nextSubscriptions.push(unsubscribe);
		}
		roomSubscriptions = nextSubscriptions;
	}

	function selectRoom(candidate: Room) {
		room = candidate;
		unread = { ...unread, [candidate.id]: 0 };
		localStorage.setItem('sgh.room-unread', JSON.stringify(unread));
	}

	async function viewPartyProfile(profileId: string) {
		try {
			partyProfile = await api<PartyProfile>(`/profiles/${profileId}/summary`);
		} catch (caught) {
			error =
				caught instanceof AppApiError
					? caught.body
					: { code: 'profile.unavailable', message: 'That party profile could not be opened.' };
		}
	}

	async function createPlayerDM() {
		if (!gameState.player || !dmTarget) return;
		try {
			const created = await api<Room>(`/games/${gameState.player.game.id}/rooms/player-dm`, {
				method: 'POST',
				...jsonBody({ participantId: dmTarget })
			});
			await gameState.refreshPlayer();
			room = created;
			tab = 'chat';
		} catch (caught) {
			error =
				caught instanceof AppApiError
					? caught.body
					: { code: 'chat.dm_failed', message: 'The private room could not be created.' };
		}
	}

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();
		error = null;
		try {
			profile = await api<Profile>('/profiles/me', {
				method: 'PATCH',
				...jsonBody(profileForm)
			});
		} catch (caught) {
			error =
				caught instanceof AppApiError
					? caught.body
					: { code: 'profile.failed', message: 'The profile could not be saved.' };
		}
	}

	async function uploadAvatar() {
		if (!avatarFile) return;
		const form = new FormData();
		try {
			form.append('file', await resizeAvatar(avatarFile), 'avatar.webp');
			profile = await api<Profile>('/profiles/me/avatar', { method: 'POST', body: form });
			avatarFile = null;
		} catch (caught) {
			error =
				caught instanceof AppApiError
					? caught.body
					: { code: 'profile.avatar_failed', message: 'The profile image could not be saved.' };
		}
	}

	async function resizeAvatar(file: File): Promise<Blob> {
		const sourceUrl = URL.createObjectURL(file);
		try {
			const image = new Image();
			image.src = sourceUrl;
			await image.decode();
			const side = Math.min(image.naturalWidth, image.naturalHeight);
			const left = (image.naturalWidth - side) / 2;
			const top = (image.naturalHeight - side) / 2;
			const canvas = document.createElement('canvas');
			canvas.width = Math.min(512, side);
			canvas.height = Math.min(512, side);
			const context = canvas.getContext('2d');
			if (!context) throw new Error('Image resizing is unavailable.');
			context.drawImage(image, left, top, side, side, 0, 0, canvas.width, canvas.height);
			const result = await new Promise<Blob | null>((resolveBlob) =>
				canvas.toBlob(resolveBlob, 'image/webp', 0.86)
			);
			if (!result || result.size > 1_048_576)
				throw new Error('The resized profile image is too large.');
			return result;
		} finally {
			URL.revokeObjectURL(sourceUrl);
		}
	}

	async function loadHistory() {
		tab = 'history';
		if (history) return;
		try {
			history = await api<History>('/profiles/me/history');
		} catch (caught) {
			error =
				caught instanceof AppApiError
					? caught.body
					: { code: 'history.failed', message: 'History could not be loaded.' };
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
</script>

<div class="player-shell stack">
	<header class="player-heading">
		<div>
			<p class="ornament">Player ledger</p>
			<h1>{gameState.player?.game.name ?? 'The table is quiet'}</h1>
			{#if gameState.player}
				<p class="muted">
					{gameState.player.ruleset.name} · Round {gameState.player.game.roundNumber || '—'} · {gameState
						.player.game.status}
				</p>
			{/if}
		</div>
		<div class="heading-actions">
			{#if gameState.player}<TimerDisplay gameId={gameState.player.game.id} compact />{/if}
			<Button variant="ghost" onclick={toggleSound}>
				{#if soundEnabled}<Volume2 size={18} /> Sound on{:else}<VolumeX size={18} /> Enable sound{/if}
			</Button>
		</div>
	</header>

	<nav class="tabs" aria-label="Player pages">
		<button class:active={tab === 'game'} onclick={() => (tab = 'game')}
			><Scroll size={18} /> Game</button
		>
		<button class:active={tab === 'chat'} onclick={() => (tab = 'chat')}
			><MessageCircle size={18} /> Rooms</button
		>
		<button class:active={tab === 'profile'} onclick={() => (tab = 'profile')}
			><UserRound size={18} /> Profile</button
		>
		<button class:active={tab === 'history'} onclick={loadHistory}
			><BookOpen size={18} /> History</button
		>
	</nav>

	<ErrorNotice {error} />
	{#if cueNotice}
		<div class="cue-notice" role="status">
			<Volume2 size={18} aria-hidden="true" />
			<span>Sound cue: {cueNotice}</span>
		</div>
	{/if}

	{#if busy}
		<div class="card"><p>Reading the game ledger…</p></div>
	{:else if tab === 'game' && gameState.player}
		<div class="split">
			<RoleCard
				role={gameState.player.role}
				knowledge={gameState.player.knowledge}
				imageUrl={gameState.player.assets.find(
					(asset) => asset.assetKey === gameState.player?.role?.imageAssetKey
				)?.preview}
			/>
			<aside class="card stack">
				<h2>At the table</h2>
				<dl>
					<div>
						<dt>Seat</dt>
						<dd>{gameState.player.participant.seatNumber}</dd>
					</div>
					<div>
						<dt>Name</dt>
						<dd>
							{gameState.player.participant.gameAlias || gameState.player.participant.displayName}
						</dd>
					</div>
					<div>
						<dt>State</dt>
						<dd>{gameState.player.participant.status}</dd>
					</div>
					<div>
						<dt>Phase</dt>
						<dd>{gameState.player.game.phaseKey || 'Not set'}</dd>
					</div>
				</dl>
				<p class="muted small">{gameState.player.ruleset.description}</p>
				{#if gameState.player.announcements.length}
					<h3>Recent announcements</h3>
					<ul class="announcements">
						{#each gameState.player.announcements as message (message.id)}
							<li>
								<time>{new Date(message.createdAt).toLocaleTimeString()}</time>
								<span>{message.deleted ? 'Announcement removed' : message.content}</span>
							</li>
						{/each}
					</ul>
				{/if}
				<h3>Party</h3>
				<div class="party-list">
					{#each gameState.player.party as member (member.id)}
						<button onclick={() => viewPartyProfile(member.profileId)}>
							<span>{member.seatNumber}. {member.gameAlias || member.displayName}</span>
							<small>{member.status}</small>
						</button>
					{/each}
				</div>
			</aside>
		</div>
		{#if partyProfile}
			<section class="card party-profile stack">
				<div class="party-profile-head">
					<div>
						<h2>{partyProfile.displayName}</h2>
						<p>{partyProfile.bio || 'No biography yet.'}</p>
					</div>
					<Button variant="ghost" onclick={() => (partyProfile = null)}>Close</Button>
				</div>
				<dl>
					{#each Object.entries(partyProfile.statistics) as [label, value] (label)}
						<div>
							<dt>{label.replace(/([A-Z])/g, ' $1')}</dt>
							<dd>{value}</dd>
						</div>
					{/each}
				</dl>
				{#if partyProfile.achievements.length}
					<h3>Achievements</h3>
					<ul>
						{#each partyProfile.achievements as achievement (achievement.id)}
							<li>
								<strong>{achievement.title}</strong> — {achievement.description}
								<small>{achievement.points ?? 0} points</small>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		{/if}
	{:else if tab === 'chat' && gameState.player}
		<div class="room-layout">
			<aside class="room-list card">
				<h2>Rooms</h2>
				<p class="moderation-note">Game masters can read and moderate every room.</p>
				{#if ['running', 'paused'].includes(gameState.player.game.status)}
					<div class="new-dm">
						<select aria-label="Player for a private room" bind:value={dmTarget}>
							<option value="">Private room with…</option>
							{#each gameState.player.party.filter((member) => member.id !== gameState.player?.participant.id) as member (member.id)}
								<option value={member.id}>{member.gameAlias || member.displayName}</option>
							{/each}
						</select>
						<Button variant="secondary" disabled={!dmTarget} onclick={createPlayerDM}>Create</Button
						>
					</div>
				{/if}
				{#each gameState.player.rooms as candidate (candidate.id)}
					<button class:active={room?.id === candidate.id} onclick={() => selectRoom(candidate)}>
						<span>{candidate.label}</span>
						<small>{candidate.kind.replace('_', ' ')}</small>
						{#if unread[candidate.id]}
							<b aria-label={`${unread[candidate.id]} unread messages`}>{unread[candidate.id]}</b>
						{/if}
					</button>
				{/each}
			</aside>
			{#if room}
				{#key room.id}<ChatPanel {room} />{/key}
			{:else}
				<div class="card">No rooms are visible.</div>
			{/if}
		</div>
	{:else if tab === 'profile' && profile}
		<form class="card profile stack" onsubmit={saveProfile}>
			<h2>Your profile</h2>
			<div class="avatar-row">
				{#if profile.avatar}
					<ProtectedMedia src={profile.avatar} kind="image" alt="" />
				{:else}
					<UserRound size={50} aria-hidden="true" />
				{/if}
				<label>
					<span>Profile image</span>
					<input
						type="file"
						accept="image/jpeg,image/png,image/webp"
						onchange={(event) =>
							(avatarFile = (event.currentTarget as HTMLInputElement).files?.[0] ?? null)}
					/>
				</label>
				<Button variant="secondary" disabled={!avatarFile} onclick={uploadAvatar}
					>Upload image</Button
				>
			</div>
			<Field
				label="Display name"
				name="displayName"
				bind:value={profileForm.displayName}
				required
			/>
			<Field
				label="Biography"
				name="bio"
				bind:value={profileForm.bio}
				multiline
				help="Visible to other party members. Maximum 280 characters."
			/>
			<label class="select-label">
				<span>Accent</span>
				<select bind:value={profileForm.accent}>
					<option value="crimson">Crimson</option>
					<option value="forest">Forest</option>
					<option value="navy">Navy</option>
					<option value="gold">Gold</option>
					<option value="plum">Plum</option>
				</select>
			</label>
			<Button type="submit">Save profile</Button>
		</form>
	{:else if tab === 'history'}
		<section class="stack">
			<h2>Private game history</h2>
			{#if history}
				<div class="card points-summary">
					<strong>{history.statistics.achievementPoints} achievement points</strong>
					<span>from {history.statistics.achievementCount} achievements</span>
				</div>
				{#each history.games as game (String(game.id))}
					<article class="card history">
						<div>
							<h3>{game.name}</h3>
							<p>
								{game.rulesetName} · {game.endedAt
									? new Date(game.endedAt).toLocaleDateString()
									: 'In review'}
							</p>
						</div>
						<dl>
							<div>
								<dt>Role</dt>
								<dd>{game.roleName || 'Unassigned'}</dd>
							</div>
							<div>
								<dt>Outcome</dt>
								<dd>{game.outcome}</dd>
							</div>
						</dl>
						{#if game.achievements.length}
							<ul class="history-achievements">
								{#each game.achievements as achievement (achievement.id)}
									<li>
										<strong>{achievement.title}</strong>
										<span>{achievement.points ?? 0} points</span>
									</li>
								{/each}
							</ul>
						{/if}
					</article>
				{:else}
					<p class="card">No completed games yet.</p>
				{/each}
			{:else}
				<p>Loading history…</p>
			{/if}
		</section>
	{/if}
</div>

<style>
	.player-shell {
		max-width: 64rem;
		margin-inline: auto;
	}

	.player-heading {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.heading-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.tabs {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		border-block: 1px solid #9a7e51;
	}

	.cue-notice {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		border: 1px solid var(--gold-dark);
		background: var(--paper-light);
		padding: 0.65rem 0.8rem;
	}

	.tabs button {
		display: inline-flex;
		min-height: 48px;
		align-items: center;
		justify-content: center;
		gap: 0.35rem;
		border: 0;
		border-bottom: 3px solid transparent;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.tabs button.active {
		border-color: var(--crimson);
		color: var(--crimson-dark);
	}

	.room-layout {
		display: grid;
		grid-template-columns: minmax(12rem, 0.3fr) minmax(0, 1fr);
		gap: 1rem;
	}

	.room-list {
		align-self: start;
	}

	.moderation-note {
		color: var(--ink-faint);
		font-size: 0.78rem;
	}

	.new-dm {
		display: grid;
		gap: 0.35rem;
		margin-block-end: 0.5rem;
	}

	.party-list {
		display: grid;
	}

	.announcements {
		display: grid;
		gap: 0.4rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.announcements li {
		display: grid;
		border-inline-start: 3px solid var(--gold-dark);
		padding-inline-start: 0.5rem;
	}

	.announcements time {
		color: var(--ink-faint);
		font-size: 0.72rem;
	}

	.party-list button {
		display: flex;
		min-height: 42px;
		align-items: center;
		justify-content: space-between;
		border: 0;
		border-bottom: 1px dotted #a98d61;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		text-align: start;
	}

	.party-profile-head {
		display: flex;
		align-items: start;
		justify-content: space-between;
		gap: 1rem;
	}

	.room-list button {
		position: relative;
		display: grid;
		width: 100%;
		min-height: 50px;
		border: 0;
		border-inline-start: 3px solid transparent;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: 0.5rem;
		text-align: start;
	}

	.room-list b {
		position: absolute;
		inset-block-start: 0.35rem;
		inset-inline-end: 0.35rem;
		min-width: 1.35rem;
		border-radius: 999px;
		background: var(--crimson);
		color: white;
		font-size: 0.7rem;
		padding: 0.15rem 0.35rem;
		text-align: center;
	}

	.room-list button.active {
		border-color: var(--crimson);
		background: rgb(166 42 42 / 8%);
	}

	.room-list small {
		color: var(--ink-faint);
		text-transform: capitalize;
	}

	dl {
		margin: 0;
	}

	dl div {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		border-bottom: 1px dotted #a98d61;
		padding-block: 0.5rem;
	}

	dt {
		color: var(--ink-soft);
	}

	dd {
		margin: 0;
		font-weight: 700;
		text-transform: capitalize;
	}

	.profile {
		max-width: 38rem;
	}

	.avatar-row {
		display: grid;
		grid-template-columns: 4rem minmax(0, 1fr);
		align-items: center;
		gap: 0.75rem;
	}

	.avatar-row :global(img) {
		width: 4rem;
		height: 4rem;
		border: 1px solid #8d7248;
		border-radius: 50%;
		object-fit: cover;
	}

	.avatar-row :global(button) {
		grid-column: 2;
	}

	.select-label {
		display: grid;
		gap: 0.3rem;
	}

	.select-label span {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	select {
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.6rem;
	}

	.history {
		display: grid;
		grid-template-columns: 1fr minmax(12rem, 0.4fr);
		gap: 1rem;
	}

	@media (max-width: 680px) {
		.player-heading,
		.heading-actions {
			display: grid;
		}

		.tabs button {
			font-size: 0;
		}

		.room-layout,
		.history {
			grid-template-columns: 1fr;
		}

		.room-list {
			display: flex;
			overflow-x: auto;
			gap: 0.3rem;
		}

		.room-list h2 {
			display: none;
		}

		.room-list button {
			min-width: 8rem;
		}
	}
</style>
