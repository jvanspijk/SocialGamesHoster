<script lang="ts">
	import { onMount } from 'svelte';
	import { api, pb } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { ProfileRequest } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let {
		oncountchange
	}: {
		oncountchange: (count: number) => void;
	} = $props();

	onMount(() => {
		let destroyed = false;
		let unsubscribe: (() => void) | null = null;

		async function refresh() {
			try {
				const requests = await api<ProfileRequest[]>('/admin/profile-requests');
				if (!destroyed) oncountchange(requests.length);
			} catch (caught) {
				if (destroyed) return;
				toasts.error(errorMessage(caught, 'Pending profile requests could not be loaded.'), {
					actionLabel: 'Retry',
					action: refresh,
					persistent: true
				});
			}
		}

		void (async () => {
			try {
				unsubscribe = await pb.realtime.subscribe('profile-requests:game-masters', refresh);
				if (destroyed) unsubscribe?.();
			} catch (caught) {
				if (destroyed) return;
				toasts.error(errorMessage(caught, 'Live profile request updates could not be started.'));
			}
			if (!destroyed) await refresh();
		})();

		return () => {
			destroyed = true;
			unsubscribe?.();
		};
	});
</script>
