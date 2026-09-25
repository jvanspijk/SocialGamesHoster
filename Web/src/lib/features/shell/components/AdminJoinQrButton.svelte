<script lang="ts">
	import { QrCode } from '@lucide/svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import IconButton from '$lib/components/IconButton.svelte';

	let open = $state(false);
</script>

<IconButton label="Show player join QR code" variant="ghost" onclick={() => (open = true)}>
	{#snippet icon()}<QrCode size={20} />{/snippet}
</IconButton>

<Dialog
	{open}
	title="Player join QR code"
	description="Players must be on the same local network to join."
	close={() => (open = false)}
>
	<div class="qr-code">
		<img src="/api/app/v1/setup/join-qr" alt="QR code for the player join page" />
	</div>
</Dialog>

<style>
	.qr-code {
		display: grid;
		place-items: center;
	}

	.qr-code img {
		display: block;
		width: min(100%, 20rem);
		height: auto;
	}
</style>
