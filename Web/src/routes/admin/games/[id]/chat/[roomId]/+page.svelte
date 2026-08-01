<script lang="ts">
	import { page } from '$app/state';
	import AdminChatPage from '$lib/features/chat/components/AdminChatPage.svelte';
</script>

<AdminChatPage roomId={page.params.roomId} />
