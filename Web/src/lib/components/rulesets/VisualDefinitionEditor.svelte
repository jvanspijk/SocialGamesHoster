<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import Field from '$lib/components/Field.svelte';
	import RoomPermissionEditor from './RoomPermissionEditor.svelte';
	import SelectorEditor from './SelectorEditor.svelte';
	import type {
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
		definition,
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
			description: ''
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
	<div class="section-title">
		<div>
			<h2>Teams</h2>
			<p class="muted">The main sides or factions in the game.</p>
		</div>
		<Button variant="secondary" onclick={addTeam}>Add team</Button>
	</div>
	<div class="cards">
		{#each definition.teams as team, index (team.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{team.name || 'Unnamed team'}</h3>
					<button class="remove" onclick={() => removeAt(definition.teams, index)}>Remove</button>
				</div>
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

	<div class="section-title subsection">
		<div>
			<h2>Categories</h2>
			<p class="muted">Optional labels such as Investigative or Support.</p>
		</div>
		<Button variant="secondary" onclick={addCategory}>Add category</Button>
	</div>
	<div class="cards compact">
		{#each definition.categories as category, index (category.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{category.name || 'Unnamed category'}</h3>
					<button class="remove" onclick={() => removeAt(definition.categories, index)}
						>Remove</button
					>
				</div>
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
	<div class="section-title">
		<div>
			<h2>Abilities</h2>
			<p class="muted">Reusable powers that can be assigned to roles.</p>
		</div>
		<Button variant="secondary" onclick={addAbility}>Add ability</Button>
	</div>
	<div class="cards compact">
		{#each definition.abilities as ability, index (ability.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{ability.name || 'Unnamed ability'}</h3>
					<button class="remove" onclick={() => removeAt(definition.abilities, index)}
						>Remove</button
					>
				</div>
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
			</article>
		{/each}
	</div>

	<div class="section-title subsection">
		<div>
			<h2>Roles</h2>
			<p class="muted">What each player may be assigned and how that role wins.</p>
		</div>
		<Button variant="secondary" onclick={addRole}>Add role</Button>
	</div>
	<div class="cards">
		{#each definition.roles as role, index (role.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{role.name || 'Unnamed role'}</h3>
					<button class="remove" onclick={() => removeAt(definition.roles, index)}>Remove</button>
				</div>
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
	<div class="section-title">
		<div>
			<h2>Phases</h2>
			<p class="muted">The ordered steps a game master advances through.</p>
		</div>
		<Button variant="secondary" onclick={addPhase}>Add phase</Button>
	</div>
	<div class="cards">
		{#each definition.phases as phase, index (phase.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{phase.order}. {phase.name || 'Unnamed phase'}</h3>
					<button class="remove" onclick={() => removeAt(definition.phases, index)}>Remove</button>
				</div>
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
	<div class="section-title">
		<div>
			<h2>Player-count bands</h2>
			<p class="muted">Define how many slots are filled for every supported party size.</p>
		</div>
		<Button variant="secondary" onclick={addBand}>Add band</Button>
	</div>
	<div class="cards">
		{#each definition.compositionBands as band, bandIndex (band.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{band.minPlayers}–{band.maxPlayers} players</h3>
					<button class="remove" onclick={() => removeAt(definition.compositionBands, bandIndex)}
						>Remove</button
					>
				</div>
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
					<strong>Role slots</strong>
					<button class="add-small" onclick={() => addSlot(bandIndex)}>Add slot</button>
				</div>
				{#each band.slots as slot, slotIndex (slot.id)}
					<div class="nested">
						<div class="card-heading">
							<strong>{slot.label || 'Unnamed slot'}</strong>
							<button class="remove" onclick={() => removeAt(band.slots, slotIndex)}>Remove</button>
						</div>
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

	<div class="section-title subsection">
		<div>
			<h2>Conditional changes</h2>
			<p class="muted">Adjust slots when a particular role appears.</p>
		</div>
		<Button variant="secondary" onclick={addModifier}>Add condition</Button>
	</div>
	<div class="cards">
		{#each definition.compositionModifiers as modifier, modifierIndex (modifier.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{modifier.id}</h3>
					<button
						class="remove"
						onclick={() => removeAt(definition.compositionModifiers, modifierIndex)}>Remove</button
					>
				</div>
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
					<strong>Slot changes</strong>
					<button class="add-small" onclick={() => addAdjustment(modifierIndex)}>Add change</button>
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
	<div class="section-title">
		<div>
			<h2>Starting knowledge</h2>
			<p class="muted">Choose what one group learns about another group.</p>
		</div>
		<Button variant="secondary" onclick={addKnowledge}>Add knowledge rule</Button>
	</div>
	<div class="cards">
		{#each definition.knowledgeRules as rule, index (rule)}
			<article class="item-card">
				<div class="card-heading">
					<h3>Knowledge rule {index + 1}</h3>
					<button class="remove" onclick={() => removeAt(definition.knowledgeRules, index)}
						>Remove</button
					>
				</div>
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
	<div class="section-title">
		<div>
			<h2>Normal chat settings</h2>
			<p class="muted">These settings apply unless a phase changes them.</p>
		</div>
	</div>
	<div class="cards">
		<article class="item-card">
			<div class="card-heading">
				<h3>General room</h3>
				{#if definition.chat.defaultPolicy.general}
					<button class="remove" onclick={() => delete definition.chat.defaultPolicy.general}
						>Disable room</button
					>
				{:else}
					<button class="add-small" onclick={() => addDefaultRoom('general')}>Enable room</button>
				{/if}
			</div>
			{#if definition.chat.defaultPolicy.general}
				<RoomPermissionEditor policy={definition.chat.defaultPolicy.general} />
			{/if}
		</article>
		<article class="item-card">
			<div class="card-heading">
				<h3>Private player messages</h3>
				{#if definition.chat.defaultPolicy.playerDm}
					<button class="remove" onclick={() => delete definition.chat.defaultPolicy.playerDm}
						>Disable room</button
					>
				{:else}
					<button class="add-small" onclick={() => addDefaultRoom('playerDm')}>Enable room</button>
				{/if}
			</div>
			{#if definition.chat.defaultPolicy.playerDm}
				<RoomPermissionEditor policy={definition.chat.defaultPolicy.playerDm} />
			{/if}
		</article>
		{#each definition.teams as team (team.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{team.name} team room</h3>
					{#if definition.chat.defaultPolicy.teams[team.id]}
						<button
							class="remove"
							onclick={() => delete definition.chat.defaultPolicy.teams[team.id]}
							>Disable room</button
						>
					{:else}
						<button class="add-small" onclick={() => addDefaultTeam(team.id)}>Enable room</button>
					{/if}
				</div>
				{#if definition.chat.defaultPolicy.teams[team.id]}
					<RoomPermissionEditor policy={definition.chat.defaultPolicy.teams[team.id]} />
				{/if}
			</article>
		{/each}
	</div>

	<div class="section-title subsection">
		<div>
			<h2>Phase changes</h2>
			<p class="muted">Temporarily override only the settings that need to change.</p>
		</div>
	</div>
	<div class="cards">
		{#each definition.phases as phase (phase.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{phase.name}</h3>
					{#if definition.chat.phaseOverrides[phase.id]}
						<button class="remove" onclick={() => delete definition.chat.phaseOverrides[phase.id]}
							>Remove changes</button
						>
					{:else}
						<button class="add-small" onclick={() => addPhaseOverride(phase.id)}
							>Add phase changes</button
						>
					{/if}
				</div>
				{#if definition.chat.phaseOverrides[phase.id]}
					{@const override = definition.chat.phaseOverrides[phase.id]}
					<div class="override-room">
						<div class="card-heading">
							<strong>General room</strong>
							{#if override.general}
								<button class="remove" onclick={() => delete override.general}>Clear</button>
							{:else}
								<button class="add-small" onclick={() => addOverrideRoom(phase.id, 'general')}
									>Change</button
								>
							{/if}
						</div>
						{#if override.general}<RoomPermissionEditor policy={override.general} partial />{/if}
					</div>
					<div class="override-room">
						<div class="card-heading">
							<strong>Private player messages</strong>
							{#if override.playerDm}
								<button class="remove" onclick={() => delete override.playerDm}>Clear</button>
							{:else}
								<button class="add-small" onclick={() => addOverrideRoom(phase.id, 'playerDm')}
									>Change</button
								>
							{/if}
						</div>
						{#if override.playerDm}<RoomPermissionEditor policy={override.playerDm} partial />{/if}
					</div>
					{#each definition.teams as team (team.id)}
						<div class="override-room">
							<div class="card-heading">
								<strong>{team.name} room</strong>
								{#if override.teams?.[team.id]}
									<button class="remove" onclick={() => delete override.teams?.[team.id]}
										>Clear</button
									>
								{:else}
									<button class="add-small" onclick={() => addOverrideTeam(phase.id, team.id)}
										>Change</button
									>
								{/if}
							</div>
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
	<div class="section-title">
		<div>
			<h2>Achievements</h2>
			<p class="muted">Awards a game master can give after a game.</p>
		</div>
		<Button variant="secondary" onclick={addAchievement}>Add achievement</Button>
	</div>
	<div class="cards">
		{#each definition.achievements as achievement, index (achievement.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{achievement.name || 'Unnamed achievement'}</h3>
					<button class="remove" onclick={() => removeAt(definition.achievements, index)}
						>Remove</button
					>
				</div>
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
	<div class="section-title">
		<div>
			<h2>Audio cues</h2>
			<p class="muted">Named sounds a game master or phase can play for selected listeners.</p>
		</div>
		<Button variant="secondary" onclick={addAudioCue}>Add audio cue</Button>
	</div>
	<div class="cards">
		{#each definition.audioCues as cue, index (cue.id)}
			<article class="item-card">
				<div class="card-heading">
					<h3>{cue.name || 'Unnamed cue'}</h3>
					<button class="remove" onclick={() => removeAt(definition.audioCues, index)}
						>Remove</button
					>
				</div>
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
	.section-title,
	.card-heading,
	.nested-heading,
	.inline-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
	}

	.section-title h2,
	.section-title p,
	.card-heading h3 {
		margin: 0;
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
	.nested-heading > strong,
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

		.section-title,
		.inline-row {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
