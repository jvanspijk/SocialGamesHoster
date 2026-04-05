<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { onMount } from 'svelte';
	import TimeInput from './TimeInput.svelte';
	import SecondaryButton from './SecondaryButton.svelte';
	import { StartTimer } from '$lib/client/Timers/StartTimer';
	import { PauseTimer } from '$lib/client/Timers/PauseTimer';
	import { ResumeTimer } from '$lib/client/Timers/ResumeTimer';
	import { StopTimer } from '$lib/client/Timers/StopTimer';
	import { AdjustTimer } from '$lib/client/Timers/AdjustTimer';
	import { invalidateAll } from '$app/navigation';

	interface TimerState {
		isRunning: boolean;
		totalSeconds: number;
		remainingSeconds: number;
	}

	interface Props {
		timer: TimerState | null;
		onClose: () => void;
	}

	let { timer, onClose }: Props = $props();

	// Create timer state
	let createDuration = $state(300); // Default 5 minutes

	// Adjust timer state
	let adjustMode = $state<'add' | 'subtract'>('add');
	let adjustAmount = $state(60); // Default 1 minute

	let isLoading = $state(false);

	onMount(() => {
		const previousOverflow = document.body.style.overflow;
		document.body.style.overflow = 'hidden';

		return () => {
			document.body.style.overflow = previousOverflow;
		};
	});

	function handleBackdropClick(e: MouseEvent): void {
		if (e.target === e.currentTarget) {
			onClose();
		}
	}

	function handleKeydown(e: KeyboardEvent): void {
		if (e.key === 'Escape') {
			onClose();
		}
	}

	async function handleStartTimer() {
		if (createDuration <= 0) return;
		isLoading = true;
		try {
			await StartTimer(fetch, { durationSeconds: createDuration });
			await invalidateAll();
			onClose();
		} finally {
			isLoading = false;
		}
	}

	async function handlePauseResume() {
		if (!timer) return;
		isLoading = true;
		try {
			if (timer.isRunning) {
				await PauseTimer(fetch);
			} else {
				await ResumeTimer(fetch);
			}
			await invalidateAll();
		} finally {
			isLoading = false;
		}
	}

	async function handleFinish() {
		isLoading = true;
		try {
			await StopTimer(fetch);
			await invalidateAll();
			onClose();
		} finally {
			isLoading = false;
		}
	}

	async function handleAdjust() {
		if (adjustAmount <= 0) return;
		isLoading = true;
		try {
			const delta = adjustMode === 'add' ? adjustAmount : -adjustAmount;
			await AdjustTimer(fetch, { deltaSeconds: delta });
			await invalidateAll();
		} finally {
			isLoading = false;
		}
	}

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
	}
</script>

<div
	class="modal-overlay"
	role="dialog"
	aria-modal="true"
	aria-label={timer ? 'Timer Controls' : 'Create Timer'}
	tabindex="0"
	onclick={handleBackdropClick}
	onkeydown={handleKeydown}
	transition:fade={{ duration: 200 }}
