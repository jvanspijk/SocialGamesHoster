<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import LogOut from '@lucide/svelte/icons/log-out';
	import IconButton from '$lib/components/IconButton.svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';

	async function signOut() {
		try {
			await api('/auth/logout', { method: 'POST' });
		} finally {
			gameState.clear();
			auth.clear();
			await goto(resolve('/'));
		}
	}
</script>

<IconButton label="Sign out" variant="ghost" onclick={signOut}>
	{#snippet icon()}<LogOut size={20} />{/snippet}
</IconButton>
