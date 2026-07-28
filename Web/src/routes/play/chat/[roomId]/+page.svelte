<script lang="ts">
	import { page } from '$app/state';
	import PlayerChatPage from '$lib/components/PlayerChatPage.svelte';
</script>

<PlayerChatPage roomId={page.params.roomId} />
