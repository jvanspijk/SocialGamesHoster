<script lang="ts">
	import { onMount } from 'svelte';
	import { Pause, Play, TimerReset } from '@lucide/svelte';
	import { api } from '$lib/api/client';
	import type { TimerProjection } from '$lib/api/types';

	let { gameId, compact = false }: { gameId: string; compact?: boolean } = $props();
	let timer = $state<TimerProjection | null>(null);
	let now = $state(Date.now());

	onMount(() => {
		void refresh();
		const interval = window.setInterval(() => (now = Date.now()), 250);
		const visibility = () => {
			if (!document.hidden) void refresh();
		};
		document.addEventListener('visibilitychange', visibility);
		return () => {
			window.clearInterval(interval);
			document.removeEventListener('visibilitychange', visibility);
		};
	});

	let remaining = $derived.by(() => {
		if (!timer) return 0;
		if (timer.status === 'running' && timer.endsAt) {
			return Math.max(0, new Date(timer.endsAt).getTime() - now);
		}
		return timer.remainingMs;
	});
	let display = $derived(
		`${Math.floor(remaining / 60_000)
			.toString()
			.padStart(2, '0')}:${Math.floor((remaining % 60_000) / 1000)
			.toString()
			.padStart(2, '0')}`
	);

	export async function refresh() {
		timer = await api<TimerProjection | null>(`/games/${gameId}/timer`);
	}
</script>

{#if timer}
	<div class:compact class:completed={timer.status === 'completed'} class="timer">
		<span class="icon" aria-hidden="true">
			{#if timer.status === 'running'}<Play size={16} />{:else if timer.status === 'paused'}<Pause
					size={16}
				/>{:else}<TimerReset size={16} />{/if}
		</span>
		<strong aria-label={`${Math.ceil(remaining / 1000)} seconds remaining`}>{display}</strong>
		<small>{timer.status}</small>
	</div>
{/if}

<style>
	.timer {
		display: inline-grid;
		grid-template-columns: auto auto;
		align-items: center;
		gap: 0.1rem 0.5rem;
		border: 1px solid #9a7e51;
		background: var(--paper-light);
		padding: 0.55rem 0.8rem;
	}

	.icon {
		grid-row: span 2;
		color: var(--crimson);
	}

	strong {
		font-family: var(--font-display);
		font-size: 1.25rem;
		letter-spacing: 0.08em;
		line-height: 1;
	}

	small {
		color: var(--ink-soft);
		font-size: 0.7rem;
		text-transform: capitalize;
	}

	.compact {
		border: 0;
		background: transparent;
		padding: 0;
	}

	.completed {
		color: var(--danger);
	}
</style>
