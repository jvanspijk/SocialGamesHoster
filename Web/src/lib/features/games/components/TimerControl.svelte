<script lang="ts">
	import { tick } from 'svelte';
	import CircleStop from '@lucide/svelte/icons/circle-stop';
	import Pause from '@lucide/svelte/icons/pause';
	import Play from '@lucide/svelte/icons/play';
	import Plus from '@lucide/svelte/icons/plus';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import Button from '$lib/components/Button.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import Countdown from '$lib/components/ui/Countdown.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { TimerProjection } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';
	import { countdownStatus, timerStatusLabel } from './timerPresentation';

	let {
		gameId,
		timer,
		onchange
	}: {
		gameId: string;
		timer: TimerProjection;
		onchange: (timer: TimerProjection) => void;
	} = $props();

	let durationMinutes = $state(5);
	let busy = $state(false);
	const durationOptions = [
		{ value: 1, label: '1 min' },
		{ value: 3, label: '3 min' },
		{ value: 5, label: '5 min' },
		{ value: 10, label: '10 min' },
		{ value: 15, label: '15 min' },
		{ value: 30, label: '30 min' }
	];

	async function primaryAction(event: MouseEvent) {
		const button = event.currentTarget as HTMLButtonElement;
		const restoreFocus = document.activeElement === button;
		const path =
			timer.status === 'running' ? 'pause' : timer.status === 'paused' ? 'resume' : 'start';
		await command(
			path,
			path === 'start'
				? {
						durationMs:
							timer.status === 'completed'
								? Math.max(timer.totalMs, durationMinutes * 60_000)
								: durationMinutes * 60_000
					}
				: {}
		);
		await tick();
		if (restoreFocus && document.activeElement === document.body) {
			button.focus({ preventScroll: true });
		}
	}

	async function command(
		path: 'start' | 'pause' | 'resume' | 'adjust' | 'stop',
		body: unknown = {}
	) {
		busy = true;
		try {
			const updated = await api<TimerProjection>(`/games/${gameId}/timer/${path}`, {
				method: 'POST',
				...jsonBody(body)
			});
			onchange(updated);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The timer could not be updated.'));
		} finally {
			busy = false;
		}
	}
</script>

<section aria-label="Game timer">
	<div class="timer-readout">
		<Countdown
			variant="readout"
			status={countdownStatus(timer.status)}
			statusLabel={timerStatusLabel(timer.status)}
			remainingMs={timer.remainingMs}
			endsAt={timer.endsAt}
		/>
	</div>

	<div class="timer-controls">
		<div class="timer-settings">
			<div
				class:concealed={timer.status !== 'inactive'}
				inert={timer.status !== 'inactive'}
				aria-hidden={timer.status !== 'inactive'}
			>
				<SelectField
					layout="inline"
					density="compact"
					label="Duration"
					name="timer-duration"
					bind:value={durationMinutes}
					options={durationOptions}
					disabled={busy || timer.status !== 'inactive'}
				/>
			</div>
			<div
				class="timer-adjustments"
				class:concealed={timer.status === 'inactive'}
				inert={timer.status === 'inactive'}
				aria-hidden={timer.status === 'inactive'}
			>
				<Button
					density="compact"
					variant="secondary"
					disabled={busy || !['running', 'paused'].includes(timer.status)}
					onclick={() => command('adjust', { deltaMs: 60_000 })}
				>
					<Plus size={16} /> 1 min
				</Button>
				<Button
					density="compact"
					variant="ghost"
					disabled={busy || timer.status === 'inactive'}
					onclick={() => command('stop')}><CircleStop size={16} /> Clear</Button
				>
			</div>
		</div>

		<div class="timer-primary">
			<Button density="compact" loading={busy} onclick={primaryAction}>
				{#if timer.status === 'running'}
					{#if !busy}<Pause size={16} />{/if} Pause timer
				{:else if timer.status === 'paused'}
					{#if !busy}<Play size={16} />{/if} Resume timer
				{:else if timer.status === 'completed'}
					{#if !busy}<RotateCcw size={16} />{/if} Start again
				{:else}
					{#if !busy}<Play size={16} />{/if} Start timer
				{/if}
			</Button>
		</div>
	</div>
</section>

<style>
	section {
		display: grid;
		gap: var(--space-2);
	}

	.timer-readout {
		text-align: center;
	}

	.timer-controls {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		align-items: end;
		gap: var(--space-2);
	}

	.timer-settings {
		grid-row: 2;
		display: grid;
		min-width: 0;
		align-items: end;
	}

	.timer-settings > div {
		grid-area: 1 / 1;
		min-width: 0;
	}

	.concealed {
		visibility: hidden;
	}

	.timer-adjustments,
	.timer-primary {
		display: grid;
		gap: var(--space-2);
	}
	.timer-adjustments {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.timer-primary {
		grid-row: 1;
	}

	@container timer-card (min-width: 24rem) {
		.timer-controls {
			grid-template-columns: minmax(0, 2fr) minmax(6.75rem, 1fr);
		}
		.timer-settings {
			grid-column: 1;
			grid-row: 1;
		}
		.timer-primary {
			grid-column: 2;
		}
	}
</style>
