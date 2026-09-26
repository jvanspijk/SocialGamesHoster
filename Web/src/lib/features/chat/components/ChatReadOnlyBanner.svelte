<script lang="ts">
	import { Shield } from '@lucide/svelte';

	let { message }: { message: string } = $props();
</script>

<div class="read-only">
	<Shield size={17} />
	{message}
</div>

<style>
	.read-only {
		display: flex;
		min-height: 3.5rem;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		background: var(--surface-muted);
		color: var(--text-secondary);
	}
</style>
