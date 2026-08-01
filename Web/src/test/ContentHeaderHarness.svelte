<script lang="ts">
	import ContentHeader from '$lib/components/ContentHeader.svelte';
</script>

<ContentHeader eyebrow="Ruleset" description="Configure the game." alignment="end">
	{#snippet title()}<h3>Teams</h3>{/snippet}
	{#snippet actions()}<button type="button">Add team</button>{/snippet}
</ContentHeader>
