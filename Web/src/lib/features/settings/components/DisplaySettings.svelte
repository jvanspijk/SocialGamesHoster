<script lang="ts">
	import ToggleSetting from '$lib/components/ToggleSetting.svelte';

	let {
		largeText,
		highContrast,
		onLargeTextChange,
		onHighContrastChange
	}: {
		largeText: boolean;
		highContrast: boolean;
		onLargeTextChange: (checked: boolean) => void;
		onHighContrastChange: (checked: boolean) => void;
	} = $props();
</script>

<div class="display-settings stack">
	<p>These settings apply only on this device.</p>
	<ToggleSetting
		title="Large text"
		description="Increase the base text size while keeping controls visible."
		name="large-text"
		checked={largeText}
		onchange={onLargeTextChange}
	/>
	<ToggleSetting
		title="High contrast"
		description="Remove page texture and strengthen text and control outlines."
		name="high-contrast"
		checked={highContrast}
		onchange={onHighContrastChange}
	/>
</div>
