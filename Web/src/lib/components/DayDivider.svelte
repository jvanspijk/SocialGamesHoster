<script lang="ts">
	let { label }: { label: string } = $props();
</script>

<div class="day-divider" role="separator" aria-label={label}><span>{label}</span></div>

<style>
	.day-divider {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--ink-soft);
		font-size: 0.72rem;
		text-align: center;
	}

	.day-divider::before,
	.day-divider::after {
		height: 1px;
		flex: 1;
		background: #b9a170;
		content: '';
	}
</style>
