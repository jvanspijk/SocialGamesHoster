<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Crown, Hourglass, QrCode, ScrollText, UsersRound } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import { api, AppApiError, jsonBody, pb } from '$lib/api/client';
	import type { AppErrorBody, AuthResponse, Game } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';

	type PendingRequest = {
		requestId: string;
		secret: string;
		realtimeTopic?: string;
		status: string;
		expiresAt: string;
		reason?: string;
	};

	let needsOwner = $state<boolean | null>(null);
	let owner = $state({
		username: '',
		displayName: '',
		password: '',
		trustedLanAcknowledged: false
	});
	let displayName = $state('');
	let pending = $state<PendingRequest | null>(null);
	let liveGame = $state<Game | null>(null);
	let joinUrl = $state('');
	let showQr = $state(false);
	let error = $state<AppErrorBody | null>(null);
	let busy = $state(false);

	onMount(() => {
		showQr = new URL(window.location.href).searchParams.get('showQr') === '1';
		void initialize();
		const timer = window.setInterval(() => void pollPending(), 1800);
		return () => window.clearInterval(timer);
	});

	async function initialize() {
		try {
			const status = await api<{ needsOwner: boolean; joinUrl: string }>('/setup/status');
			needsOwner = status.needsOwner;
			joinUrl = status.joinUrl;
			if (!needsOwner) {
				try {
					liveGame = await api<Game>('/games/live');
				} catch {
					liveGame = null;
				}
			}
			const stored = sessionStorage.getItem('sgh.profile-request');
			if (stored) {
				pending = JSON.parse(stored) as PendingRequest;
				await subscribePending();
			}
		} catch (caught) {
			setError(caught);
		}
	}

	async function createOwner(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		error = null;
		try {
			const response = await api<{
				token: string;
				gameMaster: { id: string; displayName: string; isOwner: boolean };
			}>('/setup/owner', { method: 'POST', ...jsonBody(owner) });
			pb.authStore.save(response.token, {
				id: response.gameMaster.id,
				type: 'game_masters',
				collectionId: 'game_masters',
				collectionName: 'game_masters',
				displayName: response.gameMaster.displayName,
				isOwner: true
			});
			await goto(resolve('/admin'));
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function requestProfile(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		error = null;
		try {
			pending = await api<PendingRequest>('/auth/player/requests', {
				method: 'POST',
				...jsonBody({ displayName })
			});
			sessionStorage.setItem('sgh.profile-request', JSON.stringify(pending));
			await subscribePending();
		} catch (caught) {
			setError(caught);
		} finally {
			busy = false;
		}
	}

	async function subscribePending() {
		if (!pending?.realtimeTopic) return;
		await pb.realtime.unsubscribe(pending.realtimeTopic);
		await pb.realtime.subscribe(pending.realtimeTopic, async (raw) => {
			const envelope = raw as unknown as {
				payload?: { status?: string; reason?: string; expiresAt?: string };
			};
			if (!pending || !envelope.payload?.status) return;
			pending = {
				...pending,
				status: envelope.payload.status,
				reason: envelope.payload.reason,
				expiresAt: envelope.payload.expiresAt ?? pending.expiresAt
			};
			sessionStorage.setItem('sgh.profile-request', JSON.stringify(pending));
			if (pending.status === 'approved') await redeem();
		});
	}

	async function pollPending() {
		if (!pending || pending.status !== 'pending') return;
		try {
			const status = await api<Omit<PendingRequest, 'secret'>>(
				`/auth/player/requests/${pending.requestId}`,
				{ headers: { 'X-Profile-Request-Secret': pending.secret } }
			);
			pending = { ...pending, ...status };
			sessionStorage.setItem('sgh.profile-request', JSON.stringify(pending));
			if (pending.status === 'approved') await redeem();
		} catch (caught) {
			setError(caught);
		}
	}

	async function redeem() {
		if (!pending) return;
		const response = await api<AuthResponse>(`/auth/player/requests/${pending.requestId}/redeem`, {
			method: 'POST',
			...jsonBody({ secret: pending.secret })
		});
		auth.save(response);
		sessionStorage.removeItem('sgh.profile-request');
		pending = null;
		if (liveGame?.status === 'lobby') {
			await api(`/games/${liveGame.id}/join`, { method: 'POST', ...jsonBody({}) });
		}
		await goto(resolve('/play'));
	}

	function startOver() {
		if (pending?.realtimeTopic) void pb.realtime.unsubscribe(pending.realtimeTopic);
		pending = null;
		sessionStorage.removeItem('sgh.profile-request');
	}

	function setError(caught: unknown) {
		error =
			caught instanceof AppApiError
				? caught.body
				: { code: 'network.failed', message: 'The local host could not be reached.' };
	}
</script>

{#if needsOwner === null}
	<div class="card loading-card">
		<Hourglass aria-hidden="true" />
		<p>Loading host settings…</p>
	</div>
{:else if needsOwner}
	<section class="setup stack">
		<div>
			<p class="ornament">First launch</p>
			<h1>Set up the host</h1>
			<p class="lead">
				Create the owner account on this computer. The included demonstration rulesets will be added
				automatically.
			</p>
		</div>
		<form class="card stack" onsubmit={createOwner}>
			<Crown size={34} aria-hidden="true" />
			<Field
				label="Username"
				name="username"
				bind:value={owner.username}
				autocomplete="username"
				required
			/>
			<Field label="Display name" name="displayName" bind:value={owner.displayName} required />
			<Field
				label="Password"
				name="password"
				type="password"
				bind:value={owner.password}
				autocomplete="new-password"
				help="At least 6 characters."
				required
			/>
			<div class="trust-notice">
				<strong>Use a trusted local network</strong>
				<p>
					People connected to this LAN can reach the host. Use it on your private party network, not
					on public Wi-Fi.
				</p>
				<label>
					<input type="checkbox" bind:checked={owner.trustedLanAcknowledged} required />
					I understand and trust this local network.
				</label>
			</div>
			<ErrorNotice {error} />
			<Button type="submit" loading={busy}>Create owner</Button>
		</form>
	</section>
{:else}
	<section class="hero stack">
		<div class="hero-copy">
			<p class="ornament">Your game · Your network</p>
			<h1>Gather the party.<br />Keep every secret.</h1>
			<p class="lead">
				Join the game hosted on this local network. No email, internet account, or cloud service is
				required.
			</p>
		</div>

		<div class="grid">
			{#if showQr}
				<section class="card stack qr-card">
					<QrCode size={34} aria-hidden="true" />
					<h2>Scan to join</h2>
					<img src="/api/app/v1/setup/join-qr" alt="QR code for this player join page" />
					<code>{joinUrl}</code>
				</section>
			{/if}
			{#if auth.isPlayer}
				<div class="card stack">
					<UsersRound size={34} aria-hidden="true" />
					<h2>Welcome back, {auth.actor?.displayName}</h2>
					<p>
						{liveGame ? `${liveGame.name} is ${liveGame.status}.` : 'There is no live game yet.'}
					</p>
					<Button onclick={() => goto(resolve('/play'))}>Open game</Button>
				</div>
			{:else if pending}
				<div class="card stack">
					<Hourglass size={34} aria-hidden="true" />
					<h2>{pending.status === 'pending' ? 'Awaiting approval' : 'Request update'}</h2>
					<p>
						{#if pending.status === 'pending'}
							Ask a game master to approve your profile. This page will continue automatically.
						{:else if pending.status === 'rejected'}
							The request was declined. {pending.reason ?? ''}
						{:else}
							This request is no longer available.
						{/if}
					</p>
					<ErrorNotice {error} />
					<Button variant="secondary" onclick={startOver}>Use another name</Button>
				</div>
			{:else}
				<form class="card stack" onsubmit={requestProfile}>
					<ScrollText size={34} aria-hidden="true" />
					<h2>Enter by profile</h2>
					<p class="muted">
						Use your existing name to recover it on this device, or choose a new one.
					</p>
					<Field
						label="Profile name"
						name="displayName"
						bind:value={displayName}
						autocomplete="nickname"
						required
					/>
					{#if liveGame}
						<p class="game-line"><strong>{liveGame.name}</strong> · {liveGame.status}</p>
					{/if}
					<ErrorNotice {error} />
					<Button type="submit" loading={busy}>Request entry</Button>
				</form>
			{/if}

			<a class="host-card card" href={resolve('/admin')}>
				<Crown size={34} aria-hidden="true" />
				<h2>Game master</h2>
				<p>Open the host dashboard, approve players, and guide the game.</p>
				<span class="display">Enter host view →</span>
			</a>
		</div>
	</section>
{/if}

<style>
	.hero,
	.setup {
		max-width: 58rem;
		margin: 3vh auto 0;
	}

	.hero-copy {
		max-width: 42rem;
		padding-block: 1rem;
	}

	.lead {
		max-width: 46rem;
		color: var(--ink-soft);
		font-size: clamp(1.05rem, 2.4vw, 1.35rem);
	}

	.setup form {
		width: min(100%, 31rem);
	}

	.trust-notice {
		display: grid;
		gap: 0.4rem;
		border: 1px solid var(--gold-dark);
		background: rgb(255 249 230 / 55%);
		padding: 0.8rem;
	}

	.trust-notice p {
		margin: 0;
	}

	.trust-notice label {
		display: flex;
		align-items: flex-start;
		gap: 0.45rem;
	}

	.loading-card {
		display: grid;
		max-width: 24rem;
		place-items: center;
		gap: 0.5rem;
		margin: 12vh auto;
	}

	.host-card {
		display: grid;
		align-content: start;
		color: var(--ink);
		text-decoration: none;
		transition: transform var(--speed-fast) ease-out;
	}

	.host-card:hover {
		transform: translateY(-2px);
	}

	.host-card .display {
		color: var(--crimson-dark);
		font-size: 0.72rem;
	}

	.game-line {
		border-block: 1px solid #b99b6c;
		padding-block: 0.55rem;
	}

	.qr-card img {
		width: min(100%, 20rem);
		border: 1px solid #8d7248;
		background: white;
		padding: 0.4rem;
	}

	.qr-card code {
		overflow-wrap: anywhere;
	}
</style>
