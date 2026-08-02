<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField, { type SelectOption } from '$lib/components/SelectField.svelte';
	import RoomPermissionEditor from './RoomPermissionEditor.svelte';
	import SelectorEditor from './SelectorEditor.svelte';
	import type {
		RulesetChatChannel,
		RulesetDefinition,
		RulesetPartialRoomPermission,
		RulesetRoomPermission,
		RulesetSelector
	} from '$lib/api/types';

	type Section =
		| 'teams'
		| 'roles'
		| 'phases'
		| 'composition'
		| 'knowledge'
		| 'chat'
		| 'achievements'
		| 'audio';
	type AssetOption = { assetKey: string; kind: 'image' | 'audio' };

	let {
		definition = $bindable(),
		section,
		assets
	}: { definition: RulesetDefinition; section: Section; assets: AssetOption[] } = $props();

	const blankSelector = (): RulesetSelector => ({
		roleIds: [],
		teamIds: [],
		categoryIds: [],
		tags: []
	});
	const roomPolicy = (): RulesetRoomPermission => ({
		visible: true,
		readable: true,
		sendable: true,
		gameMasterMaySend: true,
		senderDisplay: 'game_alias'
	});
	const partialPolicy = (): RulesetPartialRoomPermission => ({});

	function nextID(prefix: string, used: string[]) {
		let number = used.length + 1;
		let candidate = `${prefix}_${number}`;
		while (used.includes(candidate)) {
			number += 1;
			candidate = `${prefix}_${number}`;
		}
		return candidate;
	}

	function removeAt<T>(items: T[], index: number) {
		items.splice(index, 1);
	}

	function addTeam() {
		definition.teams.push({
			id: nextID(
				'team',
				definition.teams.map((item) => item.id)
			),
			name: 'New team',
			description: ''
		});
	}

	function addCategory() {
		definition.categories.push({
			id: nextID(
				'category',
				definition.categories.map((item) => item.id)
			),
			name: 'New category',
			description: ''
		});
	}

	function addAbility() {
		definition.abilities.push({
			id: nextID(
				'ability',
				definition.abilities.map((item) => item.id)
			),
			name: 'New ability',
			description: '',
			activationPhaseIds: [],
			canCombineWithOtherAbilities: false
		});
	}

	function addRole() {
		definition.roles.push({
			id: nextID(
				'role',
				definition.roles.map((item) => item.id)
			),
			name: 'New role',
			description: '',
			teamId: definition.teams[0]?.id ?? '',
			categoryIds: [],
			tags: [],
			abilityIds: [],
			winCondition: '',
			maxCopies: 1
		});
	}

	function addPhase() {
		const order = Math.max(0, ...definition.phases.map((item) => item.order)) + 1;
		definition.phases.push({
			id: nextID(
				'phase',
				definition.phases.map((item) => item.id)
			),
			name: 'New phase',
			description: '',
			order,
			startsRound: false,
			suggestedDurationSeconds: 0
		});
	}

	function addBand() {
		definition.compositionBands.push({
			id: nextID(
				'band',
				definition.compositionBands.map((item) => item.id)
			),
			minPlayers: definition.metadata.minPlayers,
			maxPlayers: definition.metadata.maxPlayers,
			slots: []
		});
	}

	function addSlot(bandIndex: number) {
		const band = definition.compositionBands[bandIndex];
		const used = definition.compositionBands.flatMap((item) => item.slots.map((slot) => slot.id));
		band.slots.push({
			id: nextID('slot', used),
			label: 'Role slot',
			count: 1,
			selector: blankSelector()
		});
	}

	function addModifier() {
		definition.compositionModifiers.push({
			id: nextID(
				'modifier',
				definition.compositionModifiers.map((item) => item.id)
			),
			whenRolePresent: definition.roles[0]?.id ?? '',
			slotAdjustments: [],
			requiresRoleIds: [],
			excludesRoleIds: []
		});
	}

	function addAdjustment(modifierIndex: number) {
		const firstSlot = definition.compositionBands[0]?.slots[0]?.id ?? '';
		definition.compositionModifiers[modifierIndex].slotAdjustments.push({
			slotId: firstSlot,
			delta: 1
		});
	}

	function addKnowledge() {
		definition.knowledgeRules.push({
			viewer: blankSelector(),
			target: blankSelector(),
			reveal: ['role']
		});
	}

	function addAchievement() {
		definition.achievements.push({
			id: nextID(
				'achievement',
				definition.achievements.map((item) => item.id)
			),
			name: 'New achievement',
			description: '',
			points: 0,
			hiddenUntilGameCompleted: false
		});
	}

	function addAudioCue() {
		definition.audioCues.push({
			id: nextID(
				'cue',
				definition.audioCues.map((item) => item.id)
			),
			name: 'New audio cue',
			assetKey: assets.find((asset) => asset.kind === 'audio')?.assetKey ?? '',
			defaultAudience: 'all'
		});
	}

	function setTags(index: number, value: string) {
		definition.roles[index].tags = value
			.split(',')
			.map((tag) => tag.trim())
			.filter(Boolean);
	}

	function addDefaultRoom(kind: 'general' | 'playerDm') {
		definition.chat.defaultPolicy[kind] = roomPolicy();
	}

	function addCustomChannel() {
		definition.chat.channels ??= [];
		definition.chat.channels.push({
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
		});
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
		if (override.visible === undefined && override.sendable === undefined) {
			delete channel.phaseOverrides[phaseID];
		}
	}

	function addDefaultTeam(teamID: string) {
		definition.chat.defaultPolicy.teams[teamID] = roomPolicy();
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

	function imageAssets() {
		return assets.filter((asset) => asset.kind === 'image');
	}

	function audioAssets() {
		return assets.filter((asset) => asset.kind === 'audio');
	}

	const imageOptions = () => [
		{ value: '', label: 'No image' },
		...imageAssets().map((asset) => ({ value: asset.assetKey, label: asset.assetKey }))
	];
	const senderDisplayOptions: SelectOption<RulesetChatChannel['senderDisplay']>[] = [
		{ value: 'profile_name', label: 'Profile name' },
		{ value: 'game_alias', label: 'Game alias' },
		{ value: 'seat_number', label: 'Seat number' },
		{ value: 'role_label', label: 'Role name' },
		{ value: 'team_label', label: 'Team name' }
	];
</script>

{#if section === 'teams'}
	<ContentHeader density="dense" description="The main sides or factions in the game.">
		{#snippet title()}<h2>Teams</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addTeam}>Add team</Button>{/snippet}
	</ContentHeader>
	<div class="cards">
		{#each definition.teams as team, index (team.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{team.name || 'Unnamed team'}</h3>{/snippet}
					{#snippet actions()}<button
							class="remove"
							onclick={() => removeAt(definition.teams, index)}>Remove</button
						>{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field label="Name" name={`team-name-${index}`} bind:value={team.name} required />
					<Field
						label="Stable ID"
						name={`team-id-${index}`}
						bind:value={team.id}
						help="Used when other rules refer to this team."
						required
					/>
				</div>
				<Field
					label="Description"
					name={`team-description-${index}`}
					bind:value={team.description}
					multiline
				/>
				<SelectField
					label="Team image (optional)"
					name={`team-image-${index}`}
					bind:value={team.imageAssetKey}
					options={imageOptions()}
				/>
			</article>
		{:else}
			<p class="empty">Add at least one team before creating roles.</p>
		{/each}
	</div>

	<div class="subsection">
		<ContentHeader density="dense" description="Optional labels such as Investigative or Support.">
			{#snippet title()}<h2>Categories</h2>{/snippet}
			{#snippet actions()}<Button variant="secondary" onclick={addCategory}>Add category</Button
				>{/snippet}
		</ContentHeader>
	</div>
	<div class="cards compact">
		{#each definition.categories as category, index (category.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{category.name || 'Unnamed category'}</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.categories, index)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field label="Name" name={`category-name-${index}`} bind:value={category.name} required />
					<Field
						label="Stable ID"
						name={`category-id-${index}`}
						bind:value={category.id}
						required
					/>
				</div>
				<Field
					label="Description"
					name={`category-description-${index}`}
					bind:value={category.description}
					multiline
				/>
			</article>
		{/each}
	</div>
{:else if section === 'roles'}
	<ContentHeader density="dense" description="Reusable powers that can be assigned to roles.">
		{#snippet title()}<h2>Abilities</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addAbility}>Add ability</Button
			>{/snippet}
	</ContentHeader>
	<div class="cards compact">
		{#each definition.abilities as ability, index (ability.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{ability.name || 'Unnamed ability'}</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.abilities, index)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field label="Name" name={`ability-name-${index}`} bind:value={ability.name} required />
					<Field label="Stable ID" name={`ability-id-${index}`} bind:value={ability.id} required />
				</div>
				<Field
					label="Description"
					name={`ability-description-${index}`}
					bind:value={ability.description}
					multiline
				/>
				<SelectField
					label="Ability image (optional)"
					name={`ability-image-${index}`}
					bind:value={ability.imageAssetKey}
					options={imageOptions()}
				/>
				<CheckboxField
					label="May combine with other combinable abilities"
					name={`ability-combinable-${index}`}
					checked={ability.canCombineWithOtherAbilities ?? false}
					onchange={(checked) => (ability.canCombineWithOtherAbilities = checked)}
				/>
				<div class="choice-block">
					<CheckboxGroup
						label="Playable during phases"
						name={`ability-phases-${index}`}
						bind:values={ability.activationPhaseIds}
						options={definition.phases.map((phase) => ({ value: phase.id, label: phase.name }))}
					/>
					{#if definition.phases.length === 0}
						<p class="hint compact">Add phases before making this ability playable.</p>
					{/if}
				</div>
			</article>
		{/each}
	</div>

	<div class="subsection">
		<ContentHeader
			density="dense"
			description="What each player may be assigned and how that role wins."
		>
			{#snippet title()}<h2>Roles</h2>{/snippet}
			{#snippet actions()}<Button variant="secondary" onclick={addRole}>Add role</Button>{/snippet}
		</ContentHeader>
	</div>
	<div class="cards">
		{#each definition.roles as role, index (role.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{role.name || 'Unnamed role'}</h3>{/snippet}
					{#snippet actions()}<button
							class="remove"
							onclick={() => removeAt(definition.roles, index)}>Remove</button
						>{/snippet}
				</ContentHeader>
				<div class="form-grid thirds">
					<Field label="Name" name={`role-name-${index}`} bind:value={role.name} required />
					<Field label="Stable ID" name={`role-id-${index}`} bind:value={role.id} required />
					<SelectField
						label="Team"
						name={`role-team-${index}`}
						bind:value={role.teamId}
						options={[
							{ value: '', label: 'Choose a team' },
							...definition.teams.map((team) => ({ value: team.id, label: team.name }))
						]}
						required
					/>
				</div>
				<Field
					label="Description"
					name={`role-description-${index}`}
					bind:value={role.description}
					multiline
				/>
				<Field
					label="Win condition"
					name={`role-win-${index}`}
					bind:value={role.winCondition}
					multiline
				/>
				<div class="form-grid">
					<label>
						<span>Maximum copies</span>
						<input type="number" min="1" max="30" bind:value={role.maxCopies} />
					</label>
					<SelectField
						label="Role image (optional)"
						name={`role-image-${index}`}
						bind:value={role.imageAssetKey}
						options={imageOptions()}
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Categories"
						name={`role-categories-${index}`}
						bind:values={role.categoryIds}
						options={definition.categories.map((category) => ({
							value: category.id,
							label: category.name
						}))}
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Abilities"
						name={`role-abilities-${index}`}
						bind:values={role.abilityIds}
						options={definition.abilities.map((ability) => ({
							value: ability.id,
							label: ability.name
						}))}
					/>
				</div>
				<label>
					<span>Tags (comma-separated)</span>
					<input
						value={role.tags.join(', ')}
						onchange={(event) => setTags(index, event.currentTarget.value)}
						placeholder="investigative, unique"
					/>
				</label>
			</article>
		{:else}
			<p class="empty">Add roles after defining at least one team.</p>
		{/each}
	</div>
{:else if section === 'phases'}
	<ContentHeader density="dense" description="The ordered steps a game master advances through.">
		{#snippet title()}<h2>Phases</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addPhase}>Add phase</Button>{/snippet}
	</ContentHeader>
	<div class="cards">
		{#each definition.phases as phase, index (phase.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{phase.order}. {phase.name || 'Unnamed phase'}</h3>{/snippet}
					{#snippet actions()}<button
							class="remove"
							onclick={() => removeAt(definition.phases, index)}>Remove</button
						>{/snippet}
				</ContentHeader>
				<div class="form-grid thirds">
					<Field label="Name" name={`phase-name-${index}`} bind:value={phase.name} required />
					<Field label="Stable ID" name={`phase-id-${index}`} bind:value={phase.id} required />
					<label><span>Order</span><input type="number" min="1" bind:value={phase.order} /></label>
				</div>
				<Field
					label="Instructions"
					name={`phase-description-${index}`}
					bind:value={phase.description}
					multiline
				/>
				<div class="form-grid thirds">
					<label
						><span>Suggested seconds</span><input
							type="number"
							min="0"
							bind:value={phase.suggestedDurationSeconds}
						/></label
					>
					<CheckboxField
						label="Starts a new round"
						name={`phase-starts-round-${index}`}
						bind:checked={phase.startsRound}
					/>
					<SelectField
						label="Sound when phase starts"
						name={`phase-audio-${index}`}
						bind:value={phase.audioCueId}
						options={[
							{ value: '', label: 'No sound' },
							...definition.audioCues.map((cue) => ({ value: cue.id, label: cue.name }))
						]}
					/>
				</div>
			</article>
		{:else}
			<p class="empty">Add the phases used to run this game.</p>
		{/each}
	</div>
{:else if section === 'composition'}
	<ContentHeader
		density="dense"
		description="Define how many slots are filled for every supported party size."
	>
		{#snippet title()}<h2>Player-count bands</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addBand}>Add band</Button>{/snippet}
	</ContentHeader>
	<div class="cards">
		{#each definition.compositionBands as band, bandIndex (band.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{band.minPlayers}–{band.maxPlayers} players</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.compositionBands, bandIndex)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid thirds">
					<Field label="Stable ID" name={`band-id-${bandIndex}`} bind:value={band.id} required />
					<label
						><span>Minimum players</span><input
							type="number"
							min="1"
							max="30"
							bind:value={band.minPlayers}
						/></label
					>
					<label
						><span>Maximum players</span><input
							type="number"
							min="1"
							max="30"
							bind:value={band.maxPlayers}
						/></label
					>
				</div>
				<div class="nested-heading">
					<ContentHeader density="dense">
						{#snippet title()}<strong>Role slots</strong>{/snippet}
						{#snippet actions()}<button class="add-small" onclick={() => addSlot(bandIndex)}
								>Add slot</button
							>{/snippet}
					</ContentHeader>
				</div>
				{#each band.slots as slot, slotIndex (slot.id)}
					<div class="nested">
						<ContentHeader density="dense">
							{#snippet title()}<strong>{slot.label || 'Unnamed slot'}</strong>{/snippet}
							{#snippet actions()}<button
									class="remove"
									onclick={() => removeAt(band.slots, slotIndex)}>Remove</button
								>{/snippet}
						</ContentHeader>
						<div class="form-grid thirds">
							<Field
								label="Label"
								name={`slot-label-${bandIndex}-${slotIndex}`}
								bind:value={slot.label}
							/>
							<Field
								label="Stable ID"
								name={`slot-id-${bandIndex}-${slotIndex}`}
								bind:value={slot.id}
							/>
							<label
								><span>Number of players</span><input
									type="number"
									min="0"
									max="30"
									bind:value={slot.count}
								/></label
							>
						</div>
						<SelectorEditor
							selector={slot.selector}
							roles={definition.roles}
							teams={definition.teams}
							categories={definition.categories}
							label="Roles allowed in this slot"
						/>
					</div>
				{/each}
			</article>
		{/each}
	</div>

	<div class="subsection">
		<ContentHeader density="dense" description="Adjust slots when a particular role appears.">
			{#snippet title()}<h2>Conditional changes</h2>{/snippet}
			{#snippet actions()}<Button variant="secondary" onclick={addModifier}>Add condition</Button
				>{/snippet}
		</ContentHeader>
	</div>
	<div class="cards">
		{#each definition.compositionModifiers as modifier, modifierIndex (modifier.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{modifier.id}</h3>{/snippet}
					{#snippet actions()}
						<button
							class="remove"
							onclick={() => removeAt(definition.compositionModifiers, modifierIndex)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field label="Stable ID" name={`modifier-id-${modifierIndex}`} bind:value={modifier.id} />
					<SelectField
						label="When this role is present"
						name={`modifier-role-${modifierIndex}`}
						bind:value={modifier.whenRolePresent}
						options={[
							{ value: '', label: 'Choose a role' },
							...definition.roles.map((role) => ({ value: role.id, label: role.name }))
						]}
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Also require these roles"
						name={`modifier-required-roles-${modifierIndex}`}
						bind:values={modifier.requiresRoleIds}
						options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Do not apply with these roles"
						name={`modifier-excluded-roles-${modifierIndex}`}
						bind:values={modifier.excludesRoleIds}
						options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
					/>
				</div>
				<div class="nested-heading">
					<ContentHeader density="dense">
						{#snippet title()}<strong>Slot changes</strong>{/snippet}
						{#snippet actions()}<button
								class="add-small"
								onclick={() => addAdjustment(modifierIndex)}>Add change</button
							>{/snippet}
					</ContentHeader>
				</div>
				{#each modifier.slotAdjustments as adjustment, adjustmentIndex (adjustment)}
					<div class="inline-row">
						<SelectField
							label="Slot"
							name={`modifier-slot-${modifierIndex}-${adjustmentIndex}`}
							bind:value={adjustment.slotId}
							options={[
								{ value: '', label: 'Choose a slot' },
								...definition.compositionBands.map((band) => ({
									label: `${band.minPlayers}–${band.maxPlayers} players`,
									options: band.slots.map((slot) => ({ value: slot.id, label: slot.label }))
								}))
							]}
						/>
						<label
							><span>Change count by</span><input
								type="number"
								bind:value={adjustment.delta}
							/></label
						>
						<button
							class="remove"
							onclick={() => removeAt(modifier.slotAdjustments, adjustmentIndex)}>Remove</button
						>
					</div>
				{/each}
			</article>
		{/each}
	</div>
{:else if section === 'knowledge'}
	<ContentHeader density="dense" description="Choose what one group learns about another group.">
		{#snippet title()}<h2>Starting knowledge</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addKnowledge}
				>Add knowledge rule</Button
			>{/snippet}
	</ContentHeader>
	<div class="cards">
		{#each definition.knowledgeRules as rule, index (rule)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>Knowledge rule {index + 1}</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.knowledgeRules, index)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="selector-grid">
					<SelectorEditor
						selector={rule.viewer}
						roles={definition.roles}
						teams={definition.teams}
						categories={definition.categories}
						label="Who receives the knowledge?"
					/>
					<SelectorEditor
						selector={rule.target}
						roles={definition.roles}
						teams={definition.teams}
						categories={definition.categories}
						label="Who do they learn about?"
					/>
				</div>
				<div class="choice-block">
					<CheckboxGroup
						label="Reveal"
						name={`knowledge-reveal-${index}`}
						bind:values={rule.reveal}
						options={[
							{ value: 'identity', label: 'Player identity' },
							{ value: 'role', label: 'Role' },
							{ value: 'team', label: 'Team' },
							{ value: 'elimination_state', label: 'Elimination state' }
						]}
					/>
				</div>
			</article>
		{:else}
			<p class="empty">
				No special starting knowledge. Add a rule if teammates or special roles should recognize
				anyone.
			</p>
		{/each}
	</div>
{:else if section === 'chat'}
	<ContentHeader density="dense" description="These settings apply unless a phase changes them.">
		{#snippet title()}<h2>Normal chat settings</h2>{/snippet}
	</ContentHeader>
	<div class="cards">
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>General room</h3>{/snippet}
				{#snippet actions()}
					{#if definition.chat.defaultPolicy.general}
						<button class="remove" onclick={() => delete definition.chat.defaultPolicy.general}
							>Disable room</button
						>
					{:else}
						<button class="add-small" onclick={() => addDefaultRoom('general')}>Enable room</button>
					{/if}
				{/snippet}
			</ContentHeader>
			{#if definition.chat.defaultPolicy.general}
				<RoomPermissionEditor policy={definition.chat.defaultPolicy.general} />
			{/if}
		</article>
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>Private player messages</h3>{/snippet}
				{#snippet actions()}
					{#if definition.chat.defaultPolicy.playerDm}
						<button class="remove" onclick={() => delete definition.chat.defaultPolicy.playerDm}
							>Disable room</button
						>
					{:else}
						<button class="add-small" onclick={() => addDefaultRoom('playerDm')}>Enable room</button
						>
					{/if}
				{/snippet}
			</ContentHeader>
			{#if definition.chat.defaultPolicy.playerDm}
				<RoomPermissionEditor policy={definition.chat.defaultPolicy.playerDm} />
			{/if}
		</article>
		{#each definition.teams as team (team.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{team.name} team room</h3>{/snippet}
					{#snippet actions()}
						{#if definition.chat.defaultPolicy.teams[team.id]}
							<button
								class="remove"
								onclick={() => delete definition.chat.defaultPolicy.teams[team.id]}
								>Disable room</button
							>
						{:else}
							<button class="add-small" onclick={() => addDefaultTeam(team.id)}>Enable room</button>
						{/if}
					{/snippet}
				</ContentHeader>
				{#if definition.chat.defaultPolicy.teams[team.id]}
					<RoomPermissionEditor policy={definition.chat.defaultPolicy.teams[team.id]} />
				{/if}
			</article>
		{/each}
	</div>

	<div class="subsection">
		<ContentHeader
			density="dense"
			description="Create role- or team-specific conversations, including emoji-only channels."
		>
			{#snippet title()}<h2>Custom channels</h2>{/snippet}
			{#snippet actions()}<Button variant="secondary" onclick={addCustomChannel}>Add channel</Button
				>{/snippet}
		</ContentHeader>
	</div>
	<div class="cards">
		{#each definition.chat.channels ?? [] as channel, channelIndex (channel.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{channel.name || 'Unnamed channel'}</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.chat.channels, channelIndex)}
							>Remove channel</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field
						label="Channel name"
						name={`channel-name-${channelIndex}`}
						bind:value={channel.name}
						required
					/>
					<Field
						label="Stable ID"
						name={`channel-id-${channelIndex}`}
						bind:value={channel.id}
						help="Used to preserve this channel in saved games."
						required
					/>
					<SelectField
						label="Allowed messages"
						name={`channel-message-restriction-${channelIndex}`}
						bind:value={channel.messageRestriction}
						options={[
							{ value: 'normal_text', label: 'Normal text and emoji' },
							{ value: 'emoji_only', label: 'Emoji only' }
						]}
					/>
					<SelectField
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
					/>
					<CheckboxField
						label="Allowed senders can normally post"
						name={`channel-sendable-${channelIndex}`}
						bind:checked={channel.sendable}
					/>
					<CheckboxField
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
				{#if definition.phases.length > 0}
					<div class="override-room">
						<strong>Phase permissions</strong>
						<div class="phase-permissions">
							{#each definition.phases as phase (phase.id)}
								<div>
									<b>{phase.name}</b>
									<SelectField
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
									/>
									<SelectField
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
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</article>
		{:else}
			<p class="empty">No custom chat channels in this ruleset.</p>
		{/each}
	</div>

	<div class="subsection">
		<ContentHeader
			density="dense"
			description="Temporarily override only the settings that need to change."
		>
			{#snippet title()}<h2>Phase changes</h2>{/snippet}
		</ContentHeader>
	</div>
	<div class="cards">
		{#each definition.phases as phase (phase.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{phase.name}</h3>{/snippet}
					{#snippet actions()}
						{#if definition.chat.phaseOverrides[phase.id]}
							<button class="remove" onclick={() => delete definition.chat.phaseOverrides[phase.id]}
								>Remove changes</button
							>
						{:else}
							<button class="add-small" onclick={() => addPhaseOverride(phase.id)}
								>Add phase changes</button
							>
						{/if}
					{/snippet}
				</ContentHeader>
				{#if definition.chat.phaseOverrides[phase.id]}
					{@const override = definition.chat.phaseOverrides[phase.id]}
					<div class="override-room">
						<ContentHeader density="dense">
							{#snippet title()}<strong>General room</strong>{/snippet}
							{#snippet actions()}
								{#if override.general}
									<button class="remove" onclick={() => delete override.general}>Clear</button>
								{:else}
									<button class="add-small" onclick={() => addOverrideRoom(phase.id, 'general')}
										>Change</button
									>
								{/if}
							{/snippet}
						</ContentHeader>
						{#if override.general}<RoomPermissionEditor policy={override.general} partial />{/if}
					</div>
					<div class="override-room">
						<ContentHeader density="dense">
							{#snippet title()}<strong>Private player messages</strong>{/snippet}
							{#snippet actions()}
								{#if override.playerDm}
									<button class="remove" onclick={() => delete override.playerDm}>Clear</button>
								{:else}
									<button class="add-small" onclick={() => addOverrideRoom(phase.id, 'playerDm')}
										>Change</button
									>
								{/if}
							{/snippet}
						</ContentHeader>
						{#if override.playerDm}<RoomPermissionEditor policy={override.playerDm} partial />{/if}
					</div>
					{#each definition.teams as team (team.id)}
						<div class="override-room">
							<ContentHeader density="dense">
								{#snippet title()}<strong>{team.name} room</strong>{/snippet}
								{#snippet actions()}
									{#if override.teams?.[team.id]}
										<button class="remove" onclick={() => delete override.teams?.[team.id]}
											>Clear</button
										>
									{:else}
										<button class="add-small" onclick={() => addOverrideTeam(phase.id, team.id)}
											>Change</button
										>
									{/if}
								{/snippet}
							</ContentHeader>
							{#if override.teams?.[team.id]}
								<RoomPermissionEditor policy={override.teams[team.id]} partial />
							{/if}
						</div>
					{/each}
				{/if}
			</article>
		{:else}
			<p class="empty">Create phases before adding phase-specific chat changes.</p>
		{/each}
	</div>
{:else if section === 'achievements'}
	<ContentHeader density="dense" description="Awards a game master can give after a game.">
		{#snippet title()}<h2>Achievements</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addAchievement}>Add achievement</Button
			>{/snippet}
	</ContentHeader>
	<div class="cards">
		{#each definition.achievements as achievement, index (achievement.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{achievement.name || 'Unnamed achievement'}</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.achievements, index)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field
						label="Name"
						name={`achievement-name-${index}`}
						bind:value={achievement.name}
						required
					/>
					<Field
						label="Stable ID"
						name={`achievement-id-${index}`}
						bind:value={achievement.id}
						required
					/>
				</div>
				<Field
					label="Description"
					name={`achievement-description-${index}`}
					bind:value={achievement.description}
					multiline
				/>
				<div class="form-grid">
					<label>
						<span>Achievement points</span>
						<input type="number" min="0" max="10000" bind:value={achievement.points} />
					</label>
					<CheckboxField
						label="Hide from players until the game ends"
						name={`achievement-hidden-${index}`}
						bind:checked={achievement.hiddenUntilGameCompleted}
					/>
				</div>
				<SelectField
					label="Badge image (optional)"
					name={`achievement-image-${index}`}
					bind:value={achievement.imageAssetKey}
					options={imageOptions()}
				/>
			</article>
		{:else}
			<p class="empty">No achievements in this ruleset.</p>
		{/each}
	</div>
{:else if section === 'audio'}
	<ContentHeader
		density="dense"
		description="Named sounds a game master or phase can play for selected listeners."
	>
		{#snippet title()}<h2>Audio cues</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addAudioCue}>Add audio cue</Button
			>{/snippet}
	</ContentHeader>
	<div class="cards">
		{#each definition.audioCues as cue, index (cue.id)}
			<article class="item-card">
				<ContentHeader density="dense">
					{#snippet title()}<h3>{cue.name || 'Unnamed cue'}</h3>{/snippet}
					{#snippet actions()}
						<button class="remove" onclick={() => removeAt(definition.audioCues, index)}
							>Remove</button
						>
					{/snippet}
				</ContentHeader>
				<div class="form-grid">
					<Field label="Name" name={`cue-name-${index}`} bind:value={cue.name} required />
					<Field label="Stable ID" name={`cue-id-${index}`} bind:value={cue.id} required />
					<SelectField
						label="Audio file"
						name={`cue-audio-file-${index}`}
						bind:value={cue.assetKey}
						options={[
							{ value: '', label: 'Choose uploaded audio' },
							...audioAssets().map((asset) => ({ value: asset.assetKey, label: asset.assetKey }))
						]}
					/>
					<SelectField
						label="Normal audience"
						name={`cue-audience-${index}`}
						bind:value={cue.defaultAudience}
						options={[
							{ value: 'all', label: 'All players' },
							{ value: 'game_masters', label: 'Game masters' },
							{ value: 'team', label: 'A selected team' },
							{ value: 'player', label: 'A selected player' }
						]}
					/>
				</div>
				{#if cue.defaultAudience === 'team' || cue.defaultAudience === 'player'}
					<p class="hint">
						This cue needs a team or player chosen when a game master plays it. It cannot be used as
						an automatic phase sound.
					</p>
				{/if}
			</article>
		{:else}
			<p class="empty">Upload an audio file, then add a cue that uses it.</p>
		{/each}
	</div>
{/if}

<style>
	.inline-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}

	.subsection {
		margin-top: 1.5rem;
		border-top: 2px solid #9a7e51;
		padding-top: 1.25rem;
	}

	.cards {
		display: grid;
		gap: 0.9rem;
	}

	.item-card {
		display: grid;
		gap: 0.8rem;
		border: 1px solid #a98c60;
		background: rgb(255 249 230 / 34%);
		padding: 0.9rem;
	}

	.item-card h3 {
		font-size: 1rem;
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem;
	}

	.form-grid.thirds {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	label {
		display: grid;
		gap: 0.3rem;
	}

	label > span,
	.choice-block > strong,
	.nested-heading :global(strong),
	.override-room strong {
		font-family: var(--font-display);
		font-size: 0.67rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	input:not([type='checkbox']) {
		width: 100%;
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.55rem;
	}

	.remove,
	.add-small {
		min-height: 36px;
		border: 1px solid #9a7e51;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.62rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		padding: 0.35rem 0.55rem;
		text-transform: uppercase;
	}

	.add-small {
		color: var(--ink);
	}

	.nested {
		display: grid;
		gap: 0.7rem;
		border-inline-start: 3px solid #b89b6d;
		padding: 0.7rem;
	}

	.inline-row {
		align-items: end;
	}

	.inline-row > label {
		flex: 1;
	}

	.selector-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem;
	}

	.audience-grid,
	.phase-permissions {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem;
	}

	.phase-permissions > div {
		display: grid;
		gap: 0.55rem;
		border: 1px solid #c5ad82;
		padding: 0.65rem;
	}

	.phase-permissions b {
		font-family: var(--font-display);
		font-size: 0.78rem;
	}

	.hint.compact {
		border: 0;
		background: transparent;
		padding: 0.2rem 0 0;
	}

	.override-room {
		display: grid;
		gap: 0.55rem;
		border-top: 1px solid #c5ad82;
		padding-top: 0.65rem;
	}

	.empty,
	.hint {
		border: 1px dashed #b89b6d;
		background: rgb(255 249 230 / 30%);
		margin: 0;
		padding: 0.75rem;
	}

	.hint {
		color: var(--ink-soft);
		font-size: 0.82rem;
	}

	@media (max-width: 720px) {
		.form-grid,
		.form-grid.thirds,
		.selector-grid {
			grid-template-columns: 1fr;
		}

		.audience-grid,
		.phase-permissions {
			grid-template-columns: 1fr;
		}

		.inline-row {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
