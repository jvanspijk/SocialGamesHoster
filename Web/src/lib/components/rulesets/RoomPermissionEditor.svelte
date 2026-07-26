<script lang="ts">
	import type {
		RulesetPartialRoomPermission,
		RulesetRoomPermission,
		RulesetSenderDisplay
	} from '$lib/api/types';

	let {
		policy,
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

	function state(key: (typeof labels)[number]['key']) {
		const value = policy[key];
		return value === undefined ? 'inherit' : value ? 'yes' : 'no';
	}

	function setState(key: (typeof labels)[number]['key'], value: string) {
		if (value === 'inherit') delete policy[key];
		else policy[key] = value === 'yes';
	}

	function setSender(value: string) {
		if (value === 'inherit') delete policy.senderDisplay;
		else policy.senderDisplay = value as RulesetSenderDisplay;
	}
</script>

<div class="permission">
	{#each labels as option (option.key)}
		<label>
			<span>{option.text}</span>
			{#if partial}
				<select
					value={state(option.key)}
					onchange={(event) => setState(option.key, event.currentTarget.value)}
				>
					<option value="inherit">Use normal setting</option>
					<option value="yes">Yes</option>
					<option value="no">No</option>
				</select>
			{:else}
				<input type="checkbox" bind:checked={policy[option.key]} />
			{/if}
		</label>
	{/each}
	<label>
		<span>Show senders as</span>
		<select
			value={policy.senderDisplay ?? (partial ? 'inherit' : 'profile_name')}
			onchange={(event) => setSender(event.currentTarget.value)}
		>
			{#if partial}<option value="inherit">Use normal setting</option>{/if}
			<option value="profile_name">Profile name</option>
			<option value="game_alias">Game alias</option>
			<option value="seat_number">Seat number</option>
			<option value="role_label">Role name</option>
			<option value="team_label">Team name</option>
		</select>
	</label>
</div>

<style>
	.permission {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.55rem;
	}

	label {
		display: grid;
		align-content: space-between;
		gap: 0.3rem;
		border: 1px solid #c5ad82;
		background: rgb(255 249 230 / 40%);
		padding: 0.55rem;
	}

	label:has(input) {
		grid-template-columns: 1fr auto;
		align-items: center;
	}

	span {
		font-size: 0.82rem;
	}

	select {
		min-height: 40px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.45rem;
	}
</style>