>
	<div class="modal-inner" transition:scale={{ duration: 200, start: 0.95 }}>
		{#if timer === null}
			<!-- Create Timer View -->
			<h2>Create Timer</h2>
			<div class="modal-body">
				<label class="input-group">
					<span class="label-text">Duration</span>
					<TimeInput bind:value={createDuration} placeholder="05:00" />
				</label>
			</div>
			<div class="actions">
				<SecondaryButton onclick={onClose} enabled={!isLoading}>Cancel</SecondaryButton>
				<SecondaryButton
					variant="primary"
					onclick={handleStartTimer}
					enabled={!isLoading && createDuration > 0}
				>
					{isLoading ? 'Starting...' : 'Start Timer'}
				</SecondaryButton>
			</div>
		{:else}
			<!-- Timer Controls View -->
			<h2>Timer Controls</h2>
			<div class="modal-body">
				<div class="current-time">
					<span class="time-label">Remaining</span>
					<span class="time-value" class:paused={!timer.isRunning}>
						{formatTime(timer.remainingSeconds)}
					</span>
				</div>

				<div class="control-section">
					<SecondaryButton
						variant="primary"
						onclick={handlePauseResume}
						enabled={!isLoading && timer.remainingSeconds > 0}
					>
						{#if timer.isRunning}
							<span class="btn-icon">⏸</span> Pause
						{:else}
							<span class="btn-icon">▶</span> Resume
						{/if}
					</SecondaryButton>
				</div>

				<div class="control-section adjust-section">
					<span class="section-label">Adjust Time</span>
					<div class="adjust-controls">
						<div class="mode-toggle">
							<button
								type="button"
								class="mode-btn"
								class:active={adjustMode === 'add'}
								onclick={() => (adjustMode = 'add')}
								disabled={isLoading}
							>
								+
							</button>
							<button
								type="button"
								class="mode-btn"
								class:active={adjustMode === 'subtract'}
								onclick={() => (adjustMode = 'subtract')}
								disabled={isLoading}
							>
								-
							</button>
						</div>
						<TimeInput bind:value={adjustAmount} placeholder="01:00" disabled={isLoading} />
						<SecondaryButton onclick={handleAdjust} enabled={!isLoading && adjustAmount > 0}>
							Apply
						</SecondaryButton>
					</div>
				</div>

				<div class="control-section">
					<SecondaryButton variant="danger" onclick={handleFinish} enabled={!isLoading}>
						Finish Timer
					</SecondaryButton>
				</div>
			</div>
			<div class="actions actions-single">
				<SecondaryButton onclick={onClose} enabled={!isLoading}>Close</SecondaryButton>
			</div>
		{/if}
	</div>
</div>

<style>
	.modal-overlay {
		position: fixed;
		inset: 0;
		z-index: 1100;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
		background: rgba(34, 27, 22, 0.5);
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
	}

	.modal-inner {
		background: var(--color-surface-soft);
		color: var(--color-text);
		padding: 1.5rem;
		border-radius: 3px;
		border: 2px solid var(--color-border);
		box-shadow:
			0 8px 18px rgba(0, 0, 0, 0.45),
			inset 0 0 24px rgba(91, 74, 60, 0.15),
			2px 2px 0 rgba(91, 74, 60, 0.35);
		width: 90vw;
		max-width: 420px;
	}

	h2 {
		margin: 0 0 1rem 0;
		font-size: 1.35rem;
		font-family: var(--font-heading);
		color: var(--color-border);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		text-align: center;
		border-bottom: 1px solid var(--color-border);
		padding-bottom: 0.5rem;
	}

	.modal-body {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
		margin-bottom: 1.5rem;
	}

	.input-group {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
	}

	.label-text {
		font-family: var(--font-body);
		color: var(--color-text-muted);
		font-size: 0.95rem;
	}

	.current-time {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
		padding: 0.75rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border-soft);
		border-radius: 3px;
	}

	.time-label {
		font-family: var(--font-body);
		color: var(--color-text-muted);
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.time-value {
		font-family: var(--font-heading);
		font-size: 2rem;
		font-weight: 700;
		color: var(--color-text);
		letter-spacing: 0.1em;
	}

	.time-value.paused {
		color: var(--color-text-muted);
	}

	.control-section {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
	}

	.section-label {
		font-family: var(--font-heading);
		font-size: 0.8rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--color-text-muted);
	}

	.adjust-section {
		padding: 0.75rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border-soft);
		border-radius: 3px;
	}

	.adjust-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
		justify-content: center;
	}

	.mode-toggle {
		display: flex;
		border: 1px solid var(--color-border);
		border-radius: 3px;
		overflow: hidden;
	}

	.mode-btn {
		font-family: var(--font-heading);
		font-size: 1.25rem;
		font-weight: bold;
		width: 2.5rem;
		height: 2.5rem;
		border: none;
		background: var(--color-surface);
		color: var(--color-text-muted);
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.mode-btn:first-child {
		border-right: 1px solid var(--color-border);
	}

	.mode-btn:hover:not(:disabled) {
		background: var(--color-surface-alt);
	}

	.mode-btn.active {
		background: var(--color-border);
		color: var(--color-surface);
	}

	.mode-btn:disabled {
		cursor: not-allowed;
		opacity: 0.5;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
		justify-content: center;
	}

	.actions-single {
		justify-content: center;
	}

	.btn-icon {
		margin-right: 0.25rem;
	}

	@media (max-width: 480px) {
		.modal-inner {
			padding: 1.25rem;
		}

		.adjust-controls {
			flex-direction: column;
		}

		.time-value {
			font-size: 1.75rem;
		}
	}
</style>
