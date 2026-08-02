<script lang="ts">
	let { id, label = 'New messages' }: { id: string; label?: string } = $props();
</script>

<div class="unread-divider" {id} role="separator" aria-label={label}>
	<span>{label}</span>
</div>

<style>
	.unread-divider {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-align: center;
	}

	.unread-divider::before,
	.unread-divider::after {
		height: 1px;
		flex: 1;
		background: var(--crimson);
		content: '';
	}
</style>
