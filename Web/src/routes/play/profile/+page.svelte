<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { ArrowLeft, History, Save, Settings, UserRound } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import { fieldError, toFormError, type FormError } from '$lib/forms/errors';
	import type { Profile } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let profile = $state<Profile | null>(null);
	let form = $state({ displayName: '', bio: '' });
	let busy = $state(false);
	let saveError = $state<FormError | null>(null);

	onMount(load);

	async function load() {
		try {
			profile = await api<Profile>('/profiles/me');
			form = {
				displayName: profile.displayName,
				bio: profile.bio
			};
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Profile details could not be loaded.'), {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		}
	}

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		saveError = null;
		try {
			profile = await api<Profile>('/profiles/me', { method: 'PATCH', ...jsonBody(form) });
			auth.updateDisplayName(profile.displayName);
			toasts.success('Profile saved.');
		} catch (caught) {
			const nextError = toFormError(caught, 'The profile could not be saved.');
			if (nextError.kind === 'validation') {
				saveError = nextError;
			} else {
				toasts.error(nextError.message);
			}
		} finally {
			busy = false;
		}
	}
</script>

<div class="account-page">
	<PageHeading
		title="Profile"
		description="Your player details are shared across games."
		variant="flush"
	>
		{#snippet actions()}
			<nav aria-label="Account pages">
				<a href={resolve('/play')}><ArrowLeft size={18} /> Player home</a>
				<a href={resolve('/play/history')}><History size={18} /> History</a>
				<a href={resolve('/play/settings')}><Settings size={18} /> Settings</a>
			</nav>
		{/snippet}
	</PageHeading>

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
				<ErrorNotice message={saveError?.message} traceId={saveError?.traceId} />
				<Field
					label="Display name"
					name="display-name"
					bind:value={form.displayName}
					error={fieldError(saveError, 'displayName')}
					required
				/>
				<Field
					label="Bio"
					name="bio"
					bind:value={form.bio}
					error={fieldError(saveError, 'bio')}
					multiline
				/>
				<Button type="submit" loading={busy}><Save size={18} /> Save profile</Button>
			</form>
		</Panel>
	{:else}
		<p role="status">Loading profile…</p>
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

	nav {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	nav a {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--action-dark);
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		text-decoration: none;
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

	.profile-heading span {
		color: var(--ink-soft);
	}
</style>
