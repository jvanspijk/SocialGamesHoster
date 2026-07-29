<script lang="ts">
	import { onMount } from 'svelte';
	import { Check, Clock3, RefreshCw, UserCheck, X } from '@lucide/svelte';
	import { api, AppApiError, jsonBody, pb } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { ProfileRequest } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';
	import Button from './Button.svelte';
	import Dialog from './Dialog.svelte';
	import Field from './Field.svelte';

	let {
		compact = false,
		oncountchange,
		onapproved
	}: {
		compact?: boolean;
		oncountchange?: (count: number) => void;
		onapproved?: () => void | Promise<void>;
	} = $props();

	let requests = $state<ProfileRequest[]>([]);
	let loading = $state(true);
	let loadError = $state('');
	let rejectionTarget = $state<ProfileRequest | null>(null);
	let recoveryTarget = $state<ProfileRequest | null>(null);
	let rejectionReason = $state('');
	let busyId = $state<string | null>(null);
	let unsubscribe: (() => void) | null = null;

	onMount(() => {
		let destroyed = false;

		void (async () => {
			await load();
			if (destroyed) return;
			try {
				unsubscribe = await pb.realtime.subscribe(
					'profile-requests:game-masters',
					async () => await load(false)
				);
				if (destroyed) unsubscribe?.();
			} catch (caught) {
				toasts.error(errorMessage(caught, 'Live request updates could not be started.'));
			}
		})();

		return () => {
			destroyed = true;
			unsubscribe?.();
		};
	});

	function replaceRequests(next: ProfileRequest[]) {
		requests = next;
		oncountchange?.(next.length);
	}

	async function load(showLoading = true) {
		if (showLoading) loading = true;
		try {
			const next = await api<ProfileRequest[]>('/admin/profile-requests');
			replaceRequests(next);
			loadError = '';
		} catch (caught) {
			loadError = errorMessage(caught, 'Requests could not be loaded.');
			if (!showLoading) {
				toasts.error(loadError, { actionLabel: 'Retry', action: () => void load(false) });
			}
		} finally {
			if (showLoading) loading = false;
		}
	}

	function requestApproval(request: ProfileRequest) {
		if (request.requestType === 'recover') {
			recoveryTarget = request;
			return;
		}
		void decide(request, 'approve');
	}

	function closeRejection() {
		rejectionTarget = null;
		rejectionReason = '';
	}

	async function decide(request: ProfileRequest, decision: 'approve' | 'reject') {
		busyId = request.id;
		try {
			await api(`/admin/profile-requests/${request.id}/${decision}`, {
				method: 'POST',
				...jsonBody(decision === 'reject' ? { reason: rejectionReason } : {})
			});
			replaceRequests(requests.filter((item) => item.id !== request.id));
			rejectionTarget = null;
			recoveryTarget = null;
			rejectionReason = '';
			if (decision === 'approve') await onapproved?.();
			toasts.success(decision === 'approve' ? 'Profile approved.' : 'Request rejected.');
		} catch (caught) {
			if (
				caught instanceof AppApiError &&
				(caught.status === 404 || caught.status === 409 || caught.status === 410)
			) {
				await load(false);
				rejectionTarget = null;
				recoveryTarget = null;
				rejectionReason = '';
				toasts.info('This request was already decided. The list has been updated.');
			} else {
				toasts.error(errorMessage(caught, 'The request could not be updated.'));
			}
		} finally {
			busyId = null;
		}
	}
</script>

