<script lang="ts">
	import StatusBadge from './StatusBadge.svelte';
</script>

{#snippet icon()}
	<svg data-testid="connection-icon"></svg>
{/snippet}

<StatusBadge label="Connected" dot={true} {icon} />
