<script lang="ts">
	import PageHeading from '$lib/components/PageHeading.svelte';
</script>

<PageHeading eyebrow="Account" title="Profile" description="Manage your player identity.">
	{#snippet actions()}
		<button type="button">Open settings</button>
	{/snippet}
</PageHeading>
