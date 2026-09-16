<script lang="ts">
	import { onMount } from 'svelte';
	import AdminHome from '$lib/features/games/components/AdminHome.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Game } from '$lib/api/types';

	let games = $state<Game[]>([]);
	let loading = $state(true);
	let loadError = $state('');

	onMount(load);

	async function load() {
		loading = true;
		loadError = '';
		try {
			games = await api<Game[]>('/games');
		} catch (caught) {
			loadError = errorMessage(caught, 'Home could not be loaded.');
		} finally {
			loading = false;
		}
	}
</script>

<AdminHome {games} {loading} {loadError} retry={load} />
