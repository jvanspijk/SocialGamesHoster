<script lang="ts">
	import Sheet from '$lib/components/Sheet.svelte';

	let open = $state(false);
</script>

<button type="button" onclick={() => (open = true)}>Open settings</button>
<Sheet {open} title="Display settings" close={() => (open = false)}>
	<label>Example setting <input type="checkbox" /></label>
</Sheet>
