<script lang="ts">
	import { onMount } from 'svelte';
	import { api, AppApiError, jsonBody, pb } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { ProfileRequest } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';
	import PendingProfileRequestsView from './PendingProfileRequestsView.svelte';

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

<PendingProfileRequestsView
	{compact}
	{requests}
	{loading}
	{loadError}
	{busyId}
	onretry={() => void load()}
	onapprove={(request) => void decide(request, 'approve')}
	onstartrejection={(request) => (rejectionTarget = request)}
	oncloserejection={closeRejection}
	onconfirmrejection={() => rejectionTarget && void decide(rejectionTarget, 'reject')}
	oncloserecovery={() => (recoveryTarget = null)}
	onconfirmrecovery={() => recoveryTarget && void decide(recoveryTarget, 'approve')}
	bind:rejectionTarget
	bind:recoveryTarget
	bind:rejectionReason
/>
