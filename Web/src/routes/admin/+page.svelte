<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		Archive,
		BookOpen,
		Crown,
		DatabaseBackup,
		Gamepad2,
		Settings,
		ShieldCheck,
		UserPlus,
		Users
	} from '@lucide/svelte';
	import AdminGamePanel from '$lib/components/AdminGamePanel.svelte';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import { api, AppApiError, download, jsonBody, pb } from '$lib/api/client';
	import type { AppErrorBody, AuthResponse, Game, RulesetSummary } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';

	type AdminTab = 'live' | 'games' | 'approvals' | 'rulesets' | 'owner';
	type ProfileRequest = {
		id: string;
		requestType: string;
		requestedName: string;
		createdAt: string;
		expiresAt: string;
	};
	type GameMaster = {
		id: string;
		username: string;
		displayName: string;
		isOwner: boolean;
		active: boolean;
		lastLoginAt?: string;
	};
	type HostSettings = {
		port: number;
		bindAddress: string;
		preferredAdapter: string;
		trustedLanAcknowledged: boolean;
		automaticBackups: boolean;
		privateAddresses: Array<{ adapter: string; address: string }>;
		restartRequired: boolean;
		lastRestore?: {
			status: 'success' | 'failed';
			backupName?: string;
			message: string;
			finishedAt: string;
		};
	};
	type Backup = {
		id: string;
		size: number;
		modifiedAt: string;
		automatic: boolean;
	};
	type AdminProfile = {
		id: string;
		displayName: string;
		avatar: string;
		bio: string;
		accent: string;
		active: boolean;
		approvedAt?: string;
	};

	let tab = $state<AdminTab>('live');
	let credentials = $state({ username: '', password: '' });
	let gameForm = $state({ name: '', rulesetVersionId: '' });
	let accountForm = $state({ username: '', displayName: '', password: '' });
	let games = $state<Game[]>([]);
	let rulesets = $state<RulesetSummary[]>([]);
	let requests = $state<ProfileRequest[]>([]);
	let gameMasters = $state<GameMaster[]>([]);
	let profiles = $state<AdminProfile[]>([]);
	let diagnostics = $state<Record<string, unknown> | null>(null);
	let hostSettings = $state<HostSettings | null>(null);
	let backups = $state<Backup[]>([]);
	let importFile = $state<File | null>(null);
	let error = $state<AppErrorBody | null>(null);
	let busy = $state(false);
	let showArchivedGames = $state(false);
	let gameSubscription: (() => void) | null = null;
	let profileRequestSubscription: (() => void) | null = null;

	onMount(() => {
		const requestedTab = new URL(window.location.href).searchParams.get('tab');
		if (requestedTab === 'owner') tab = 'owner';
		if (auth.isGameMaster) void loadDashboard();
		return () => {
			gameSubscription?.();
			profileRequestSubscription?.();
		};
	});

	async function login(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		error = null;
		try {
			const response = await api<AuthResponse>('/auth/game-master/login', {
				method: 'POST',
				...jsonBody(credentials)
			});
			auth.save(response);
			await loadDashboard();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function loadDashboard() {
		error = null;
		try {
			const [loadedGames, loadedRulesets, loadedRequests, loadedProfiles] = await Promise.all([
				api<Game[]>('/games'),
				api<RulesetSummary[]>('/rulesets'),
				api<ProfileRequest[]>('/admin/profile-requests'),
				api<AdminProfile[]>('/admin/profiles')
			]);
			games = loadedGames;
			rulesets = loadedRulesets;
			requests = loadedRequests;
			profiles = loadedProfiles;
			profileRequestSubscription?.();
			profileRequestSubscription = await pb.realtime.subscribe(
				'profile-requests:game-masters',
				async () => {
					requests = await api<ProfileRequest[]>('/admin/profile-requests');
				}
			);
			gameForm.rulesetVersionId ||=
				rulesets.find((item) => item.latestPublishedVersion)?.latestPublishedVersion ?? '';
			const current =
				games.find((game) => ['lobby', 'running', 'paused'].includes(game.status)) ??
				games.find((game) => game.status !== 'archived');
			if (current) await selectGame(current.id);
			if (auth.isOwner) {
				[gameMasters, hostSettings, backups] = await Promise.all([
					api<GameMaster[]>('/owner/game-masters'),
					api<HostSettings>('/owner/settings'),
					api<Backup[]>('/owner/backups')
				]);
				try {
					diagnostics = await api<Record<string, unknown>>('/diagnostics/resources');
				} catch {
					diagnostics = null;
				}
			}
		} catch (caught) {
			setError(caught);
		}
	}

	async function saveHostSettings(event: SubmitEvent) {
		event.preventDefault();
		if (!hostSettings) return;
		busy = true;
		try {
			hostSettings = await api<HostSettings>('/owner/settings', {
				method: 'PATCH',
				...jsonBody(hostSettings)
			});
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function createBackup() {
		busy = true;
		try {
			await api('/owner/backups', { method: 'POST' });
			backups = await api<Backup[]>('/owner/backups');
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function restoreBackup(backup: Backup) {
		const expected = `RESTORE ${backup.id}`;
		const confirmation = window.prompt(
			`Restoring replaces the current ledger and restarts the host. A rollback backup will be created first.\n\nType ${expected} to continue.`
		);
		if (confirmation !== expected) return;
		busy = true;
		try {
			await api(`/owner/backups/${encodeURIComponent(backup.id)}/restore`, {
				method: 'POST',
				...jsonBody({ confirmation })
			});
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function selectGame(gameId: string) {
		await gameState.refreshAdmin(gameId);
		gameSubscription?.();
		gameSubscription = await gameState.subscribe(`game:${gameId}:game-masters`, () =>
			gameState.refreshAdmin(gameId)
		);
		tab = 'live';
	}

	async function createGame(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		try {
			const created = await api<Game>('/games', {
				method: 'POST',
				...jsonBody(gameForm)
			});
			gameForm.name = '';
			games = await api<Game[]>('/games');
			await selectGame(created.id);
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function duplicateGame(game: Game) {
		try {
			const created = await api<Game>(`/games/${game.id}/duplicate`, {
				method: 'POST',
				...jsonBody({})
			});
			games = await api<Game[]>('/games');
			await selectGame(created.id);
		} catch (caught) {
			setError(caught);
		}
	}

	async function deleteGame(game: Game) {
		const expected = `DELETE ${game.id}`;
		if (
			window.prompt(
				`Permanently delete “${game.name}” and its roster, messages, awards, and audit history?\n\nType ${expected} to continue.`
			) !== expected
		)
			return;
		try {
			await api(`/games/${game.id}`, {
				method: 'DELETE',
				...jsonBody({ confirmation: expected })
			});
			games = await api<Game[]>('/games');
			if (gameState.admin?.game.id === game.id) gameState.clear();
		} catch (caught) {
			setError(caught);
		}
	}

	async function importRuleset() {
		if (!importFile) return;
		busy = true;
		try {
			const created = await api<{ ruleset: RulesetSummary }>('/rulesets/import', {
				method: 'POST',
				body: importFile,
				headers: { 'Content-Type': 'application/vnd.socialgameshoster.ruleset+zip' }
			});
			await goto(resolve('/admin/rulesets/[id]', { id: created.ruleset.id }));
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function decide(requestId: string, decision: 'approve' | 'reject') {
		try {
			await api(`/admin/profile-requests/${requestId}/${decision}`, {
				method: 'POST',
				...jsonBody(decision === 'reject' ? { reason: 'Declined by a game master.' } : {})
			});
			requests = await api<ProfileRequest[]>('/admin/profile-requests');
		} catch (caught) {
			setError(caught);
		}
	}

	async function addGameMaster(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		try {
			await api('/owner/game-masters', { method: 'POST', ...jsonBody(accountForm) });
			accountForm = { username: '', displayName: '', password: '' };
			gameMasters = await api<GameMaster[]>('/owner/game-masters');
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function setProfileActive(profile: AdminProfile, active: boolean) {
		try {
			await api(`/admin/profiles/${profile.id}/${active ? 'restore' : 'disable'}`, {
				method: 'POST',
				...jsonBody({})
			});
			profiles = await api<AdminProfile[]>('/admin/profiles');
		} catch (caught) {
			setError(caught);
		}
	}

	async function updateMaster(master: GameMaster, body: Record<string, unknown>) {
		try {
			await api(`/owner/game-masters/${master.id}`, {
				method: 'PATCH',
				...jsonBody(body)
			});
			gameMasters = await api<GameMaster[]>('/owner/game-masters');
		} catch (caught) {
			setError(caught);
		}
	}

	async function resetMasterPassword(master: GameMaster) {
		const password = window.prompt(
			`Enter a new password of at least 6 characters for ${master.displayName}.`
		);
		if (!password) return;
		try {
			await api(`/owner/game-masters/${master.id}/reset-password`, {
				method: 'POST',
				...jsonBody({ password })
			});
		} catch (caught) {
			setError(caught);
		}
	}

	async function deleteMaster(master: GameMaster) {
		if (!window.confirm(`Permanently remove the game-master account @${master.username}?`)) return;
		try {
			await api(`/owner/game-masters/${master.id}`, { method: 'DELETE' });
			gameMasters = await api<GameMaster[]>('/owner/game-masters');
		} catch (caught) {
			setError(caught);
		}
	}

	async function transferOwner(master: GameMaster) {
		const expected = `TRANSFER @${master.username}`;
		if (
			window.prompt(
				`This signs out the current owner and makes ${master.displayName} the installation owner.\n\nType ${expected} to continue.`
			) !== expected
		)
			return;
		try {
			await api(`/owner/game-masters/${master.id}`, {
				method: 'PATCH',
				...jsonBody({ makeOwner: true })
			});
			auth.clear();
			gameState.clear();
		} catch (caught) {
			setError(caught);
		}
	}

	function setError(caught: unknown) {
		error =
			caught instanceof AppApiError
				? caught.body
				: { code: 'admin.failed', message: 'The host dashboard could not be updated.' };
	}
</script>

{#if !auth.isGameMaster}
	<section class="login stack">
		<div>
			<p class="ornament">Game-master entrance</p>
			<h1>Take the host’s chair</h1>
			<p class="muted">Sign in with the named account created on the host computer.</p>
		</div>
		<form class="card stack" onsubmit={login}>
			<Crown size={34} aria-hidden="true" />
			<Field
				label="Username"
				name="username"
				bind:value={credentials.username}
				autocomplete="username"
				required
			/>
			<Field
				label="Password"
				name="password"
				type="password"
				bind:value={credentials.password}
				autocomplete="current-password"
				required
			/>
			<ErrorNotice {error} />
			<Button type="submit" loading={busy}>Enter dashboard</Button>
		</form>
	</section>
{:else}
	<div class="admin-shell">
		<aside class="admin-nav">
			<div class="host-name">
				<ShieldCheck size={26} />
				<div>
					<strong>{auth.actor?.displayName}</strong><small
						>{auth.isOwner ? 'Owner' : 'Game master'}</small
					>
				</div>
			</div>
			<nav aria-label="Game-master dashboard">
				<button class:active={tab === 'live'} onclick={() => (tab = 'live')}
					><Gamepad2 size={19} /> Live table</button
				>
				<button class:active={tab === 'games'} onclick={() => (tab = 'games')}
					><Archive size={19} /> Games</button
				>
				<button class:active={tab === 'approvals'} onclick={() => (tab = 'approvals')}
					><UserPlus size={19} /> Approvals <em>{requests.length}</em></button
				>
				<button class:active={tab === 'rulesets'} onclick={() => (tab = 'rulesets')}
					><BookOpen size={19} /> Rulesets</button
				>
				{#if auth.isOwner}<button class:active={tab === 'owner'} onclick={() => (tab = 'owner')}
						><Settings size={19} /> Installation</button
					>{/if}
			</nav>
		</aside>

		<main class="admin-content stack">
			<ErrorNotice {error} />
			{#if tab === 'live'}
				{#if gameState.admin}
					<AdminGamePanel
						view={gameState.admin}
						onrefresh={async () => {
							await gameState.refreshAdmin(gameState.admin!.game.id);
							games = await api<Game[]>('/games');
						}}
					/>
				{:else}
					<div class="card empty">
						<Gamepad2 size={42} />
						<h2>No game selected</h2>
						<p>Create a draft or select one from the games ledger.</p>
						<Button onclick={() => (tab = 'games')}>Open games</Button>
					</div>
				{/if}
			{:else if tab === 'games'}
				<section class="stack">
					<div>
						<p class="ornament">Game ledger</p>
						<h1>Drafts and history</h1>
					</div>
					<form class="card new-game" onsubmit={createGame}>
						<Field label="Game name" name="gameName" bind:value={gameForm.name} required />
						<label>
							<span>Ruleset</span>
							<select bind:value={gameForm.rulesetVersionId} required>
								{#each rulesets.filter((item) => item.latestPublishedVersion) as ruleset (ruleset.id)}
									<option value={ruleset.latestPublishedVersion}>{ruleset.name}</option>
								{/each}
							</select>
						</label>
						<Button type="submit" loading={busy}>Create draft</Button>
					</form>
					<label class="check">
						<input type="checkbox" bind:checked={showArchivedGames} />
						<span>Show archived games</span>
					</label>
					<div class="game-list">
						{#each games.filter((game) => showArchivedGames || game.status !== 'archived') as game (game.id)}
							<article>
								<button class="game-open" onclick={() => selectGame(game.id)}>
									<div><strong>{game.name}</strong><small>{game.status}</small></div>
									<span
										>{game.startedAt
											? new Date(game.startedAt).toLocaleDateString()
											: 'Not started'} →</span
									>
								</button>
								<div class="game-actions">
									<Button variant="ghost" onclick={() => duplicateGame(game)}>Duplicate</Button>
									{#if ['draft', 'review', 'archived'].includes(game.status)}
										<Button variant="danger" onclick={() => deleteGame(game)}>Delete</Button>
									{/if}
								</div>
							</article>
						{:else}
							<p class="card">No game drafts yet.</p>
						{/each}
					</div>
				</section>
			{:else if tab === 'approvals'}
				<section class="stack">
					<div>
						<p class="ornament">Passwordless profiles</p>
						<h1>Entry requests</h1>
					</div>
					{#each requests as request (request.id)}
						<article class="card request">
							<div>
								<h3>{request.requestedName}</h3>
								<p>
									{request.requestType === 'recover'
										? 'Recover existing profile'
										: 'Create new profile'} · {new Date(request.createdAt).toLocaleTimeString()}
								</p>
							</div>
							<div>
								<Button variant="secondary" onclick={() => decide(request.id, 'reject')}
									>Reject</Button
								>
								<Button onclick={() => decide(request.id, 'approve')}>Approve</Button>
							</div>
						</article>
					{:else}
						<div class="card empty">
							<Users size={40} />
							<h2>No pending requests</h2>
							<p>New and returning players will appear here.</p>
						</div>
					{/each}
					<div>
						<p class="ornament">Approved devices</p>
						<h2>Player profiles</h2>
					</div>
					<div class="profile-list">
						{#each profiles as profile (profile.id)}
							<article class="card profile-row">
								<div>
									<strong>{profile.displayName}</strong>
									<small>{profile.active ? 'Active' : 'Disabled'}</small>
								</div>
								<Button
									variant={profile.active ? 'danger' : 'secondary'}
									onclick={() => setProfileActive(profile, !profile.active)}
								>
									{profile.active ? 'Disable and sign out' : 'Restore'}
								</Button>
							</article>
						{:else}
							<p class="card">No approved player profiles yet.</p>
						{/each}
					</div>
				</section>
			{:else if tab === 'rulesets'}
				<section class="stack">
					<div>
						<p class="ornament">Rules of play</p>
						<h1>Ruleset library</h1>
					</div>
					<div class="card import-row">
						<label>
							<span>Import a `.sghrules` bundle</span>
							<input
								type="file"
								accept=".sghrules,application/vnd.socialgameshoster.ruleset+zip"
								onchange={(event) =>
									(importFile = (event.currentTarget as HTMLInputElement).files?.[0] ?? null)}
							/>
						</label>
						<Button
							variant="secondary"
							disabled={!importFile}
							loading={busy}
							onclick={importRuleset}>Review imported draft</Button
						>
					</div>
					<div class="grid">
						{#each rulesets as ruleset (ruleset.id)}
							<a class="card ruleset" href={resolve('/admin/rulesets/[id]', { id: ruleset.id })}>
								<BookOpen size={28} />
								<h2>{ruleset.name}</h2>
								<p>
									{ruleset.latestPublishedVersion
										? 'Published and ready for games.'
										: 'Draft only.'}
								</p>
								<span class="display">Open creator →</span>
							</a>
						{/each}
						<a class="card ruleset new" href={resolve('/admin/rulesets/[id]', { id: 'new' })}
							><UserPlus size={28} />
							<h2>New ruleset</h2>
							<p>Start a declarative game definition.</p></a
						>
					</div>
				</section>
			{:else if tab === 'owner' && auth.isOwner}
				<section class="stack">
					<div>
						<p class="ornament">Owner controls</p>
						<h1>Installation and accounts</h1>
					</div>
					<div class="grid">
						{#if hostSettings}
							<form class="card stack" onsubmit={saveHostSettings}>
								<h2>Network hosting</h2>
								<label>
									<span>Port</span>
									<input
										name="hostPort"
										type="number"
										min="1"
										max="65535"
										bind:value={hostSettings.port}
										required
									/>
								</label>
								<label>
									<span>Private network adapter</span>
									<select bind:value={hostSettings.preferredAdapter}>
										<option value="">Choose automatically</option>
										{#each hostSettings.privateAddresses as address (address.adapter + address.address)}
											<option value={address.adapter}>{address.adapter} · {address.address}</option>
										{/each}
									</select>
								</label>
								<label class="check">
									<input type="checkbox" bind:checked={hostSettings.automaticBackups} />
									<span>Create one automatic backup each day</span>
								</label>
								<label class="check">
									<input type="checkbox" bind:checked={hostSettings.trustedLanAcknowledged} />
									<span
										>I understand that everyone on this trusted private network can reach the host.</span
									>
								</label>
								<p class="muted">
									Network changes take effect after restarting Social Games Hoster.
								</p>
								<Button type="submit" loading={busy}>Save hosting settings</Button>
							</form>
							<section class="card stack join-card">
								<h2>Phone join</h2>
								<img src="/api/app/v1/setup/join-qr" alt="QR code for the player join page" />
								{#if hostSettings.privateAddresses.length}
									<code>http://{hostSettings.privateAddresses[0].address}:{hostSettings.port}/</code
									>
								{:else}
									<p>No active private IPv4 network was found.</p>
								{/if}
							</section>
						{/if}
					</div>
					<div class="grid">
						<section class="card stack">
							<div class="section-title">
								<Users size={24} />
								<h2>Game masters</h2>
							</div>
							{#each gameMasters as master (master.id)}
								<div class="master">
									<div><strong>{master.displayName}</strong><small>@{master.username}</small></div>
									<span>{master.isOwner ? 'Owner' : master.active ? 'Active' : 'Disabled'}</span>
									{#if !master.isOwner}
										<div class="master-actions">
											<Button
												variant="secondary"
												onclick={() => updateMaster(master, { active: !master.active })}
												>{master.active ? 'Disable' : 'Enable'}</Button
											>
											<Button variant="ghost" onclick={() => resetMasterPassword(master)}
												>Reset password</Button
											>
											<Button variant="ghost" onclick={() => transferOwner(master)}
												>Make owner</Button
											>
											<Button variant="danger" onclick={() => deleteMaster(master)}>Remove</Button>
										</div>
									{/if}
								</div>
							{/each}
						</section>
						<form class="card stack" onsubmit={addGameMaster}>
							<h2>Add game master</h2>
							<Field
								label="Username"
								name="newUsername"
								bind:value={accountForm.username}
								required
							/>
							<Field
								label="Display name"
								name="newDisplayName"
								bind:value={accountForm.displayName}
								required
							/>
							<Field
								label="Temporary password"
								name="newPassword"
								type="password"
								bind:value={accountForm.password}
								required
							/>
							<Button type="submit" loading={busy}>Create account</Button>
						</form>
					</div>
					<div class="grid">
						<section class="card stack">
							<div class="section-title">
								<DatabaseBackup size={24} />
								<h2>Backups</h2>
							</div>
							<p>Automatic daily and pre-upgrade backups protect the local game ledger.</p>
							{#if hostSettings?.lastRestore}
								<p
									class:restore-failed={hostSettings.lastRestore.status === 'failed'}
									class="restore-report"
								>
									<strong
										>Last restore: {hostSettings.lastRestore.status === 'success'
											? 'completed'
											: 'failed'}</strong
									>
									<span>{hostSettings.lastRestore.message}</span>
									<small>{new Date(hostSettings.lastRestore.finishedAt).toLocaleString()}</small>
								</p>
							{/if}
							<Button variant="secondary" loading={busy} onclick={createBackup}
								>Create backup now</Button
							>
							<div class="backup-list">
								{#each backups as backup (backup.id)}
									<div>
										<span
											><strong>{backup.automatic ? 'Automatic' : 'Manual'}</strong><small
												>{new Date(backup.modifiedAt).toLocaleString()}</small
											></span
										>
										<Button variant="danger" onclick={() => restoreBackup(backup)}>Restore</Button>
									</div>
								{:else}
									<p class="muted">No backups have been created yet.</p>
								{/each}
							</div>
						</section>
						<section class="card stack">
							<h2>Diagnostics</h2>
							{#if diagnostics}
								<dl>
									{#each Object.entries(diagnostics) as [key, value] (key)}
										<div>
											<dt>{key}</dt>
											<dd>{String(value)}</dd>
										</div>
									{/each}
								</dl>
							{:else}
								<p>Diagnostic mode is not enabled for this launch.</p>
							{/if}
							{#if diagnostics}
								<Button
									variant="secondary"
									onclick={() =>
										download(
											'/diagnostics/support-bundle',
											'social-games-hoster-support.zip',
											'POST'
										)}>Download support bundle</Button
								>
							{/if}
						</section>
					</div>
				</section>
			{/if}
		</main>
	</div>
{/if}

<style>
	.login {
		max-width: 32rem;
		margin: 5vh auto;
	}

	.admin-shell {
		display: grid;
		grid-template-columns: 13rem minmax(0, 1fr);
		gap: 1.2rem;
	}

	.admin-nav {
		position: sticky;
		top: 1rem;
		align-self: start;
		border: 1px solid #8d7248;
		background: rgb(255 249 230 / 52%);
		padding: 0.6rem;
	}

	.host-name,
	.section-title {
		display: flex;
		align-items: center;
		gap: 0.55rem;
	}

	.check {
		display: flex;
		align-items: flex-start;
		gap: 0.6rem;
	}

	.check input {
		width: 1.1rem;
		height: 1.1rem;
		margin-top: 0.15rem;
	}

	.join-card img {
		width: min(100%, 18rem);
		border: 1px solid #8d7248;
		background: white;
		padding: 0.4rem;
	}

	.join-card code {
		overflow-wrap: anywhere;
	}

	.backup-list > div {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		border-top: 1px solid #c7aa78;
		padding-block: 0.65rem;
	}

	.profile-list,
	.master-actions {
		display: grid;
		gap: 0.45rem;
	}

	.profile-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}

	.profile-row > div {
		display: grid;
	}

	.master {
		display: grid;
		gap: 0.45rem;
		border-top: 1px solid #c7aa78;
		padding-block: 0.65rem;
	}

	.restore-report {
		display: grid;
		gap: 0.15rem;
		border-inline-start: 3px solid var(--success);
		background: rgb(255 249 230 / 45%);
		padding: 0.55rem 0.7rem;
	}

	.restore-report.restore-failed {
		border-color: var(--danger);
	}

	.master-actions {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.backup-list span {
		display: grid;
	}

	.import-row {
		display: flex;
		align-items: end;
		justify-content: space-between;
		gap: 1rem;
	}

	.import-row label {
		display: grid;
		flex: 1;
		gap: 0.35rem;
	}

	.host-name {
		border-bottom: 1px solid #b99b6c;
		padding: 0.5rem;
	}

	.host-name div,
	.master div {
		display: grid;
	}

	.host-name small,
	.master small {
		color: var(--ink-faint);
	}

	.admin-nav nav {
		display: grid;
		margin-top: 0.5rem;
	}

	.admin-nav button {
		display: grid;
		min-height: 48px;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 0.5rem;
		border: 0;
		border-inline-start: 3px solid transparent;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.67rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		padding: 0.5rem;
		text-align: start;
		text-transform: uppercase;
	}

	.admin-nav button.active {
		border-color: var(--crimson);
		background: rgb(166 42 42 / 8%);
		color: var(--crimson-dark);
	}

	.admin-nav em {
		border-radius: 50%;
		background: var(--crimson);
		color: white;
		font-style: normal;
		padding: 0.15rem 0.4rem;
	}

	.admin-content {
		min-width: 0;
	}

	.empty {
		display: grid;
		place-items: center;
		max-width: 35rem;
		margin: 5vh auto;
		text-align: center;
	}

	.new-game {
		display: grid;
		grid-template-columns: minmax(12rem, 1fr) minmax(12rem, 1fr) auto;
		align-items: end;
	}

	.new-game label {
		display: grid;
		gap: 0.3rem;
	}

	.new-game label span {
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

	.game-list {
		display: grid;
		gap: 0.5rem;
	}

	.game-list article {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		border: 1px solid #a98d61;
		background: rgb(255 249 230 / 62%);
	}

	.game-open {
		display: flex;
		min-height: 58px;
		align-items: center;
		justify-content: space-between;
		border: 0;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: 0.65rem 0.8rem;
		text-align: start;
	}

	.game-open div {
		display: grid;
	}

	.game-actions {
		display: flex;
		gap: 0.35rem;
		padding-inline-end: 0.5rem;
	}

	.game-list small,
	.game-list span {
		color: var(--ink-faint);
		text-transform: capitalize;
	}

	.request,
	.master {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}

	.request > div:last-child {
		display: flex;
		gap: 0.5rem;
	}

	.ruleset {
		color: var(--ink);
		text-decoration: none;
		transition: transform var(--speed-fast) ease-out;
	}

	.ruleset:hover {
		transform: translateY(-2px);
	}

	.ruleset .display {
		color: var(--crimson-dark);
		font-size: 0.7rem;
	}

	.ruleset.new {
		border-style: dashed;
	}

	.master {
		border-bottom: 1px dotted #a98d61;
		padding-block: 0.45rem;
	}

	dl div {
		display: flex;
		justify-content: space-between;
		gap: 0.6rem;
		border-bottom: 1px dotted #a98d61;
		padding-block: 0.3rem;
	}

	dd {
		margin: 0;
	}

	@media (max-width: 800px) {
		.admin-shell {
			grid-template-columns: 1fr;
		}

		.admin-nav {
			position: static;
		}

		.admin-nav nav {
			display: flex;
			overflow-x: auto;
		}

		.admin-nav button {
			min-width: 9rem;
		}

		.new-game {
			grid-template-columns: 1fr;
		}
	}
</style>
