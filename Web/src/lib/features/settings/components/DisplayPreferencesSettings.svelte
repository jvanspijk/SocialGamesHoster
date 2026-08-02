<script lang="ts">
	import { displayPreferences } from '$lib/state/display.svelte';
	import DisplaySettings from './DisplaySettings.svelte';
</script>

<DisplaySettings
	largeText={displayPreferences.largeText}
	highContrast={displayPreferences.highContrast}
	onLargeTextChange={displayPreferences.setLargeText}
	onHighContrastChange={displayPreferences.setHighContrast}
/>
