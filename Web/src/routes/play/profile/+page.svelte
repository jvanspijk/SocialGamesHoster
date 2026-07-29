<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { ArrowLeft, Save, Settings, UserRound } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import ProtectedMedia from '$lib/components/ProtectedMedia.svelte';
	import { api, AppApiError, jsonBody } from '$lib/api/client';
	import type { AppErrorBody, Profile } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { profilePreferences } from '$lib/state/profilePreferences.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type HistoryView = {
		profile: Profile;
		games: Array<{
			id: string;
			name: string;
			rulesetName: string;
			roleName: string;
			outcome: string;
			endedAt?: string;
			achievements: Array<{ id: string; title: string }>;
		}>;
		statistics: { achievementCount: number; achievementPoints: number };
	};

	let profile = $state<Profile | null>(null);
	let history = $state<HistoryView | null>(null);
	let form = $state({ displayName: '', bio: '', accent: 'crimson' });
	let busy = $state(false);
	let saveError = $state<AppErrorBody | null>(null);

	onMount(load);

	async function load() {
		try {
			[profile, history] = await Promise.all([
				api<Profile>('/profiles/me'),
				api<HistoryView>('/profiles/me/history')
			]);
			form = {
				displayName: profile.displayName,
				bio: profile.bio,
				accent: profile.accent || 'crimson'
			};
			profilePreferences.applyProfile(profile);
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'Profile details could not be loaded.',
				{
					actionLabel: 'Retry',
					action: load,
					persistent: true
				}
			);
		}
	}

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		saveError = null;
		try {
			profile = await api<Profile>('/profiles/me', { method: 'PATCH', ...jsonBody(form) });
			profilePreferences.applyProfile(profile);
			toasts.success('Profile saved.');
		} catch (caught) {
			if (caught instanceof AppApiError && caught.status === 422) {
				saveError = caught.body;
			} else {
				toasts.error(caught instanceof Error ? caught.message : 'The profile could not be saved.');
			}
		} finally {
			busy = false;
		}
	}
</script>

<div class="account-page">
	<header class="account-heading">
		<div>
			<p class="eyebrow">Player account</p>
			<h1>Profile</h1>
			<p>Your identity and game history stay outside the current game.</p>
		</div>
		<nav aria-label="Account pages">
			{#if gameState.player}
				<a href={resolve('/play')}><ArrowLeft size={18} /> Return to game</a>
			{:else}
				<a href={resolve('/')}><ArrowLeft size={18} /> Return to join page</a>
			{/if}
			<a href={resolve('/play/settings')}><Settings size={18} /> Settings</a>
		</nav>
	</header>

	{#if profile}
		<Panel title="Personal details" variant="focal">
			<form onsubmit={saveProfile}>
				<div class="profile-heading">
					<div class="avatar">
						{#if profile.avatar}<ProtectedMedia
								src={profile.avatar}
								kind="image"
								alt=""
							/>{:else}<UserRound size={34} />{/if}
					</div>
					<div><strong>{profile.displayName}</strong><span>Player profile</span></div>
				</div>
				<ErrorNotice error={saveError} />
				<Field
					label="Display name"
					name="display-name"
					bind:value={form.displayName}
					error={saveError?.fieldErrors?.displayName?.[0] ?? ''}
					required
				/>
				<Field
					label="Bio"
					name="bio"
					bind:value={form.bio}
					error={saveError?.fieldErrors?.bio?.[0] ?? ''}
					multiline
				/>
				<label>
					<span>Accent colour</span>
					<select bind:value={form.accent}>
						<option value="crimson">Crimson</option>
						<option value="forest">Forest</option>
						<option value="navy">Navy</option>
						<option value="gold">Gold</option>
						<option value="plum">Plum</option>
					</select>
				</label>
				<Button type="submit" loading={busy}><Save size={18} /> Save profile</Button>
			</form>
		</Panel>
	{:else}
		<p role="status">Loading profile…</p>
	{/if}

	{#if history}
		<Panel
			title="Game history"
			description={`${history.statistics.achievementPoints} achievement points across ${history.statistics.achievementCount} achievements`}
		>
			{#if history.games.length === 0}
				<p>No completed games yet.</p>
			{:else}
				<div class="history">
					{#each history.games as game (game.id)}
						<article>
							<div>
								<h3>{game.name}</h3>
								<p>{game.rulesetName} · {game.roleName || 'No role'}</p>
							</div>
							<strong>{game.outcome}</strong>
						</article>
					{/each}
				</div>
			{/if}
		</Panel>
	{/if}
</div>

<style>
	.account-page {
		display: grid;
		width: min(100%, 48rem);
		gap: var(--space-5);
		margin-inline: auto;
		padding: clamp(var(--space-4), 5vw, var(--space-6));
	}

	.account-heading {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-4);
	}

	.account-heading h1,
	.account-heading p {
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

	.account-heading nav {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.account-heading a {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	form {
		display: grid;
		gap: var(--space-4);
	}

	.profile-heading {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}

	.avatar {
		display: grid;
		width: 4rem;
		height: 4rem;
		overflow: hidden;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
	}

	.avatar :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.profile-heading strong,
	.profile-heading span {
		display: block;
	}

	.profile-heading span,
	.history p {
		color: var(--ink-soft);
	}

	form > label {
		display: grid;
		gap: var(--space-1);
	}

	form > label > span {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
	}

	select {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-2);
	}

	.history article {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-2) 0;
	}

	.history h3,
	.history p {
		margin: 0;
	}

	.history > article > strong {
		text-transform: capitalize;
	}

	@media (max-width: 39.99rem) {
		.account-heading {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
