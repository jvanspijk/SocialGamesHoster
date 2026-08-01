<script lang="ts">
	import Panel from '$lib/components/Panel.svelte';
</script>

<Panel title="Players" description="People approved for this game.">
	{#snippet actions()}<button type="button">Invite player</button>{/snippet}
	<p>Rowan</p>
</Panel>
