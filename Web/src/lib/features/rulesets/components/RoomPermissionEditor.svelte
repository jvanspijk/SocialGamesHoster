<script lang="ts">
	import type {
		RulesetPartialRoomPermission,
		RulesetRoomPermission,
		RulesetSenderDisplay
	} from '$lib/api/types';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import SelectField, { type SelectOption } from '$lib/components/SelectField.svelte';

	let {
		policy = $bindable(),
		partial = false
	}: {
		policy: RulesetRoomPermission | RulesetPartialRoomPermission;
		partial?: boolean;
	} = $props();

	const labels: Array<{
		key: 'visible' | 'readable' | 'sendable' | 'gameMasterMaySend';
		text: string;
	}> = [
		{ key: 'visible', text: 'Players can see the room' },
		{ key: 'readable', text: 'Players can read messages' },
		{ key: 'sendable', text: 'Players can send messages' },
		{ key: 'gameMasterMaySend', text: 'Game masters can send messages' }
	];
	const inheritedOptions: SelectOption<'inherit' | 'yes' | 'no'>[] = [
		{ value: 'inherit', label: 'Use normal setting' },
		{ value: 'yes', label: 'Yes' },
		{ value: 'no', label: 'No' }
	];
	const senderOptions: SelectOption<RulesetSenderDisplay | 'inherit'>[] = [
		{ value: 'inherit', label: 'Use normal setting' },
		{ value: 'profile_name', label: 'Profile name' },
		{ value: 'game_alias', label: 'Game alias' },
		{ value: 'seat_number', label: 'Seat number' },
		{ value: 'role_label', label: 'Role name' },
		{ value: 'team_label', label: 'Team name' }
	];

	function state(key: (typeof labels)[number]['key']) {
		const value = policy[key];
		return value === undefined ? 'inherit' : value ? 'yes' : 'no';
	}

	function setState(key: (typeof labels)[number]['key'], value: 'inherit' | 'yes' | 'no') {
		if (value === 'inherit') delete policy[key];
		else policy[key] = value === 'yes';
	}

	function setSender(value: RulesetSenderDisplay | 'inherit') {
		if (value === 'inherit') delete policy.senderDisplay;
		else policy.senderDisplay = value as RulesetSenderDisplay;
	}
</script>

<div class="permission">
	{#each labels as option (option.key)}
		{#if partial}
			<SelectField
				label={option.text}
				name={`permission-${option.key}`}
				value={state(option.key)}
				options={inheritedOptions}
				onchange={(value) => setState(option.key, value)}
			/>
		{:else}
			<CheckboxField
				label={option.text}
				name={`permission-${option.key}`}
				checked={policy[option.key] ?? false}
				onchange={(checked) => (policy[option.key] = checked)}
			/>
		{/if}
	{/each}
	<SelectField
		label="Show senders as"
		name="permission-sender-display"
		value={policy.senderDisplay ?? (partial ? 'inherit' : 'profile_name')}
		options={partial ? senderOptions : senderOptions.slice(1)}
		onchange={setSender}
	/>
</div>

<style>
	.permission {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.55rem;
	}

	:global(label) {
		background: rgb(255 249 230 / 40%);
	}
</style>
