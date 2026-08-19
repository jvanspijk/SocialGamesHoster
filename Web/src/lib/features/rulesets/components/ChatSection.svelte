<script lang="ts">
	import type {
		RulesetChatChannel,
		RulesetDefinition,
		RulesetPartialRoomPermission,
		RulesetRoomPermission
	} from '$lib/api/types';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField, { type SelectOption } from '$lib/components/SelectField.svelte';
	import CollectionEditor from './CollectionEditor.svelte';
	import { duplicateByID, moveByID, nextID } from './definition-editor';
	import RoomPermissionEditor from './RoomPermissionEditor.svelte';

	let {
		definition = $bindable(),
		selectedItems
	}: { definition: RulesetDefinition; selectedItems: Record<string, string> } = $props();
	const channelEntries = $derived(
		definition.chat.channels.map((channel) => ({
			id: channel.id,
			label: channel.name || 'Unnamed channel',
			supportingLabel: channel.messageRestriction === 'emoji_only' ? 'Emoji only' : 'Text and emoji'
		}))
	);
	const defaultRoomKinds: Array<{ key: 'general' | 'playerDm'; title: string }> = [
		{ key: 'general', title: 'General room' },
		{ key: 'playerDm', title: 'Private player messages' }
	];
	const senderDisplayOptions: SelectOption<RulesetChatChannel['senderDisplay']>[] = [
		{ value: 'profile_name', label: 'Profile name' },
		{ value: 'game_alias', label: 'Game alias' },
		{ value: 'seat_number', label: 'Seat number' },
		{ value: 'role_label', label: 'Role name' },
		{ value: 'team_label', label: 'Team name' }
	];
	const roomPolicy = (): RulesetRoomPermission => ({
		visible: true,
		readable: true,
		sendable: true,
		gameMasterMaySend: true,
		senderDisplay: 'game_alias'
	});
	const partialPolicy = (): RulesetPartialRoomPermission => ({});
	function addDefaultRoom(kind: 'general' | 'playerDm') {
		definition.chat.defaultPolicy[kind] = roomPolicy();
	}
	function addDefaultTeam(teamID: string) {
		definition.chat.defaultPolicy.teams[teamID] = roomPolicy();
	}
	function addCustomChannel() {
		definition.chat.channels ??= [];
		const channel: RulesetChatChannel = {
			id: nextID(
				'channel',
				definition.chat.channels.map((item) => item.id)
			),
			name: 'New channel',
			readerRoleIds: [],
			readerTeamIds: [],
			senderRoleIds: [],
			senderTeamIds: [],
			messageRestriction: 'normal_text',
			visible: true,
			sendable: true,
			gameMasterMaySend: true,
			senderDisplay: 'game_alias',
			phaseOverrides: {}
		};
		definition.chat.channels.push(channel);
		selectedItems.channels = channel.id;
	}
	function phaseChannelState(
		channel: RulesetChatChannel,
		phaseID: string,
		key: 'visible' | 'sendable'
	) {
		const value = channel.phaseOverrides[phaseID]?.[key];
		return value === undefined ? 'inherit' : value ? 'yes' : 'no';
	}
	function setPhaseChannelState(
		channel: RulesetChatChannel,
		phaseID: string,
		key: 'visible' | 'sendable',
		value: string
	) {
		channel.phaseOverrides[phaseID] ??= {};
		if (value === 'inherit') delete channel.phaseOverrides[phaseID][key];
		else channel.phaseOverrides[phaseID][key] = value === 'yes';
		const override = channel.phaseOverrides[phaseID];
		if (override.visible === undefined && override.sendable === undefined)
			delete channel.phaseOverrides[phaseID];
	}
	function addPhaseOverride(phaseID: string) {
		definition.chat.phaseOverrides[phaseID] = { teams: {} };
	}
	function addOverrideRoom(phaseID: string, kind: 'general' | 'playerDm') {
		definition.chat.phaseOverrides[phaseID][kind] = partialPolicy();
	}
	function addOverrideTeam(phaseID: string, teamID: string) {
		const override = definition.chat.phaseOverrides[phaseID];
		override.teams ??= {};
		override.teams[teamID] = partialPolicy();
	}
</script>

