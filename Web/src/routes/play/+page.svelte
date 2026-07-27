<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		BookOpen,
		ChevronLeft,
		LogOut,
		MessageCircle,
		Plus,
		Settings,
		UserRound,
		Volume2,
		VolumeX,
		Wifi
	} from '@lucide/svelte';
	import AttentionCard from '$lib/components/AttentionCard.svelte';
	import BottomTaskNav from '$lib/components/BottomTaskNav.svelte';
	import Button from '$lib/components/Button.svelte';
	import ChatPanel from '$lib/components/ChatPanel.svelte';
	import DisplaySettings from '$lib/components/DisplaySettings.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import ProtectedMedia from '$lib/components/ProtectedMedia.svelte';
	import RoleCard from '$lib/components/RoleCard.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import TimerDisplay from '$lib/components/TimerDisplay.svelte';
	import { api, AppApiError, fetchBlob, jsonBody, pb } from '$lib/api/client';
	import type {
		AppErrorBody,
		ChatMessage,
		Game,
		MessageCursor,
		PlayerGameView,
		Profile,
		RealtimeEnvelope,
		Room
	} from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { cursorIsAfter, readMarkerStorageKey } from '$lib/state/chatReadMarkers';
	import { gameState } from '$lib/state/game.svelte';

	type Task = 'game' | 'role' | 'party' | 'more';
	type MoreView = 'menu' | 'profile' | 'history' | 'display' | 'connection';
	type AchievementView = {
		id: string;
		title: string;
		description: string;
		points: number;
		awardedAt: string;
	};
	type History = {
		profile: Profile;
		games: Array<{
			id: string;
			name: string;
			rulesetName: string;
			roleName: string;
			outcome: string;
			endedAt?: string;
			achievements: AchievementView[];
		}>;
		statistics: { achievementCount: number; achievementPoints: number };
	};
	type PartyProfile = Profile & {
		statistics: Record<string, number>;
		achievements: AchievementView[];
	};

	let activeTask = $state<Task>('game');
	let chatOpen = $state(false);
	let chatView = $state<'list' | 'conversation'>('list');
	let moreView = $state<MoreView>('menu');
	let profile = $state<Profile | null>(null);
	let profileForm = $state({ displayName: '', bio: '', accent: '' });
	let history = $state<History | null>(null);
	let selectedRoom = $state<Room | null>(null);
	let error = $state<AppErrorBody | null>(null);
	let busy = $state(true);
	let acknowledging = $state(false);
	let soundEnabled = $state(false);
	let avatarFile = $state<File | null>(null);
	let readMarkers = $state<Record<string, MessageCursor>>({});
	let observedLatest = $state<Record<string, MessageCursor>>({});
	let roomSubscriptions: Array<() => void> = [];
	let pageSubscriptions: Array<() => void> = [];
	let roomSubscriptionGeneration = 0;
	let audioContext: AudioContext | null = null;
	let cueNotice = $state('');
	let partyProfile = $state<PartyProfile | null>(null);
	let dmTarget = $state('');
	let roleRevealSession = $state(0);

	const activeAttention = $derived(gameState.player?.attentionItems[0] ?? null);
	const unreadRoomIds = $derived.by(() => {
		const result: string[] = [];
		for (const candidate of gameState.player?.rooms ?? []) {
			const latest = observedLatest[candidate.id] ?? candidate.latestMessage;
			if (cursorIsAfter(latest, readMarkers[candidate.id])) result.push(candidate.id);
		}
		return result;
	});
	const hasUnreadChat = $derived(unreadRoomIds.length > 0);
	const sheetOpen = $derived(activeTask !== 'game');

	onMount(() => {
		soundEnabled = localStorage.getItem('sgh.sound-enabled') === 'true';
		void initialize();
		const resume = () => {
			if (!document.hidden && gameState.player) void refreshAndReconcile();
		};
		document.addEventListener('visibilitychange', resume);
		return () => {
			document.removeEventListener('visibilitychange', resume);
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
		if (selectedRoom) {
			selectedRoom =
				rooms.find((candidate) => candidate.id === selectedRoom?.id) ?? rooms[0] ?? null;
		}
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
			loadReadMarkers(view);
			pageSubscriptions.push(
				await gameState.subscribe(`game:${view.game.id}:public`, () => refreshAndReconcile())
			);
			pageSubscriptions.push(
				await gameState.subscribe(`participant:${view.participant.id}:private`, () =>
					refreshAndReconcile()
				)
			);
			const audioUnsubscribe = await pb.realtime.subscribe(`game:${view.game.id}:public`, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<{ name: string; preview: string }>;
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

	async function refreshAndReconcile() {
		try {
			await gameState.refreshPlayer();
		} catch {
			// The existing screen remains useful while the connection store reports recovery.
		}
	}

	function markerStorageKey(view = gameState.player) {
		return view && auth.actor ? readMarkerStorageKey(auth.actor.id, view.game.id) : '';
	}

	function loadReadMarkers(view: PlayerGameView) {
		try {
			readMarkers = JSON.parse(localStorage.getItem(markerStorageKey(view)) ?? '{}') as Record<
				string,
				MessageCursor
			>;
		} catch {
			readMarkers = {};
		}
	}

	function saveReadMarkers() {
		const key = markerStorageKey();
		if (key) localStorage.setItem(key, JSON.stringify(readMarkers));
	}

	async function subscribeRooms(rooms: Room[]) {
		const generation = ++roomSubscriptionGeneration;
		for (const unsubscribe of roomSubscriptions) unsubscribe();
		roomSubscriptions = [];
		const nextSubscriptions: Array<() => void> = [];
		for (const candidate of rooms) {
			const unsubscribe = await pb.realtime.subscribe(`room:${candidate.id}`, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<ChatMessage>;
				if (event.kind !== 'chat.message_created') return;
				const cursor = { createdAt: event.payload.createdAt, id: event.payload.id };
				observedLatest = { ...observedLatest, [candidate.id]: cursor };
				if (chatOpen && chatView === 'conversation' && selectedRoom?.id === candidate.id) {
					readMarkers = { ...readMarkers, [candidate.id]: cursor };
					saveReadMarkers();
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

	function openChat() {
		chatOpen = true;
		chatView = 'list';
	}

	function openRoom(candidate: Room) {
		selectedRoom = candidate;
		chatView = 'conversation';
		const latest = observedLatest[candidate.id] ?? candidate.latestMessage;
		if (latest) {
			readMarkers = { ...readMarkers, [candidate.id]: latest };
			saveReadMarkers();
		}
	}

	function closeChat() {
		chatOpen = false;
		chatView = 'list';
	}

	function selectTask(id: string) {
		activeTask = id as Task;
		if (activeTask === 'role') roleRevealSession++;
		if (activeTask === 'more') moreView = 'menu';
	}

	function closeTaskSheet() {
		activeTask = 'game';
		partyProfile = null;
		moreView = 'menu';
	}

	async function acknowledgeAttention() {
		const view = gameState.player;
		if (!view || !activeAttention || activeAttention.kind !== 'announcement') return;
		acknowledging = true;
		error = null;
		try {
			await api(`/games/${view.game.id}/announcements/${activeAttention.id}/acknowledge`, {
				method: 'POST'
			});
			await gameState.refreshPlayer();
		} catch (caught) {
			setError(caught, 'The announcement could not be acknowledged.');
		} finally {
			acknowledging = false;
		}
	}

	async function viewPartyProfile(profileId: string) {
		try {
			partyProfile = await api<PartyProfile>(`/profiles/${profileId}/summary`);
		} catch (caught) {
			setError(caught, 'That party profile could not be opened.');
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
			openRoom(created);
			dmTarget = '';
		} catch (caught) {
			setError(caught, 'The private channel could not be created.');
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
			setError(caught, 'The profile could not be saved.');
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
			setError(caught, 'The profile image could not be saved.');
		}
	}

	async function resizeAvatar(file: File): Promise<Blob> {
		const sourceUrl = URL.createObjectURL(file);
		try {
			const image = new Image();
			image.src = sourceUrl;
			await image.decode();
			const side = Math.min(image.naturalWidth, image.naturalHeight);
			const canvas = document.createElement('canvas');
			canvas.width = Math.min(512, side);
			canvas.height = Math.min(512, side);
			const context = canvas.getContext('2d');
			if (!context) throw new Error('Image resizing is unavailable.');
			context.drawImage(
				image,
				(image.naturalWidth - side) / 2,
				(image.naturalHeight - side) / 2,
				side,
				side,
				0,
				0,
				canvas.width,
				canvas.height
			);
			const result = await new Promise<Blob | null>((done) =>
				canvas.toBlob(done, 'image/webp', 0.86)
			);
			if (!result || result.size > 1_048_576) throw new Error('The resized image is too large.');
			return result;
		} finally {
			URL.revokeObjectURL(sourceUrl);
		}
	}

	async function loadHistory() {
		moreView = 'history';
		if (history) return;
		try {
			history = await api<History>('/profiles/me/history');
		} catch (caught) {
			setError(caught, 'History could not be loaded.');
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

	async function signOut() {
		try {
			await api('/auth/logout', { method: 'POST' });
		} finally {
			auth.clear();
			await goto(resolve('/'));
		}
	}

	function setError(caught: unknown, fallback: string) {
		error =
			caught instanceof AppApiError ? caught.body : { code: 'request.failed', message: fallback };
	}
</script>

<div class="player-shell">
	<header class="status-bar">
		<div class="game-status">
			<strong>{gameState.player?.game.name ?? 'Live game'}</strong>
			{#if gameState.player}
				<span
					>Round {gameState.player.game.roundNumber || '—'} · {gameState.player.game.status}</span
				>
			{/if}
		</div>
		{#if gameState.player}
			<div class="compact-timer">
				<span>{gameState.player.game.phaseKey || 'Waiting for phase'}</span>
				<TimerDisplay gameId={gameState.player.game.id} compact />
			</div>
		{/if}
		<button class="chat-action" type="button" aria-label="Open chat" onclick={openChat}>
			<MessageCircle size={22} />
			{#if hasUnreadChat}<i aria-label="New messages"></i>{/if}
		</button>
	</header>

	<main class="current-stage">
		<ErrorNotice {error} />
		{#if cueNotice}
			<div class="cue-notice" role="status"><Volume2 size={18} /> {cueNotice}</div>
		{/if}
		{#if busy}
			<div class="state"><p>Loading game…</p></div>
		{:else if !gameState.player}
			<div class="state">
				<h1>No game selected</h1>
				<p>Return to the join screen to continue.</p>
			</div>
		{:else if activeAttention}
			<AttentionCard
				item={activeAttention}
				position={1}
				total={gameState.player.attentionItems.length}
				acknowledge={acknowledgeAttention}
				busy={acknowledging}
			/>
		{:else}
			<section class="phase-stage">
				<p>{gameState.player.game.status === 'lobby' ? 'Lobby' : 'Current phase'}</p>
				<h1>{gameState.player.game.phaseKey || 'Waiting for phase'}</h1>
				{#if gameState.player.game.status === 'lobby'}
					<span>Waiting for the game master to start.</span>
				{:else}
					<span>{gameState.player.ruleset.name}</span>
				{/if}
			</section>
		{/if}
	</main>

	<BottomTaskNav
		items={[
			{ id: 'game', label: 'Game', attention: gameState.player?.attentionItems.length !== 0 },
			{ id: 'role', label: 'Role' },
			{ id: 'party', label: 'Party' },
			{ id: 'more', label: 'More' }
		]}
		current={activeTask}
		select={selectTask}
	/>
</div>

<Sheet
	open={chatOpen}
	title={chatView === 'list' ? 'Chat' : selectedRoom?.label || 'Chat'}
	close={closeChat}
>
	{#if chatView === 'conversation' && selectedRoom}
		<div class="conversation stack">
			<button class="back-link" type="button" onclick={() => (chatView = 'list')}>
				<ChevronLeft size={18} /> Channels
			</button>
			{#key selectedRoom.id}<ChatPanel room={selectedRoom} />{/key}
		</div>
	{:else}
		<div class="channel-list stack">
			<p class="moderation-note">Game masters can read and moderate every channel.</p>
			{#if gameState.player && ['running', 'paused'].includes(gameState.player.game.status)}
				<div class="compose-channel">
					<Plus size={18} aria-hidden="true" />
					<select aria-label="Player for a private channel" bind:value={dmTarget}>
						<option value="">New private channel…</option>
						{#each gameState.player.party.filter((member) => member.id !== gameState.player?.participant.id) as member (member.id)}
							<option value={member.id}>{member.gameAlias || member.displayName}</option>
						{/each}
					</select>
					<Button variant="secondary" disabled={!dmTarget} onclick={createPlayerDM}>Create</Button>
				</div>
			{/if}
			{#each gameState.player?.rooms ?? [] as candidate (candidate.id)}
				<button class="channel" type="button" onclick={() => openRoom(candidate)}>
					<span
						><strong>{candidate.label}</strong><small>{candidate.kind.replace('_', ' ')}</small
						></span
					>
					{#if unreadRoomIds.includes(candidate.id)}<b>New</b>{/if}
				</button>
			{:else}
				<p>No chat channels are available.</p>
			{/each}
		</div>
	{/if}
</Sheet>

<Sheet
	open={sheetOpen}
	title={activeTask === 'role' ? 'Your role' : activeTask === 'party' ? 'Party' : 'More'}
	close={closeTaskSheet}
>
	{#if activeTask === 'role' && gameState.player}
		{#key roleRevealSession}
			<RoleCard
				role={gameState.player.role}
				knowledge={gameState.player.knowledge}
				privacyKey={auth.actor?.id ?? ''}
				imageUrl={gameState.player.assets.find(
					(asset) => asset.assetKey === gameState.player?.role?.imageAssetKey
				)?.preview}
			/>
		{/key}
	{:else if activeTask === 'party' && gameState.player}
		{#if partyProfile}
			<div class="stack">
				<button class="back-link" type="button" onclick={() => (partyProfile = null)}>
					<ChevronLeft size={18} /> Party
				</button>
				<h3>{partyProfile.displayName}</h3>
				<p>{partyProfile.bio || 'No biography yet.'}</p>
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
							<li><strong>{achievement.title}</strong> — {achievement.description}</li>
						{/each}
					</ul>
				{/if}
			</div>
		{:else}
			<div class="party-list">
				{#each gameState.player.party as member (member.id)}
					<button type="button" onclick={() => viewPartyProfile(member.profileId)}>
						<span>{member.seatNumber}. {member.gameAlias || member.displayName}</span>
						<small>{member.status}</small>
					</button>
				{/each}
			</div>
		{/if}
	{:else if activeTask === 'more'}
		{#if moreView !== 'menu'}
			<button class="back-link" type="button" onclick={() => (moreView = 'menu')}>
				<ChevronLeft size={18} /> More
			</button>
		{/if}
		{#if moreView === 'menu'}
			<div class="more-menu">
				<button type="button" onclick={() => (moreView = 'profile')}
					><UserRound size={20} /> Profile</button
				>
				<button type="button" onclick={loadHistory}><BookOpen size={20} /> History</button>
				<button type="button" onclick={toggleSound}>
					{#if soundEnabled}<Volume2 size={20} /> Sound on{:else}<VolumeX size={20} /> Sound off{/if}
				</button>
				<button type="button" onclick={() => (moreView = 'display')}
					><Settings size={20} /> Display</button
				>
				<button type="button" onclick={() => (moreView = 'connection')}
					><Wifi size={20} /> Connection details</button
				>
				<button class="danger" type="button" onclick={signOut}><LogOut size={20} /> Sign out</button
				>
			</div>
		{:else if moreView === 'profile' && profile}
			<form class="profile stack" onsubmit={saveProfile}>
				<div class="avatar-row">
					{#if profile.avatar}
						<ProtectedMedia src={profile.avatar} kind="image" alt="" />
					{:else}
						<UserRound size={50} aria-hidden="true" />
					{/if}
					<label
						>Profile image <input
							type="file"
							accept="image/jpeg,image/png,image/webp"
							onchange={(event) =>
								(avatarFile = (event.currentTarget as HTMLInputElement).files?.[0] ?? null)}
						/></label
					>
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
					help="Visible to party members. Maximum 280 characters."
				/>
				<label class="select-label"
					>Accent
					<select bind:value={profileForm.accent}>
						<option value="crimson">Crimson</option><option value="forest">Forest</option>
						<option value="navy">Navy</option><option value="gold">Gold</option><option value="plum"
							>Plum</option
						>
					</select>
				</label>
				<Button type="submit">Save profile</Button>
			</form>
		{:else if moreView === 'history'}
			{#if history}
				<div class="history stack">
					<p>
						<strong>{history.statistics.achievementPoints} points</strong> from {history.statistics
							.achievementCount} achievements
					</p>
					{#each history.games as game (game.id)}
						<article>
							<h3>{game.name}</h3>
							<p>{game.rulesetName} · {game.roleName || 'Unassigned'} · {game.outcome}</p>
						</article>
					{:else}<p>No completed games yet.</p>{/each}
				</div>
			{:else}<p>Loading history…</p>{/if}
		{:else if moreView === 'display'}
			<DisplaySettings />
		{:else if moreView === 'connection'}
			<div class="stack">
				<h3>Connection details</h3>
				<p>Connected to <strong>{location.host}</strong>.</p>
				<p>Your game state refreshes automatically after a reconnect.</p>
			</div>
		{/if}
	{/if}
</Sheet>

<style>
	.player-shell {
		display: grid;
		height: 100dvh;
		grid-template-rows: auto minmax(0, 1fr) calc(
				var(--target-size) + var(--space-2) + env(safe-area-inset-bottom)
			);
		overflow: hidden;
		background: linear-gradient(rgb(255 255 255 / 12%), transparent 14rem), var(--paper);
	}

	.status-bar {
		z-index: var(--layer-navigation);
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto var(--target-size);
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: var(--paper-light);
		padding: max(var(--space-2), env(safe-area-inset-top))
			max(var(--space-3), env(safe-area-inset-right)) var(--space-2)
			max(var(--space-3), env(safe-area-inset-left));
	}

	.game-status,
	.compact-timer {
		display: grid;
		min-width: 0;
	}

	.game-status strong,
	.game-status span,
	.compact-timer > span {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.game-status strong {
		font-family: var(--font-display);
		font-size: 0.85rem;
	}

	.game-status span,
	.compact-timer > span {
		color: var(--ink-soft);
		font-size: 0.72rem;
		text-transform: capitalize;
	}

	.compact-timer {
		justify-items: end;
	}

	.chat-action {
		position: relative;
		display: grid;
		width: var(--target-size);
		height: var(--target-size);
		place-items: center;
		border: 1px solid transparent;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
	}

	.chat-action i {
		position: absolute;
		inset-block-start: 0.35rem;
		inset-inline-end: 0.35rem;
		width: 0.65rem;
		height: 0.65rem;
		border: 2px solid var(--paper-light);
		border-radius: 50%;
		background: var(--crimson);
	}

	.current-stage {
		display: grid;
		min-height: 0;
		place-items: center;
		gap: var(--space-3);
		overflow: hidden;
		padding: var(--space-4);
	}

	.phase-stage,
	.state {
		display: grid;
		width: min(100%, 42rem);
		place-items: center;
		text-align: center;
	}

	.phase-stage p {
		margin: 0;
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	.phase-stage h1 {
		margin-block: var(--space-3);
		font-size: clamp(2rem, 11vw, 5rem);
		overflow-wrap: anywhere;
	}

	.phase-stage span {
		color: var(--ink-soft);
	}

	.cue-notice {
		position: fixed;
		z-index: var(--layer-toast);
		inset-block-start: calc(env(safe-area-inset-top) + 4.5rem);
		display: flex;
		align-items: center;
		gap: var(--space-2);
		border: var(--border-strong);
		background: var(--paper-light);
		padding: var(--space-2) var(--space-3);
	}

	.back-link {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		align-self: start;
		gap: var(--space-1);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		padding: 0;
	}

	.channel-list,
	.conversation {
		height: 100%;
	}

	.moderation-note {
		color: var(--ink-soft);
		font-size: 0.85rem;
	}

	.compose-channel {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-2);
	}

	select {
		min-width: 0;
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-2);
	}

	.channel,
	.party-list button,
	.more-menu button {
		display: flex;
		width: 100%;
		min-height: calc(var(--target-size) + var(--space-2));
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: var(--space-2);
		text-align: start;
	}

	.channel span {
		display: grid;
	}

	.channel small,
	.party-list small {
		color: var(--ink-soft);
		text-transform: capitalize;
	}

	.channel b {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.7rem;
		text-transform: uppercase;
	}

	.more-menu {
		display: grid;
	}

	.more-menu button {
		justify-content: flex-start;
	}

	.more-menu button.danger {
		color: var(--danger);
	}

	dl {
		margin: 0;
	}

	dl div {
		display: flex;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding-block: var(--space-2);
	}

	dd {
		margin: 0;
		font-weight: 700;
	}

	.avatar-row {
		display: grid;
		grid-template-columns: 4rem minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
	}

	.avatar-row :global(img) {
		width: 4rem;
		height: 4rem;
		border-radius: 50%;
		object-fit: cover;
	}

	.avatar-row :global(button) {
		grid-column: 2;
	}

	.avatar-row label,
	.avatar-row input {
		min-width: 0;
		max-width: 100%;
	}

	.select-label {
		display: grid;
		gap: var(--space-1);
	}

	.history article {
		border-block-end: var(--border-subtle);
	}

	@media (min-width: 64rem) {
		.player-shell {
			grid-template-rows: auto minmax(0, 1fr);
			padding-inline-start: 6rem;
		}

		.current-stage {
			padding: var(--space-7);
		}
	}

	@media (max-width: 24rem) {
		.compact-timer > span {
			display: none;
		}

		.status-bar {
			gap: var(--space-1);
		}
	}
</style>
