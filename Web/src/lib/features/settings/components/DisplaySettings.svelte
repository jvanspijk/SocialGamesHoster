<script lang="ts">
	import { displayPreferences } from '$lib/state/display.svelte';
	import ToggleSetting from '$lib/components/ToggleSetting.svelte';
</script>

<div class="display-settings stack">
	<p>These settings apply only on this device.</p>
	<ToggleSetting
		title="Large text"
		description="Increase the base text size while keeping controls visible."
		name="large-text"
		checked={displayPreferences.largeText}
		onchange={displayPreferences.setLargeText}
	/>
	<ToggleSetting
		title="High contrast"
		description="Remove page texture and strengthen text and control outlines."
		name="high-contrast"
		checked={displayPreferences.highContrast}
		onchange={displayPreferences.setHighContrast}
	/>
</div>

<style>
</style>