<ContentHeader density="dense" description="These settings apply unless a phase changes them."
	>{#snippet title()}<h2>Normal chat settings</h2>{/snippet}</ContentHeader
>
<div class="cards">
	{#each defaultRoomKinds as room (room.key)}
		{@const policy = definition.chat.defaultPolicy[room.key]}
		<article class="item-card">
			<ContentHeader density="dense"
				>{#snippet title()}<h3>
						{room.title}
					</h3>{/snippet}{#snippet actions()}{#if policy}<button
							class="remove"
							onclick={() => delete definition.chat.defaultPolicy[room.key]}>Disable room</button
						>{:else}<button class="add-small" onclick={() => addDefaultRoom(room.key)}
							>Enable room</button
						>{/if}{/snippet}</ContentHeader
			>{#if policy}<RoomPermissionEditor {policy} />{/if}
		</article>
	{/each}
	{#each definition.teams as team (team.id)}
		<article class="item-card">
			<ContentHeader density="dense"
				>{#snippet title()}<h3>
						{team.name} team room
					</h3>{/snippet}{#snippet actions()}{#if definition.chat.defaultPolicy.teams[team.id]}<button
							class="remove"
							onclick={() => delete definition.chat.defaultPolicy.teams[team.id]}
							>Disable room</button
						>{:else}<button class="add-small" onclick={() => addDefaultTeam(team.id)}
							>Enable room</button
						>{/if}{/snippet}</ContentHeader
			>{#if definition.chat.defaultPolicy.teams[team.id]}<RoomPermissionEditor
					policy={definition.chat.defaultPolicy.teams[team.id]}
				/>{/if}
		</article>
	{/each}
</div>

<CollectionEditor
	title="Custom channels"
	description="Create optional role- or team-specific conversations, including emoji-only channels."
	entries={channelEntries}
	selectedId={selectedItems.channels ?? ''}
	onselect={(id) => (selectedItems.channels = id)}
	onadd={addCustomChannel}
	onduplicate={(id) => {
		const item = duplicateByID(definition.chat.channels, id, 'channel');
		if (item) selectedItems.channels = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.chat.channels, id, direction)}
	onremove={(id) =>
		definition.chat.channels.splice(
			definition.chat.channels.findIndex((item) => item.id === id),
			1
		)}
>
	{#snippet editor(id)}{@const channelIndex = definition.chat.channels.findIndex(
			(item) => item.id === id
		)}{@const channel = definition.chat.channels[channelIndex]}{#if channel}
			<h3>{channel.name || 'Unnamed channel'}</h3>
			<div class="form-grid">
				<Field
					label="Channel name"
					name={`channel-name-${channelIndex}`}
					bind:value={channel.name}
					required
				/><SelectField
					label="Allowed messages"
					name={`channel-message-restriction-${channelIndex}`}
					bind:value={channel.messageRestriction}
					options={[
						{ value: 'normal_text', label: 'Normal text and emoji' },
						{ value: 'emoji_only', label: 'Emoji only' }
					]}
				/><SelectField
					label="Show senders as"
					name={`channel-sender-display-${channelIndex}`}
					bind:value={channel.senderDisplay}
					options={senderDisplayOptions}
				/>
			</div>
			<div class="form-grid thirds">
				<CheckboxField
					label="Players can normally see this channel"
					name={`channel-visible-${channelIndex}`}
					bind:checked={channel.visible}
				/><CheckboxField
					label="Allowed senders can normally post"
					name={`channel-sendable-${channelIndex}`}
					bind:checked={channel.sendable}
				/><CheckboxField
					label="Game masters can post"
					name={`channel-gm-sendable-${channelIndex}`}
					bind:checked={channel.gameMasterMaySend}
				/>
			</div>
			<div class="audience-grid">
				<div class="choice-block">
					<p class="hint compact">No reader selections means every player.</p>
					<CheckboxGroup
						label="Readers by team"
						name={`channel-reader-teams-${channelIndex}`}
						bind:values={channel.readerTeamIds}
						options={definition.teams.map((team) => ({ value: team.id, label: team.name }))}
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Readers by role"
						name={`channel-reader-roles-${channelIndex}`}
						bind:values={channel.readerRoleIds}
						options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
					/>
				</div>
				<div class="choice-block">
					<p class="hint compact">No sender selections means every reader.</p>
					<CheckboxGroup
						label="Senders by team"
						name={`channel-sender-teams-${channelIndex}`}
						bind:values={channel.senderTeamIds}
						options={definition.teams.map((team) => ({ value: team.id, label: team.name }))}
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Senders by role"
						name={`channel-sender-roles-${channelIndex}`}
						bind:values={channel.senderRoleIds}
						options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
					/>
				</div>
			</div>
			{#if definition.phases.length > 0}<div class="override-room">
					<strong>Phase permissions</strong>
					<div class="phase-permissions">
						{#each definition.phases as phase (phase.id)}<div>
								<b>{phase.name}</b><SelectField
									label="Visibility"
									name={`channel-visibility-${channelIndex}-${phase.id}`}
									accessibleLabel={`${channel.name} visibility during ${phase.name}`}
									value={phaseChannelState(channel, phase.id, 'visible')}
									options={[
										{ value: 'inherit', label: 'Use normal setting' },
										{ value: 'yes', label: 'Visible' },
										{ value: 'no', label: 'Hidden' }
									]}
									onchange={(value) => setPhaseChannelState(channel, phase.id, 'visible', value)}
								/><SelectField
									label="Posting"
									name={`channel-posting-${channelIndex}-${phase.id}`}
									accessibleLabel={`${channel.name} posting during ${phase.name}`}
									value={phaseChannelState(channel, phase.id, 'sendable')}
									options={[
										{ value: 'inherit', label: 'Use normal setting' },
										{ value: 'yes', label: 'Allowed' },
										{ value: 'no', label: 'Read-only' }
									]}
									onchange={(value) => setPhaseChannelState(channel, phase.id, 'sendable', value)}
								/>
							</div>{/each}
					</div>
				</div>{/if}
		{/if}{/snippet}
</CollectionEditor>

<div class="subsection">
	<ContentHeader
		density="dense"
		description="Temporarily override only the settings that need to change."
		>{#snippet title()}<h2>Phase changes</h2>{/snippet}</ContentHeader
	>
</div>
<div class="cards">
	{#each definition.phases as phase (phase.id)}
		<article class="item-card">
			<ContentHeader density="dense"
				>{#snippet title()}<h3>
						{phase.name}
					</h3>{/snippet}{#snippet actions()}{#if definition.chat.phaseOverrides[phase.id]}<button
							class="remove"
							onclick={() => delete definition.chat.phaseOverrides[phase.id]}>Remove changes</button
						>{:else}<button class="add-small" onclick={() => addPhaseOverride(phase.id)}
							>Add phase changes</button
						>{/if}{/snippet}</ContentHeader
			>
			{#if definition.chat.phaseOverrides[phase.id]}{@const override =
					definition.chat.phaseOverrides[
						phase.id
					]}{#each defaultRoomKinds as room (room.key)}{@const policy = override[room.key]}
					<div class="override-room">
						<ContentHeader density="dense"
							>{#snippet title()}<strong>{room.title}</strong
								>{/snippet}{#snippet actions()}{#if policy}<button
										class="remove"
										onclick={() => delete override[room.key]}>Clear</button
									>{:else}<button
										class="add-small"
										onclick={() => addOverrideRoom(phase.id, room.key)}>Change</button
									>{/if}{/snippet}</ContentHeader
						>{#if policy}<RoomPermissionEditor {policy} partial />{/if}
					</div>{/each}{#each definition.teams as team (team.id)}<div class="override-room">
						<ContentHeader density="dense"
							>{#snippet title()}<strong>{team.name} room</strong
								>{/snippet}{#snippet actions()}{#if override.teams?.[team.id]}<button
										class="remove"
										onclick={() => delete override.teams?.[team.id]}>Clear</button
									>{:else}<button
										class="add-small"
										onclick={() => addOverrideTeam(phase.id, team.id)}>Change</button
									>{/if}{/snippet}</ContentHeader
						>{#if override.teams?.[team.id]}<RoomPermissionEditor
								policy={override.teams[team.id]}
								partial
							/>{/if}
					</div>{/each}{/if}
		</article>
	{:else}<p class="empty">Create phases before adding phase-specific chat changes.</p>{/each}
</div>
