<script lang="ts">
	import Dialog from '$lib/components/Dialog.svelte';

	let open = $state(false);
	let chosen = $state('');
</script>

<button type="button" onclick={() => (open = true)}>Open recipients</button>
<Dialog
	{open}
	title="New message"
	description="Choose a player to start a direct conversation."
	close={() => (open = false)}
>
	<div class="player-list">
		<button type="button" onclick={() => (chosen = 'Rowan')}>Rowan<small>Seat 2</small></button>
	</div>
	<p data-testid="chosen">{chosen}</p>
</Dialog>
