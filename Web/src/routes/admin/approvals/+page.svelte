<script lang="ts">
	import { onMount } from 'svelte';
	import { Check, Clock3, UserCheck, UserX, X } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import Field from '$lib/components/Field.svelte';
	import { api, jsonBody, pb } from '$lib/api/client';
	import { toasts } from '$lib/state/toasts.svelte';

	type ProfileRequest = {
		id: string;
		requestType: string;
		requestedName: string;
		createdAt: string;
		expiresAt: string;
	};
	type Profile = {
		id: string;
		displayName: string;
		avatar: string;
		bio: string;
		accent: string;
		active: boolean;
	};

	let requests = $state<ProfileRequest[]>([]);
	let profiles = $state<Profile[]>([]);
	let loading = $state(true);
	let rejectionTarget = $state<ProfileRequest | null>(null);
	let rejectionReason = $state('');
	let busy = $state(false);
	let unsubscribe: (() => void) | null = null;

	onMount(() => {
		void load();
		return () => unsubscribe?.();
	});

	async function load() {
		try {
			[requests, profiles] = await Promise.all([
				api<ProfileRequest[]>('/admin/profile-requests'),
				api<Profile[]>('/admin/profiles')
			]);
			unsubscribe?.();
			unsubscribe = await pb.realtime.subscribe('profile-requests:game-masters', async () => {
				requests = await api<ProfileRequest[]>('/admin/profile-requests');
			});
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'Approvals could not be loaded.', {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function decide(request: ProfileRequest, decision: 'approve' | 'reject') {
		busy = true;
		try {
			await api(`/admin/profile-requests/${request.id}/${decision}`, {
				method: 'POST',
				...jsonBody(decision === 'reject' ? { reason: rejectionReason } : {})
			});
			requests = requests.filter((item) => item.id !== request.id);
			if (decision === 'approve') profiles = await api<Profile[]>('/admin/profiles');
			rejectionTarget = null;
			rejectionReason = '';
			toasts.success(decision === 'approve' ? 'Profile approved.' : 'Request rejected.');
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The request could not be updated.');
		} finally {
			busy = false;
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
			toasts.error(caught instanceof Error ? caught.message : 'The profile could not be updated.');
		}
	}
</script>

<header class="page-heading">
	<p class="eyebrow">Entry and profiles</p>
	<h1>Approvals</h1>
	<p>Review new entry requests and manage approved player profiles.</p>
</header>

{#if loading}
	<p role="status">Loading approvals…</p>
{:else}
	<section aria-labelledby="pending-heading">
		<div class="section-heading">
			<h2 id="pending-heading">Pending requests</h2>
			<span>{requests.length}</span>
		</div>
		{#if requests.length === 0}
			<div class="empty">
				<UserCheck size={34} strokeWidth={1.5} />
				<div>
					<h3>No pending requests</h3>
					<p>New profile requests will appear here.</p>
				</div>
			</div>
		{:else}
			<div class="request-list">
				{#each requests as request (request.id)}
					<article>
						<div class="avatar">{request.requestedName.slice(0, 1).toUpperCase()}</div>
						<div>
							<h3>{request.requestedName}</h3>
							<p><Clock3 size={15} /> Requested {new Date(request.createdAt).toLocaleString()}</p>
						</div>
						<div class="actions">
							<Button onclick={() => decide(request, 'approve')}><Check size={17} /> Approve</Button
							>
							<Button variant="secondary" onclick={() => (rejectionTarget = request)}
								><X size={17} /> Reject</Button
							>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</section>

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

<Dialog
	open={rejectionTarget !== null}
	title="Reject profile request?"
	description={rejectionTarget ? `Explain why ${rejectionTarget.requestedName} cannot enter.` : ''}
	close={() => (rejectionTarget = null)}
>
	<Field label="Reason" name="rejection-reason" bind:value={rejectionReason} multiline required />
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (rejectionTarget = null)}>Cancel</Button>
		<Button
			variant="danger"
			loading={busy}
			disabled={!rejectionReason.trim()}
			onclick={() => rejectionTarget && decide(rejectionTarget, 'reject')}
		>
			Reject request
		</Button>
	{/snippet}
</Dialog>

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

	.request-list article,
	.profile-grid article,
	.empty {
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
	article p,
	.empty p {
		margin: 0;
	}

	article p {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		color: var(--ink-soft);
	}

	.actions {
		display: flex;
		gap: var(--space-2);
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

	.empty {
		grid-template-columns: auto 1fr;
	}

	@media (max-width: 47.99rem) {
		.request-list article {
			grid-template-columns: auto minmax(0, 1fr);
		}

		.actions {
			grid-column: 1 / -1;
			display: grid;
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
