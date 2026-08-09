<script lang="ts">
	import { onMount } from 'svelte';
	import Countdown from '$lib/components/ui/Countdown.svelte';
	import { api } from '$lib/api/client';
	import type { TimerProjection } from '$lib/api/types';
	import { countdownStatus } from './timerPresentation';

	let { gameId, compact = false }: { gameId: string; compact?: boolean } = $props();
	let timer = $state<TimerProjection | null>(null);

	onMount(() => {
		void refresh();
		const visibility = () => {
			if (!document.hidden) void refresh();
		};
		document.addEventListener('visibilitychange', visibility);
		return () => {
			document.removeEventListener('visibilitychange', visibility);
		};
	});

	export async function refresh() {
		timer = await api<TimerProjection | null>(`/games/${gameId}/timer`);
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
