<script lang="ts">
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { connection } from '$lib/state/connection.svelte';

	const presentation = $derived.by(() => {
		switch (connection.state) {
			case 'offline':
				return { label: 'offline', tone: 'danger' as const };
			case 'reconnecting':
				return { label: 'reconnecting', tone: 'warning' as const };
			default:
				return { label: 'connected', tone: 'success' as const };
		}
	});
</script>

<span class="connection-label"><StatusBadge {...presentation} /></span>
