<script lang="ts">
	import RoomPermissionEditor from '$lib/components/rulesets/RoomPermissionEditor.svelte';
	import type { RulesetPartialRoomPermission } from '$lib/api/types';

	let policy = $state<RulesetPartialRoomPermission>({});
</script>

<RoomPermissionEditor {policy} partial />
<output aria-label="Visibility result">
	{policy.visible === undefined ? 'inherit' : policy.visible ? 'yes' : 'no'}
</output>
