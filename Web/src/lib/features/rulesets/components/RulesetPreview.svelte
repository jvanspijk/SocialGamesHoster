<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import type {
		RulesetAsset,
		RulesetDefinition,
		RulesetPreviewMode,
		RulesetPreviewRequest,
		RulesetPreviewResponse
	} from '$lib/api/types';

	let {
		open,
		close,
		definition,
		assets,
		dirty,
		loadPreview,
		onresult = () => {}
	}: {
		open: boolean;
		close: () => void;
		definition: RulesetDefinition;
		assets: RulesetAsset[];
		dirty: boolean;
		loadPreview: (request: RulesetPreviewRequest) => Promise<RulesetPreviewResponse>;
		onresult?: (response: RulesetPreviewResponse) => void;
	} = $props();

	const modes: Array<{ id: RulesetPreviewMode; label: string }> = [
		{ id: 'role', label: 'Role card' },
		{ id: 'phases', label: 'Phase flow' },
		{ id: 'composition', label: 'Player setup' },
		{ id: 'chat', label: 'Chat' },
		{ id: 'media', label: 'Media' }
	];
	let mode = $state<RulesetPreviewMode>('role');
	let roleId = $state('');
	let phaseId = $state('');
	let playerCount = $state(3);
	let assetKey = $state('');
	let preview = $state<RulesetPreviewResponse | null>(null);
	let loading = $state(false);
	let error = $state('');
	let sequence = 0;

	$effect(() => {
		if (!definition.roles.some((role) => role.id === roleId))
			roleId = definition.roles[0]?.id ?? '';
		if (!definition.phases.some((phase) => phase.id === phaseId))
			phaseId = definition.phases[0]?.id ?? '';
		if (!assets.some((asset) => asset.assetKey === assetKey)) assetKey = assets[0]?.assetKey ?? '';
		if (
			playerCount < definition.metadata.minPlayers ||
			playerCount > definition.metadata.maxPlayers
		) {
			playerCount = definition.metadata.minPlayers;
		}
	});

	$effect(() => {
		if (!open) return;
		const currentMode = mode;
		const currentRole = roleId;
		const currentPhase = phaseId;
		const currentPlayers = playerCount;
		const currentAsset = assetKey;
		JSON.stringify(definition);
		JSON.stringify(assets.map((asset) => [asset.assetKey, asset.checksum, asset.displayName]));
		const timer = setTimeout(
			() =>
				void refresh({
					mode: currentMode,
					roleId: currentRole,
					phaseId: currentPhase,
					playerCount: currentPlayers,
					assetKey: currentAsset
				}),
			100
		);
		return () => clearTimeout(timer);
	});

	async function refresh(request: RulesetPreviewRequest) {
		const current = ++sequence;
		loading = true;
		error = '';
		try {
			const response = await loadPreview(request);
			if (current === sequence) {
				preview = response;
				onresult(response);
			}
		} catch (caught) {
			if (current === sequence) {
				error = caught instanceof Error ? caught.message : 'The preview could not be loaded.';
			}
		} finally {
			if (current === sequence) loading = false;
		}
	}
</script>

<Sheet
	{open}
	title="Preview ruleset"
	description="Check the working ruleset in the forms players and game masters will use."
	{close}
