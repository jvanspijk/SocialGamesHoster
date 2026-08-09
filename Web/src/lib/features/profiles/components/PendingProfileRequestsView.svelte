<script lang="ts">
	import { Check, Clock3, UserCheck, X } from '@lucide/svelte';
	import type { ProfileRequest } from '$lib/api/types';
	import Alert from '$lib/components/Alert.svelte';
	import Button from '$lib/components/Button.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Field from '$lib/components/Field.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';

	let {
		compact = false,
		requests,
		loading,
		loadError,
		busyId,
		onretry,
		onapprove,
		onstartrejection,
		oncloserejection,
		onconfirmrejection,
		oncloserecovery,
		onconfirmrecovery,
		rejectionTarget = $bindable<ProfileRequest | null>(null),
		recoveryTarget = $bindable<ProfileRequest | null>(null),
		rejectionReason = $bindable('')
	}: {
		compact?: boolean;
		requests: ProfileRequest[];
		loading: boolean;
		loadError: string;
		busyId: string | null;
		onretry: () => void;
		onapprove: (request: ProfileRequest) => void;
		onstartrejection: (request: ProfileRequest) => void;
		oncloserejection: () => void;
		onconfirmrejection: () => void;
		oncloserecovery: () => void;
		onconfirmrecovery: () => void;
		rejectionTarget?: ProfileRequest | null;
		recoveryTarget?: ProfileRequest | null;
		rejectionReason?: string;
	} = $props();

	function requestApproval(request: ProfileRequest) {
		if (request.requestType === 'recover') {
			recoveryTarget = request;
			return;
		}
		onapprove(request);
	}
</script>

{#snippet content()}
	{#if loading}
		<LoadingState label="Loading requests…" />
	{:else if loadError}
		<Alert
			tone="error"
			title="Requests could not be loaded"
			message={loadError}
			actionLabel="Retry"
			onaction={onretry}
		/>
	{:else if requests.length === 0}
		<EmptyState title="No pending requests" description="New profile requests will appear here.">
			{#snippet icon()}<UserCheck size={34} strokeWidth={1.5} />{/snippet}
		</EmptyState>
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
							<StatusBadge
								label={request.requestType === 'recover' ? 'Profile recovery' : 'New profile'}
								tone={request.requestType === 'recover' ? 'warning' : 'info'}
							/>
						</div>
						<p><Clock3 size={15} /> Requested {new Date(request.createdAt).toLocaleString()}</p>
					</div>
					<div class="actions">
						<Button
							loading={busyId === request.id}
							disabled={busyId !== null && busyId !== request.id}
							onclick={() => requestApproval(request)}><Check size={17} /> Approve</Button
						>
						<Button
							variant="secondary"
							disabled={busyId !== null}
							onclick={() => onstartrejection(request)}><X size={17} /> Reject</Button
						>
					</div>
				</article>
			{/each}
		</div>
	{/if}
{/snippet}

{#if compact}
	<div class="compact-content">{@render content()}</div>
{:else}
	<section aria-labelledby="pending-profile-requests-heading">
		<ContentHeader density="dense" alignment="center">
			{#snippet title()}<h2 id="pending-profile-requests-heading">Pending requests</h2>{/snippet}
			{#snippet actions()}<StatusBadge
					label={`${requests.length} pending requests`}
					tone="info"
				/>{/snippet}
		</ContentHeader>
		{@render content()}
	</section>
{/if}

<Dialog
	open={recoveryTarget !== null}
	title="Approve profile recovery?"
	description={recoveryTarget ? `Approve ${recoveryTarget.requestedName} on this device?` : ''}
	close={oncloserecovery}
>
	<p>The previous device session will stop working immediately.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={oncloserecovery}>Cancel</Button>
		<Button
			loading={recoveryTarget !== null && busyId === recoveryTarget.id}
			onclick={onconfirmrecovery}>Approve recovery</Button
		>
	{/snippet}
</Dialog>

<Dialog
	open={rejectionTarget !== null}
	title="Reject profile request?"
	description={rejectionTarget ? `Explain why ${rejectionTarget.requestedName} cannot enter.` : ''}
	close={oncloserejection}
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
		<Button variant="ghost" disabled={busyId !== null} onclick={oncloserejection}>Cancel</Button>
		<Button
			variant="danger"
			loading={rejectionTarget !== null && busyId === rejectionTarget.id}
			onclick={onconfirmrejection}>Reject request</Button
		>
	{/snippet}
</Dialog>

<style>
	section {
		margin-block-end: var(--space-7);
	}
	.compact-content,
	.request-detail {
		min-width: 0;
	}
	section :global(.content-header) {
		border-block-end: var(--border-strong);
		margin-block-end: var(--space-3);
		padding-block-end: var(--space-2);
	}
	.request-list article {
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
	@media (max-width: 47.99rem) {
		.request-list article {
			grid-template-columns: auto minmax(0, 1fr);
		}
		.actions {
			display: grid;
			grid-column: 1 / -1;
			grid-template-columns: 1fr 1fr;
		}
	}
</style>
