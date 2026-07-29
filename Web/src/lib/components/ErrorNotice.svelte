<script lang="ts">
	import type { FormError } from '$lib/forms/errors';

	let { error }: { error: FormError | null } = $props();
</script>

{#if error}
	<div class="error" role="alert" aria-live="assertive">
		<strong>{error.message}</strong>
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
