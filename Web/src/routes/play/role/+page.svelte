<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import RoleReveal from '$lib/features/play/components/RoleReveal.svelte';
	import { gameState } from '$lib/state/game.svelte';

	let revealed = $state(false);
	let revision = $state(0);
	const view = $derived(gameState.player);

	onMount(() => {
		const conceal = () => {
			if (document.hidden) revealed = false;
		};
		document.addEventListener('visibilitychange', conceal);
		return () => document.removeEventListener('visibilitychange', conceal);
	});

	$effect(() => {
		if (!view?.roleAvailable || view.roleRevision !== revision) {
			revealed = false;
			revision = view?.roleRevision ?? 0;
		}
	});
</script>

{#if view}
	<RoleReveal
		{view}
		{revealed}
		reveal={() => (revealed = true)}
		hide={() => (revealed = false)}
		back={() => goto(resolve('/play'))}
	/>
{/if}
