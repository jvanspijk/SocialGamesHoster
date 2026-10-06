<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import Panel from '$lib/components/Panel.svelte';

	let version = $state('');

	onMount(async () => {
		try {
			const status = await api<{ version: string }>('/setup/status');
			const release = status.version.match(/^(\d{4}\.\d{2}\.\d{2})-(.+)$/);
			version = release ? `${release[1]} (${release[2]})` : status.version;
		} catch {
			version = '';
		}
	});
</script>

{#if version}
	<Panel title="App version">
		<p class="version-number">{version}</p>
	</Panel>
{/if}

<style>
	.version-number {
		font-family: monospace;
	}
</style>
