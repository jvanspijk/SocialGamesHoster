<script lang="ts">
	let {
		position,
		total
	}: {
		position: number;
		total: number;
	} = $props();
</script>

<p class="queue">{position} of {total}</p>

<style>
	.queue {
		margin: 0;
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.09em;
		text-transform: uppercase;
	}
</style>
