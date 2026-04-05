<script lang="ts">
	export let label;
	export let variant = 'primary';
</script>

<span class="badge {variant}">
	{label}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		padding: 0.125rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		border: 1px solid transparent;
	}

	.primary {
		background-color: var(--color-surface-soft);
		color: var(--color-border);
		border-color: var(--color-border-soft);
	}

	.secondary {
		background-color: var(--color-surface-alt);
		color: var(--color-text);
	}

	.outline {
		background-color: var(--color-on-accent);
		border-color: var(--color-border-soft);
		color: var(--color-text-muted);
	}
</style>
