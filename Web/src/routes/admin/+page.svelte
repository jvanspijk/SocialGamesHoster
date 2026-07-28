<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { api } from '$lib/api/client';
	import type { Game } from '$lib/api/types';

	onMount(async () => {
		const games = await api<Game[]>('/games');
		const active = games.find((game) =>
			['lobby', 'running', 'paused', 'review'].includes(game.status)
		);
		await goto(resolve(active ? `/admin/games/${active.id}/overview` : '/admin/games'), {
			replaceState: true
		});
	});
</script>

<p class="loading" role="status">Opening games…</p>

<style>
	.loading {
		padding: var(--space-6);
	}
</style>
