<script lang="ts">
	import { onMount } from 'svelte';
	import { UserX, UserCheck } from '@lucide/svelte';
	import PendingProfileRequests from '$lib/components/PendingProfileRequests.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import { toasts } from '$lib/state/toasts.svelte';

	type Profile = {
		id: string;
		displayName: string;
		avatar: string;
		bio: string;
		accent: string;
		active: boolean;
	};

	let profiles = $state<Profile[]>([]);
	let loading = $state(true);

	onMount(() => {
		void loadProfiles();
	});

	async function loadProfiles() {
		try {
			profiles = await api<Profile[]>('/admin/profiles');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Profiles could not be loaded.'), {
				actionLabel: 'Retry profiles',
				action: loadProfiles,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function setActive(profile: Profile, active: boolean) {
		try {
			await api(`/admin/profiles/${profile.id}/${active ? 'restore' : 'disable'}`, {
				method: 'POST'
			});
			profiles = profiles.map((item) => (item.id === profile.id ? { ...item, active } : item));
			toasts.success(active ? 'Profile restored.' : 'Profile disabled.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The profile could not be updated.'));
		}
	}
</script>

<header class="page-heading">
	<p class="eyebrow">Entry and profiles</p>
	<h1>Approvals</h1>
	<p>Review new entry requests and manage approved player profiles.</p>
</header>

<PendingProfileRequests onapproved={loadProfiles} />

{#if loading}
	<p role="status">Loading profiles…</p>
{:else}
	<section aria-labelledby="profiles-heading">
		<div class="section-heading">
			<h2 id="profiles-heading">Approved profiles</h2>
			<span>{profiles.length}</span>
		</div>
		<div class="profile-grid">
			{#each profiles as profile (profile.id)}
				<article class:disabled={!profile.active}>
					<div class="avatar">{profile.displayName.slice(0, 1).toUpperCase()}</div>
					<div>
						<h3>{profile.displayName}</h3>
						<p>{profile.active ? 'Active' : 'Disabled'}</p>
					</div>
					<button type="button" onclick={() => setActive(profile, !profile.active)}>
						{#if profile.active}<UserX size={17} /> Disable{:else}<UserCheck size={17} /> Restore{/if}
					</button>
				</article>
			{/each}
		</div>
	</section>
{/if}

<style>
	.page-heading {
		margin-block-end: var(--space-6);
	}

	.page-heading h1,
	.page-heading p {
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

	section {
		margin-block-end: var(--space-7);
	}

	.section-heading {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		border-block-end: var(--border-strong);
		margin-block-end: var(--space-3);
	}

	.section-heading h2 {
		margin: 0;
	}

	.section-heading span {
		display: grid;
		min-width: 1.6rem;
		height: 1.6rem;
		place-items: center;
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--paper-light);
		font-size: 0.78rem;
	}

	.profile-grid article {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.avatar {
		display: grid;
		width: 2.8rem;
		height: 2.8rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	h3,
	.profile-grid article p {
		margin: 0;
	}

	.profile-grid article p {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		color: var(--ink-soft);
	}

	.profile-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 20rem), 1fr));
		gap: 0 var(--space-5);
	}

	.profile-grid article.disabled {
		opacity: 0.65;
	}

	.profile-grid button {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
	}
</style>
