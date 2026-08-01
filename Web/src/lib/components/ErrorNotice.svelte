<script lang="ts">
	import Alert from './Alert.svelte';
	import type { FormError } from '$lib/forms/errors';

	let { error }: { error: FormError | null } = $props();
</script>

{#if error}
	<section class="form-error">
		<Alert tone="error" message={error.message} />
		{#if error.traceId}
			<details>
				<summary>Technical details</summary>
				Trace ID: <code>{error.traceId}</code>
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