{#snippet content()}
	{#if loading}
		<p class="status" role="status">Loading requests…</p>
	{:else if loadError}
		<div class="error-state" role="alert">
			<div>
				<h3>Requests could not be loaded</h3>
				<p>{loadError}</p>
			</div>
			<Button variant="secondary" onclick={() => void load()}>
				<RefreshCw size={17} /> Retry
			</Button>
		</div>
	{:else if requests.length === 0}
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
					<div class="avatar" aria-hidden="true">
						{request.requestedName.slice(0, 1).toUpperCase()}
					</div>
					<div class="request-detail">
						<div class="name-row">
							<h3>{request.requestedName}</h3>
							<span class:recovery={request.requestType === 'recover'}>
								{request.requestType === 'recover' ? 'Profile recovery' : 'New profile'}
							</span>
						</div>
						<p>
							<Clock3 size={15} />
							Requested {new Date(request.createdAt).toLocaleString()}
						</p>
					</div>
					<div class="actions">
						<Button
							loading={busyId === request.id}
							disabled={busyId !== null && busyId !== request.id}
							onclick={() => requestApproval(request)}
						>
							<Check size={17} /> Approve
						</Button>
						<Button
							variant="secondary"
							disabled={busyId !== null}
							onclick={() => (rejectionTarget = request)}
						>
							<X size={17} /> Reject
						</Button>
					</div>
				</article>
			{/each}
		</div>
	{/if}
{/snippet}

{#if compact}
	<div class="compact-content">
		{@render content()}
	</div>
{:else}
	<section aria-labelledby="pending-profile-requests-heading">
		<div class="section-heading">
			<h2 id="pending-profile-requests-heading">Pending requests</h2>
			<span aria-label={`${requests.length} pending requests`}>{requests.length}</span>
		</div>
		{@render content()}
	</section>
{/if}

<Dialog
	open={recoveryTarget !== null}
	title="Approve profile recovery?"
	description={recoveryTarget ? `Approve ${recoveryTarget.requestedName} on this device?` : ''}
	close={() => (recoveryTarget = null)}
>
	<p>The previous device session will stop working immediately.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (recoveryTarget = null)}>Cancel</Button>
		<Button
			loading={recoveryTarget !== null && busyId === recoveryTarget.id}
			onclick={() => recoveryTarget && void decide(recoveryTarget, 'approve')}
		>
			Approve recovery
		</Button>
	{/snippet}
</Dialog>

<Dialog
	open={rejectionTarget !== null}
	title="Reject profile request?"
	description={rejectionTarget ? `Explain why ${rejectionTarget.requestedName} cannot enter.` : ''}
	close={closeRejection}
>
	<Field
		label="Reason (optional)"
		name="rejection-reason"
		bind:value={rejectionReason}
		multiline
		help="Add a rejection reason for the player."
		disabled={rejectionTarget !== null && busyId === rejectionTarget.id}
	/>
	{#snippet actions()}
		<Button variant="ghost" disabled={busyId !== null} onclick={closeRejection}>Cancel</Button>
		<Button
			variant="danger"
			loading={rejectionTarget !== null && busyId === rejectionTarget.id}
			onclick={() => rejectionTarget && void decide(rejectionTarget, 'reject')}
		>
			Reject request
		</Button>
	{/snippet}
</Dialog>

<style>
	section {
		margin-block-end: var(--space-7);
	}

	.compact-content {
		min-width: 0;
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
	.empty,
	.error-state {
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

	.request-detail {
		min-width: 0;
	}

	.name-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-1) var(--space-2);
	}

	h3,
	p {
		margin: 0;
	}

	.name-row span {
		border: 1px solid var(--ink-soft);
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.66rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		padding: 0.15rem 0.35rem;
		text-transform: uppercase;
	}

	.name-row span.recovery {
		border-color: var(--warning);
		color: var(--warning-ink);
	}

	article p,
	.empty p,
	.error-state p,
	.status {
		color: var(--ink-soft);
	}

	article p {
		display: flex;
		align-items: center;
		gap: var(--space-1);
	}

	.actions {
		display: flex;
		gap: var(--space-2);
	}

	.empty {
		grid-template-columns: auto 1fr;
	}

	.error-state {
		border: 1px solid var(--danger);
		padding-inline: var(--space-3);
	}

	.status {
		padding-block: var(--space-3);
	}

	@media (max-width: 47.99rem) {
		.request-list article,
		.error-state {
			grid-template-columns: auto minmax(0, 1fr);
		}

		.actions,
		.error-state > :global(button) {
			grid-column: 1 / -1;
		}

		.actions {
			display: grid;
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
