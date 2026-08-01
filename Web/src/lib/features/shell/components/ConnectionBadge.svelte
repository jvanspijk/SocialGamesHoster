<script lang="ts">
	import { connection } from '$lib/state/connection.svelte';
</script>

<span
	class:offline={connection.state === 'offline'}
	class:waiting={connection.state === 'reconnecting'}
>
	<i aria-hidden="true"></i>
	{connection.state}
</span>

<style>
	span {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		color: var(--success);
		font-family: var(--font-display);
		font-size: 0.65rem;
		letter-spacing: 0.09em;
		text-transform: uppercase;
	}

	i {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: currentColor;
	}

	.waiting {
		color: var(--gold);
	}

	.offline {
		color: var(--danger);
	}
</style>
