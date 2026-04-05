<script lang="ts">
	let { title, content } = $props();
</script>

<div class="detail-card">
	<h2>{title}</h2>
	<p>{content}</p>
</div>

<style>
	.detail-card {
		background-color: var(--color-surface-alt);
		border: 1px solid var(--color-border-soft);
		padding: 20px;
		box-shadow: 4px 4px 10px rgba(0, 0, 0, 0.1);
		border-radius: 4px;
	}

	.detail-card h2 {
		font-family: var(--font-heading);
		color: var(--color-accent-strong);
		margin-top: 0;
		margin-bottom: 10px;
		border-bottom: 2px solid var(--color-accent-strong);
		padding-bottom: 5px;
		font-size: 1.25em;
	}
</style>
