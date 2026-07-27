<script lang="ts">
	import { displayPreferences } from '$lib/state/display.svelte';
</script>

<div class="display-settings stack">
	<p>These settings apply only on this device.</p>
	<label>
		<input
			type="checkbox"
			checked={displayPreferences.largeText}
			onchange={(event) =>
				displayPreferences.setLargeText((event.currentTarget as HTMLInputElement).checked)}
		/>
		<span>
			<strong>Large text</strong>
			<small>Increase the base text size while keeping controls visible.</small>
		</span>
	</label>
	<label>
		<input
			type="checkbox"
			checked={displayPreferences.highContrast}
			onchange={(event) =>
				displayPreferences.setHighContrast((event.currentTarget as HTMLInputElement).checked)}
		/>
		<span>
			<strong>High contrast</strong>
			<small>Remove page texture and strengthen text and control outlines.</small>
		</span>
	</label>
</div>

<style>
	label {
		display: grid;
		grid-template-columns: var(--target-size) minmax(0, 1fr);
		align-items: start;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding-block: var(--space-3);
		cursor: pointer;
	}

	input {
		width: 1.35rem;
		height: 1.35rem;
		margin: 0.15rem auto;
		accent-color: var(--crimson);
	}

	span {
		display: grid;
		gap: var(--space-1);
	}

	small {
		color: var(--ink-soft);
		line-height: 1.4;
	}
</style>
