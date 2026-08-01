<script lang="ts">
	import Dialog from '$lib/components/Dialog.svelte';

	let open = $state(false);
	let chosen = $state('');
	let updates = $state(0);
	let closeCount = $state(0);

	function close() {
		closeCount += 1;
		open = false;
	}
</script>

<button type="button" onclick={() => (open = true)}>Open recipients</button>
<output data-testid="close-count">{closeCount}</output>
<Dialog
	{open}
	title="New message"
	description="Choose a player to start a direct conversation."
	{close}
>
	<div class="player-list">
		<button type="button" onclick={() => (chosen = 'Rowan')}>Rowan<small>Seat 2</small></button>
	</div>
	<p data-testid="chosen">{chosen}</p>
	<p data-testid="updates">{updates}</p>
	<button type="button" onclick={() => (updates += 1)}>Update dialog content</button>
	<button type="button" onclick={() => (open = false)}>Close externally</button>
</Dialog>
