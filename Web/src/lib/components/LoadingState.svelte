<script lang="ts">
	let {
		label,
		progress
	}: {
		label: string;
		progress?: number;
	} = $props();

	const normalizedProgress = $derived(
		progress === undefined ? undefined : Math.min(100, Math.max(0, progress))
	);
</script>

<section class="loading-state" role="status" aria-live="polite">
	<span class="spinner" aria-hidden="true"></span>
	<p>{label}</p>
	{#if normalizedProgress !== undefined}
		<progress aria-label={`${label}: ${normalizedProgress}%`} value={normalizedProgress} max="100">
			{normalizedProgress}%
		</progress>
	{/if}
</section>

<style>
	.loading-state {
		display: grid;
		justify-items: center;
		gap: var(--space-2);
		padding: var(--space-5);
		color: var(--ink-soft);
		text-align: center;
	}

	p {
		margin: 0;
	}

	.spinner {
		width: 1.5rem;
		height: 1.5rem;
		border: 2px solid var(--gold-dark);
		border-right-color: transparent;
		border-radius: 50%;
		animation: spin 700ms linear infinite;
	}

	progress {
		width: min(100%, 18rem);
		accent-color: var(--crimson);
	}

	@keyframes spin {
		to {
			transform: rotate(1turn);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.spinner {
			animation-duration: 1ms;
		}
	}
</style>
