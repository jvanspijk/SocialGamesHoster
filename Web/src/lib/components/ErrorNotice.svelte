<script lang="ts">
	import type { AppErrorBody } from '$lib/api/types';

	let { error }: { error: AppErrorBody | null } = $props();
	const hasFieldErrors = $derived(Object.keys(error?.fieldErrors ?? {}).length > 0);
</script>

{#if error}
	<div class="error" role="alert" aria-live="assertive">
		<strong>{hasFieldErrors ? 'Please correct the highlighted details.' : error.message}</strong>
		{#if error.traceId}
			<details>
				<summary>Technical details</summary>
				Trace ID: <code>{error.traceId}</code>
			</details>
		{/if}
	</div>
{/if}

<style>
	.error {
		border-inline-start: 4px solid var(--danger);
		background: rgb(140 42 62 / 10%);
		padding: 0.75rem 0.9rem;
	}

	details {
		margin-top: 0.35rem;
		font-size: 0.8rem;
	}
</style>