>
	<div class="preview-workspace">
		<p class:unsaved={dirty} class="source-status" role="status">
			{dirty ? 'Previewing unsaved working changes' : 'Previewing the saved ruleset'}
		</p>
		<div class="mode-list" role="group" aria-label="Preview type">
			{#each modes as item (item.id)}
				<Button
					variant={mode === item.id ? 'primary' : 'secondary'}
					onclick={() => (mode = item.id)}
				>
					{item.label}
				</Button>
			{/each}
		</div>

		<div class="preview-controls">
			{#if mode === 'role' || mode === 'chat'}
				<label>
					<span>{mode === 'role' ? 'Role' : 'Audience role'}</span>
					<select bind:value={roleId}>
						{#each definition.roles as role (role.id)}<option value={role.id}>{role.name}</option
							>{/each}
					</select>
				</label>
			{/if}
			{#if mode === 'phases' || mode === 'chat'}
				<label>
					<span>Phase</span>
					<select bind:value={phaseId}>
						{#if definition.phases.length === 0}<option value="">No phase</option>{/if}
						{#each definition.phases as phase (phase.id)}<option value={phase.id}
								>{phase.name}</option
							>{/each}
					</select>
				</label>
			{/if}
			{#if mode === 'composition'}
				<label>
					<span>Player count</span>
					<input type="number" min="1" max="30" bind:value={playerCount} />
				</label>
			{/if}
			{#if mode === 'media'}
				<label>
					<span>Media item</span>
					<select bind:value={assetKey}>
						{#each assets as asset (asset.assetKey)}<option value={asset.assetKey}
								>{asset.displayName}</option
							>{/each}
					</select>
				</label>
			{/if}
		</div>

		<section class="preview-stage" aria-live="polite" aria-busy={loading}>
			{#if loading && !preview}<p>Loading preview…</p>
			{:else if error}<p class="error" role="alert">{error}</p>
			{:else if preview?.empty}<div class="empty">
					<h2>Nothing to preview yet</h2>
					<p>{preview.message}</p>
				</div>
			{:else if preview?.mode === 'role' && preview.role}
				<article class="role-card">
					{#if preview.media}<ProtectedMedia
							src={preview.media.preview}
							kind="image"
							alt={preview.media.accessibilityText || preview.media.displayName}
						/>{/if}
					<p class="eyebrow">{preview.role.teamName}</p>
					<h2>{preview.role.name}</h2>
					<p>{preview.role.description}</p>
					<h3>How to win</h3>
					<p>{preview.role.winCondition || 'No win condition added.'}</p>
					<h3>Abilities</h3>
					{#if preview.role.abilities.length}<ul>
							{#each preview.role.abilities as ability, index (`${ability.name}:${index}`)}<li>
									<strong>{ability.name}</strong><span>{ability.description}</span>
								</li>{/each}
						</ul>{:else}<p>No special abilities.</p>{/if}
				</article>
			{:else if preview?.mode === 'phases' && preview.phases}
				<div class="phase-flow">
					<h2>Game-master phase sequence</h2>
					<ol>
						{#each preview.phases as phase, index (`${phase.name}:${index}`)}<li
								class:selected={phase.selected}
							>
								<h3>{phase.name}</h3>
								<p>{phase.description}</p>
								<small
									>{phase.startsRound
										? 'Starts a new round'
										: 'Continues the round'}{phase.suggestedDurationSeconds
										? ` · ${Math.round(phase.suggestedDurationSeconds / 60)} minute timer`
										: ''}{phase.sound ? ` · Sound: ${phase.sound}` : ''}</small
								>{#if phase.media}<ProtectedMedia
										src={phase.media.preview}
										kind={phase.media.kind}
										alt={phase.media.accessibilityText || phase.media.displayName}
										controls={phase.media.kind === 'audio'}
									/>{/if}
							</li>{/each}
					</ol>
				</div>
			{:else if preview?.mode === 'composition'}
				<div class="composition-result">
					<h2>{preview.feasible ? 'Setup is feasible' : 'Setup is not feasible'}</h2>
					<p>{preview.message}</p>
					{#if preview.roles?.length}<ul>
							{#each preview.roles as role, index (`${role.teamName}:${role.name}:${index}`)}<li>
									<span>{role.name}<small>{role.teamName}</small></span><strong
										>× {role.count}</strong
									>
								</li>{/each}
						</ul>{/if}
				</div>
			{:else if preview?.mode === 'chat' && preview.rooms}
				<div class="chat-preview">
					<h2>{preview.audience} · {preview.phase}</h2>
					<ul>
						{#each preview.rooms as room, index (`${room.kind}:${room.name}:${index}`)}<li>
								<span><strong>{room.name}</strong><small>{room.kind}</small></span><span
									class="permissions"
									>{room.visible ? 'Visible' : 'Hidden'} · {room.readable
										? 'Can read'
										: 'Cannot read'} · {room.sendable
										? 'Can post'
										: 'Cannot post'}{room.messageRestriction === 'emoji_only'
										? ' · Emoji only'
										: ''}</span
								>
							</li>{/each}
					</ul>
				</div>
			{:else if preview?.mode === 'media' && preview.media}
				<div class="media-preview">
					<h2>{preview.media.displayName}</h2>
					<p>{preview.media.accessibilityText || 'No accessibility text added.'}</p>
					{#if preview.contexts?.length}
						<h3>In the game</h3>
						<div class="media-contexts">
							{#each preview.contexts as context, index (`${context.kind}:${context.title}:${index}`)}
								<article class:cover-context={context.kind === 'cover'} class="media-context">
									{#if context.kind === 'cover'}
										<ProtectedMedia
											src={preview.media.preview}
											kind={preview.media.kind}
											alt={preview.media.accessibilityText || preview.media.displayName}
											controls={preview.media.kind === 'audio'}
										/>
										<div>
											<p class="eyebrow">{context.label}</p>
											<h4>{context.title}</h4>
											{#if context.description}<p>
													{context.description}
												</p>{/if}{#if context.detail}<strong>{context.detail}</strong>{/if}
										</div>
									{:else}
										<div class="context-media">
											<ProtectedMedia
												src={preview.media.preview}
												kind={preview.media.kind}
												alt={preview.media.accessibilityText || preview.media.displayName}
												controls={preview.media.kind === 'audio'}
											/>
										</div>
										<div>
											<p class="eyebrow">{context.label}</p>
											<h4>{context.title}</h4>
											{#if context.description}<p>
													{context.description}
												</p>{/if}{#if context.detail}<strong>{context.detail}</strong>{/if}
										</div>
									{/if}
								</article>
							{/each}
						</div>
					{:else}
						<div class="unused-media">
							<ProtectedMedia
								src={preview.media.preview}
								kind={preview.media.kind}
								alt={preview.media.accessibilityText || preview.media.displayName}
								controls={preview.media.kind === 'audio'}
							/>
							<p>This item is not currently used in the ruleset.</p>
						</div>
					{/if}
				</div>
			{/if}
		</section>
	</div>
	{#snippet actions()}<Button variant="secondary" onclick={close}>Close preview</Button>{/snippet}
</Sheet>

<style>
	.preview-workspace,
	.preview-stage,
	.role-card,
	.phase-flow,
	.composition-result,
	.chat-preview,
	.media-preview {
		display: grid;
		gap: var(--space-3);
	}
	.media-contexts,
	.unused-media {
		display: grid;
		gap: var(--space-4);
	}
	.media-context {
		display: grid;
		grid-template-columns: minmax(6rem, 10rem) minmax(0, 1fr);
		gap: var(--space-3);
		align-items: center;
		border-block-start: var(--border-subtle);
		padding-block-start: var(--space-3);
	}
	.media-context > div {
		display: grid;
		gap: var(--space-2);
	}
	.media-context h4,
	.media-context p {
		margin: 0;
	}
	.cover-context {
		grid-template-columns: 1fr;
		border: var(--border-strong);
		background: var(--paper-deep);
		padding: var(--space-4);
	}
	.context-media :global(img) {
		aspect-ratio: 1;
		width: 100%;
		object-fit: cover;
		border: var(--border-strong);
	}
	.context-media :global(audio) {
		width: 100%;
	}
	.source-status {
		margin: 0;
		color: var(--ink-soft);
	}
	.source-status.unsaved {
		border-inline-start: 0.25rem solid var(--warning);
		padding-inline-start: var(--space-2);
		color: var(--ink);
	}
	.mode-list {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}
	.preview-controls {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 12rem), 1fr));
		gap: var(--space-3);
	}
	.preview-controls label,
	.role-card li span,
	.composition-result li span,
	.chat-preview li > span:first-child {
		display: grid;
		gap: var(--space-1);
	}
	.preview-controls label > span,
	.eyebrow {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
	}
	.preview-controls select,
	.preview-controls input {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-2);
	}
	.preview-stage {
		min-height: 20rem;
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-4);
	}
	.preview-stage h2,
	.preview-stage h3,
	.preview-stage p,
	.eyebrow {
		margin: 0;
	}
	.role-card :global(img),
	.media-preview :global(img) {
		width: min(100%, 28rem);
		max-height: 18rem;
		object-fit: cover;
		border: var(--border-strong);
	}
	.role-card ul,
	.phase-flow ol,
	.composition-result ul,
	.chat-preview ul {
		display: grid;
		gap: var(--space-2);
		margin: 0;
		padding-inline-start: var(--space-5);
	}
	.phase-flow li,
	.composition-result li,
	.chat-preview li {
		border-block-start: var(--border-subtle);
		padding: var(--space-3) 0;
	}
	.phase-flow li.selected {
		border-inline-start: 0.25rem solid var(--crimson);
		padding-inline-start: var(--space-3);
	}
	.composition-result li,
	.chat-preview li {
		display: flex;
		justify-content: space-between;
		gap: var(--space-3);
	}
	.permissions,
	small {
		color: var(--ink-soft);
	}
	.error {
		color: var(--danger);
	}
	@media (max-width: 47.99rem) {
		.composition-result li,
		.chat-preview li {
			display: grid;
		}
		.media-context {
			grid-template-columns: 1fr;
		}
	}
</style>
