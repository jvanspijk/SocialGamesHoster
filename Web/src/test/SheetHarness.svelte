<script lang="ts">
	import Sheet from '$lib/components/Sheet.svelte';

	let open = $state(false);
	let updates = $state(0);
	let closeCount = $state(0);

	function close() {
		closeCount += 1;
		open = false;
	}
</script>

<button type="button" onclick={() => (open = true)}>Open settings</button>
<output data-testid="close-count">{closeCount}</output>
<Sheet
	{open}
	title="Display settings"
	description="Choose how the host looks on this device."
	{close}
>
	<label>Example setting <input type="checkbox" /></label>
	<p data-testid="updates">{updates}</p>
	<button type="button" onclick={() => (updates += 1)}>Update sheet content</button>
	<button type="button" onclick={() => (open = false)}>Close externally</button>
	{#snippet actions()}<button type="button">Save display settings</button>{/snippet}
</Sheet>
