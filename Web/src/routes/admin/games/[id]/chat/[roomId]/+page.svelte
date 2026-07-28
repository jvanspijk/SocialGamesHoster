<script lang="ts">
	import { page } from '$app/state';
	import AdminChatPage from '$lib/components/AdminChatPage.svelte';
</script>

<AdminChatPage roomId={page.params.roomId} />
