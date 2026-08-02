<script lang="ts">
	import Alert from './Alert.svelte';

	let { message, traceId }: { message?: string; traceId?: string } = $props();
</script>

{#if message}
	<section class="form-error">
		<Alert tone="error" {message} />
		{#if traceId}
			<details>
				<summary>Technical details</summary>
				Trace ID: <code>{traceId}</code>
			</details>
		{/if}
	</section>
{/if}

<style>
	details {
		margin: var(--space-2) var(--space-3) 0;
		font-size: 0.8rem;
	}
</style>
