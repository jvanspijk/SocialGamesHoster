<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
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
				<label>
					<span>Team image (optional)</span>
					<select bind:value={team.imageAssetKey}>
						<option value="">No image</option>
						{#each imageAssets() as asset (asset.assetKey)}
							<option value={asset.assetKey}>{asset.assetKey}</option>
						{/each}
					</select>
				</label>
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
				<label>
					<span>Ability image (optional)</span>
					<select bind:value={ability.imageAssetKey}>
						<option value="">No image</option>
						{#each imageAssets() as asset (asset.assetKey)}
							<option value={asset.assetKey}>{asset.assetKey}</option>
						{/each}
					</select>
				</label>
				<label class="check">
					<input type="checkbox" bind:checked={ability.canCombineWithOtherAbilities} />
					May combine with other combinable abilities
				</label>
				<div class="choice-block">
					<strong>Playable during phases</strong>
					<div class="choices">
						{#each definition.phases as phase (phase.id)}
							<label>
								<input type="checkbox" value={phase.id} bind:group={ability.activationPhaseIds} />
								{phase.name}
							</label>
						{/each}
					</div>
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
					<label>
						<span>Team</span>
						<select bind:value={role.teamId} required>
							<option value="">Choose a team</option>
							{#each definition.teams as team (team.id)}
								<option value={team.id}>{team.name}</option>
							{/each}
						</select>
					</label>
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
					<label>
						<span>Role image (optional)</span>
						<select bind:value={role.imageAssetKey}>
							<option value="">No image</option>
							{#each imageAssets() as asset (asset.assetKey)}
								<option value={asset.assetKey}>{asset.assetKey}</option>
							{/each}
						</select>
					</label>
				</div>
				<div class="choice-block">
					<strong>Categories</strong>
					<div class="choices">
						{#each definition.categories as category (category.id)}
							<label
								><input type="checkbox" value={category.id} bind:group={role.categoryIds} />
								{category.name}</label
							>
						{/each}
					</div>
				</div>
				<div class="choice-block">
					<strong>Abilities</strong>
					<div class="choices">
						{#each definition.abilities as ability (ability.id)}
							<label
								><input type="checkbox" value={ability.id} bind:group={role.abilityIds} />
								{ability.name}</label
							>
						{/each}
					</div>
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
					<label class="check"
						><input type="checkbox" bind:checked={phase.startsRound} /> Starts a new round</label
					>
					<label>
						<span>Sound when phase starts</span>
						<select bind:value={phase.audioCueId}>
							<option value="">No sound</option>
							{#each definition.audioCues as cue (cue.id)}
								<option value={cue.id}>{cue.name}</option>
							{/each}
						</select>
					</label>
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
					<label>
						<span>When this role is present</span>
						<select bind:value={modifier.whenRolePresent}>
							<option value="">Choose a role</option>
							{#each definition.roles as role (role.id)}<option value={role.id}>{role.name}</option
								>{/each}
						</select>
					</label>
				</div>
				<div class="choice-block">
					<strong>Also require these roles</strong>
					<div class="choices">
						{#each definition.roles as role (role.id)}
							<label
								><input type="checkbox" value={role.id} bind:group={modifier.requiresRoleIds} />
								{role.name}</label
							>
						{/each}
					</div>
				</div>
				<div class="choice-block">
					<strong>Do not apply with these roles</strong>
					<div class="choices">
						{#each definition.roles as role (role.id)}
							<label
								><input type="checkbox" value={role.id} bind:group={modifier.excludesRoleIds} />
								{role.name}</label
							>
						{/each}
					</div>
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
						<label>
							<span>Slot</span>
							<select bind:value={adjustment.slotId}>
								<option value="">Choose a slot</option>
								{#each definition.compositionBands as band (band.id)}
									<optgroup label={`${band.minPlayers}–${band.maxPlayers} players`}>
										{#each band.slots as slot (slot.id)}<option value={slot.id}>{slot.label}</option
											>{/each}
									</optgroup>
								{/each}
							</select>
						</label>
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
					<strong>Reveal</strong>
					<div class="choices">
						<label
							><input type="checkbox" value="identity" bind:group={rule.reveal} /> Player identity</label
						>
						<label><input type="checkbox" value="role" bind:group={rule.reveal} /> Role</label>
						<label><input type="checkbox" value="team" bind:group={rule.reveal} /> Team</label>
						<label
							><input type="checkbox" value="elimination_state" bind:group={rule.reveal} />
							Elimination state</label
						>
					</div>
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
					<label>
						<span>Allowed messages</span>
						<select bind:value={channel.messageRestriction}>
							<option value="normal_text">Normal text and emoji</option>
							<option value="emoji_only">Emoji only</option>
						</select>
					</label>
					<label>
						<span>Show senders as</span>
						<select bind:value={channel.senderDisplay}>
							<option value="profile_name">Profile name</option>
							<option value="game_alias">Game alias</option>
							<option value="seat_number">Seat number</option>
							<option value="role_label">Role name</option>
							<option value="team_label">Team name</option>
						</select>
					</label>
				</div>
				<div class="form-grid thirds">
					<label class="check">
						<input type="checkbox" bind:checked={channel.visible} />
						Players can normally see this channel
					</label>
					<label class="check">
						<input type="checkbox" bind:checked={channel.sendable} />
						Allowed senders can normally post
					</label>
					<label class="check">
						<input type="checkbox" bind:checked={channel.gameMasterMaySend} />
						Game masters can post
					</label>
				</div>
				<div class="audience-grid">
					<div class="choice-block">
						<strong>Readers by team</strong>
						<p class="hint compact">No reader selections means every player.</p>
						<div class="choices">
							{#each definition.teams as team (team.id)}
								<label>
									<input type="checkbox" value={team.id} bind:group={channel.readerTeamIds} />
									{team.name}
								</label>
							{/each}
						</div>
					</div>
					<div class="choice-block">
						<strong>Readers by role</strong>
						<div class="choices">
							{#each definition.roles as role (role.id)}
								<label>
									<input type="checkbox" value={role.id} bind:group={channel.readerRoleIds} />
									{role.name}
								</label>
							{/each}
						</div>
					</div>
					<div class="choice-block">
						<strong>Senders by team</strong>
						<p class="hint compact">No sender selections means every reader.</p>
						<div class="choices">
							{#each definition.teams as team (team.id)}
								<label>
									<input type="checkbox" value={team.id} bind:group={channel.senderTeamIds} />
									{team.name}
								</label>
							{/each}
						</div>
					</div>
					<div class="choice-block">
						<strong>Senders by role</strong>
						<div class="choices">
							{#each definition.roles as role (role.id)}
								<label>
									<input type="checkbox" value={role.id} bind:group={channel.senderRoleIds} />
									{role.name}
								</label>
							{/each}
						</div>
					</div>
				</div>
				{#if definition.phases.length > 0}
					<div class="override-room">
						<strong>Phase permissions</strong>
						<div class="phase-permissions">
							{#each definition.phases as phase (phase.id)}
								<div>
									<b>{phase.name}</b>
									<label>
										<span>Visibility</span>
										<select
											aria-label={`${channel.name} visibility during ${phase.name}`}
											value={phaseChannelState(channel, phase.id, 'visible')}
											onchange={(event) =>
												setPhaseChannelState(
													channel,
													phase.id,
													'visible',
													event.currentTarget.value
												)}
										>
											<option value="inherit">Use normal setting</option>
											<option value="yes">Visible</option>
											<option value="no">Hidden</option>
										</select>
									</label>
									<label>
										<span>Posting</span>
										<select
											aria-label={`${channel.name} posting during ${phase.name}`}
											value={phaseChannelState(channel, phase.id, 'sendable')}
											onchange={(event) =>
												setPhaseChannelState(
													channel,
													phase.id,
													'sendable',
													event.currentTarget.value
												)}
										>
											<option value="inherit">Use normal setting</option>
											<option value="yes">Allowed</option>
											<option value="no">Read-only</option>
										</select>
									</label>
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
					<label class="check">
						<input type="checkbox" bind:checked={achievement.hiddenUntilGameCompleted} />
						Hide from players until the game ends
					</label>
				</div>
				<label>
					<span>Badge image (optional)</span>
					<select bind:value={achievement.imageAssetKey}>
						<option value="">No image</option>
						{#each imageAssets() as asset (asset.assetKey)}
							<option value={asset.assetKey}>{asset.assetKey}</option>
						{/each}
					</select>
				</label>
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
					<label>
						<span>Audio file</span>
						<select bind:value={cue.assetKey}>
							<option value="">Choose uploaded audio</option>
							{#each audioAssets() as asset (asset.assetKey)}
								<option value={asset.assetKey}>{asset.assetKey}</option>
							{/each}
						</select>
					</label>
					<label>
						<span>Normal audience</span>
						<select bind:value={cue.defaultAudience}>
							<option value="all">All players</option>
							<option value="game_masters">Game masters</option>
							<option value="team">A selected team</option>
							<option value="player">A selected player</option>
						</select>
					</label>
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

	input:not([type='checkbox']),
	select {
		width: 100%;
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.55rem;
	}

	.check {
		display: flex;
		min-height: 44px;
		align-items: center;
		gap: 0.5rem;
		border: 1px solid #b89b6d;
		padding: 0.55rem;
	}

	.choices {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem 0.9rem;
		margin-top: 0.35rem;
	}

	.choices label {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
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
