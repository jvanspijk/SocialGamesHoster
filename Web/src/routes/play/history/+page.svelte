<script lang="ts">
	import { onMount } from 'svelte';
	import PlayerHistory from '$lib/features/profiles/components/PlayerHistory.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { PlayerHistoryView } from '$lib/api/types';

	let history = $state<PlayerHistoryView | null>(null);
	let loading = $state(true);
	let loadError = $state('');

	onMount(load);

	async function load() {
		loading = true;
		loadError = '';
		try {
			history = await api<PlayerHistoryView>('/profiles/me/history');
		} catch (caught) {
			loadError = errorMessage(caught, 'History could not be loaded.');
		} finally {
			loading = false;
		}
	}
</script>

<PlayerHistory {history} {loading} {loadError} retry={load} />
