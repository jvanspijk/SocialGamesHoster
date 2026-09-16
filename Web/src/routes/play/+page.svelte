<script lang="ts">
	import { getContext } from 'svelte';
	import PlayerHome from '$lib/features/play/components/PlayerHome.svelte';
	import {
		playerShellContextKey,
		type PlayerShellContext
	} from '$lib/features/play/playerShellContext';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';

	const shell = getContext<PlayerShellContext>(playerShellContextKey);
</script>

<PlayerHome
	displayName={auth.actor?.displayName ?? 'player'}
	loading={shell.loading}
	view={gameState.player}
	availableLobby={shell.availableLobby}
	liveGame={shell.liveGame}
	joiningLobby={shell.joiningLobby}
	loadError={shell.loadError}
	join={shell.joinAvailableLobby}
	retry={shell.refresh}
/>
