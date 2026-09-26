<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { UserX, UserCheck } from '@lucide/svelte';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import ManagementTable from '$lib/components/ManagementTable.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import SearchField from '$lib/components/SearchField.svelte';
	import PendingProfileRequests from '$lib/features/profiles/components/PendingProfileRequests.svelte';
	import { api } from '$lib/api/client';
	import type { Profile } from '$lib/api/types';
	import { errorMessage } from '$lib/api/errors';
	import { toasts } from '$lib/state/toasts.svelte';

	let profiles = $state<Profile[]>([]);
	let loading = $state(true);
	let showDisabledProfiles = $state(false);
	let search = $state('');
	let visibleProfiles = $derived(
		profiles.filter(
			(profile) =>
				(showDisabledProfiles || profile.active) &&
				profile.displayName.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase())
		)
	);
	let emptyProfileMessage = $derived(
		search.trim()
			? 'No profiles match your search.'
			: showDisabledProfiles
				? 'No profiles.'
				: 'No active profiles.'
	);

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

<PageHeading title="Profiles & requests" variant="spacious" />

<PendingProfileRequests onapproved={loadProfiles} />

{#if loading}
	<p role="status">Loading profiles…</p>
{:else}
	<section aria-labelledby="profiles-heading">
		<h2 id="profiles-heading">Profiles</h2>
		<ManagementTable
			caption="Profiles and their current state"
			columns={[{ label: 'Profile' }, { label: 'Status' }, { label: '', align: 'end' }]}
			empty={visibleProfiles.length === 0}
			emptyMessage={emptyProfileMessage}
		>
			{#snippet controls()}
				<SearchField label="Search profiles" placeholder="Search profiles" bind:value={search} />
				<CheckboxField
					label="Show disabled profiles"
					name="show-disabled-profiles"
					bind:checked={showDisabledProfiles}
				/>
			{/snippet}
			{#each visibleProfiles as profile (profile.id)}
				<tr>
					<th scope="row" data-label="Profile">
						<a class="profile-name" href={resolve(`/admin/profiles/${profile.id}`)}>
							<div class="avatar" aria-hidden="true">
								{profile.displayName.slice(0, 1).toUpperCase()}
							</div>
							{profile.displayName}
						</a>
					</th>
					<td data-label="Status">
						<span class:disabled={!profile.active} class="status">
							{profile.active ? 'Active' : 'Disabled'}
						</span>
					</td>
					<td class="row-actions" data-label="Actions">
						<button
							class:danger={profile.active}
							type="button"
							onclick={() => setActive(profile, !profile.active)}
						>
							{#if profile.active}<UserX size={17} /> Disable{:else}<UserCheck size={17} /> Restore{/if}
						</button>
					</td>
				</tr>
			{/each}
		</ManagementTable>
	</section>
{/if}

<style>
	section {
		margin-block-end: var(--space-7);
	}

	section > h2 {
		margin-block: 0 var(--space-3);
	}

	.profile-name {
		display: flex;
		min-width: 0;
		align-items: center;
		color: inherit;
		gap: var(--space-2);
		font-family: var(--font-display);
		font-size: var(--font-size-base);
		text-decoration: none;
	}

	.profile-name:hover {
		color: var(--action-dark);
		text-decoration: underline;
		text-underline-offset: 0.2em;
	}

	.profile-name:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 3px;
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

	.status {
		display: inline-block;
		border: 1px solid var(--success);
		color: var(--success);
		font-size: var(--font-size-sm);
		padding: 0.15rem 0.45rem;
	}

	.status.disabled {
		border-color: var(--ink-faint);
		color: var(--ink-soft);
	}
</style>
