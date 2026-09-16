<script lang="ts">
	import { onMount } from 'svelte';
	import Countdown from '$lib/components/ui/Countdown.svelte';
	import { api } from '$lib/api/client';
	import type { TimerProjection } from '$lib/api/types';
	import { countdownStatus } from './timerPresentation';

	let {
		gameId,
		revision,
		compact = false
	}: { gameId: string; revision?: number; compact?: boolean } = $props();
	let timer = $state<TimerProjection | null>(null);
	let refreshRequest = 0;

	$effect(() => {
		// A player-view realtime update changes this revision. Re-fetch the timer
		// because it is projected independently from the player game view.
		void revision;
		void refresh();
	});

	onMount(() => {
		const visibility = () => {
			if (!document.hidden) void refresh();
		};
		document.addEventListener('visibilitychange', visibility);
		return () => document.removeEventListener('visibilitychange', visibility);
	});

	export async function refresh() {
		const request = ++refreshRequest;
		const requestedGameID = gameId;
		const current = await api<TimerProjection | null>(`/games/${requestedGameID}/timer`);
		if (request === refreshRequest && requestedGameID === gameId) timer = current;
	}
</script>

{#if timer}
	<Countdown
		status={countdownStatus(timer.status)}
		statusLabel={timer.status}
		remainingMs={timer.remainingMs}
		endsAt={timer.endsAt}
		{compact}
	/>
{/if}
