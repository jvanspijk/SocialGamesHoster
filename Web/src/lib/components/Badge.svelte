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
		background-color: #eef2ff;
		color: #4338ca;
		border-color: #e0e7ff;
	}

	.secondary {
		background-color: #f1f5f9;
		color: #475569;
	}

	.outline {
		background-color: white;
		border-color: #e2e8f0;
		color: #64748b;
	}
</style>