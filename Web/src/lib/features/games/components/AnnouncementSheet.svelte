<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { AdminGameView, AdminAttentionSummary } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { open, view, close }: { open: boolean; view: AdminGameView; close: () => void } = $props();
	type Attachment = {
		source: 'none' | 'ruleset' | 'upload';
		key: string;
		file: File | null;
		description: string;
	};
	const attachmentKinds = ['image', 'audio'] as const;
	let content = $state('');
	let audience = $state<'all' | 'team' | 'player'>('all');
	let targetId = $state('');
	let cueKey = $state('');
	let media = $state<Record<'image' | 'audio', Attachment>>({
		image: { source: 'none', key: '', file: null, description: '' },
		audio: { source: 'none', key: '', file: null, description: '' }
	});
	let busy = $state(false);
	let error = $state('');
	$effect(() => {
		if (open) {
			content = '';
			audience = 'all';
			targetId = '';
			cueKey = '';
			media = {
				image: { source: 'none', key: '', file: null, description: '' },
				audio: { source: 'none', key: '', file: null, description: '' }
			};
			error = '';
		}
	});
	const participants = $derived(
		view.participants.filter((player) => !['kicked', 'left'].includes(player.status))
	);
	const recipients = $derived(
		participants.filter(
			(player) =>
				audience === 'all' ||
				(audience === 'player'
					? player.id === targetId
					: view.ruleset.roles.find((role) => role.id === player.roleKey)?.teamId === targetId)
		)
	);
	const audienceName = $derived(
		audience === 'all'
			? 'everyone'
			: audience === 'team'
				? view.ruleset.teams.find((team) => team.id === targetId)?.name
				: participants.find((player) => player.id === targetId)?.gameAlias ||
					participants.find((player) => player.id === targetId)?.displayNameSnapshot
	);
	const canSend = $derived(
		content.trim().length > 0 &&
			content.trim().length <= 1000 &&
			recipients.length > 0 &&
			['lobby', 'running', 'paused'].includes(view.game.status)
	);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy || !canSend) return;
		busy = true;
		error = '';
		try {
			const request = {
				content,
				audience,
				targetId: audience === 'all' ? '' : targetId,
				cueKey,
				imageAssetKey: media.image.source === 'ruleset' ? media.image.key : '',
				imageDescription: media.image.description,
				audioAssetKey: media.audio.source === 'ruleset' ? media.audio.key : '',
				audioAlternative: media.audio.description
			};
			let body: ReturnType<typeof jsonBody> | { body: FormData } = jsonBody(request);
			if (media.image.source === 'upload' || media.audio.source === 'upload') {
				const form = new FormData();
				for (const [key, value] of Object.entries(request)) form.set(key, value);
				for (const kind of ['image', 'audio'] as const) {
					if (media[kind].source === 'upload' && media[kind].file)
						form.set(`${kind}File`, media[kind].file!);
				}
				body = { body: form };
			}
			const sent = await api<AdminAttentionSummary>(`/games/${view.game.id}/announcements`, {
				method: 'POST',
				...body
			});
			toasts.success(
				`Sent to ${sent.recipientTotal} ${sent.recipientTotal === 1 ? 'player' : 'players'}.`
			);
			close();
			void gameState
				.refreshAdmin(view.game.id)
				.catch((caught) =>
					toasts.error(
						errorMessage(caught, 'The announcement was sent, but the game could not be refreshed.')
					)
				);
		} catch (caught) {
			error = errorMessage(caught, 'The announcement could not be sent.');
		} finally {
			busy = false;
		}
	}
</script>

<Sheet {open} title="Send announcement" {close}>
	<form id="announcement-form" onsubmit={submit}>
		<ErrorNotice message={error} />
		<Field
			label="Announcement message"
			name="announcement"
			bind:value={content}
			multiline
			required
			disabled={busy}
			help={`${content.length} / 1000 characters`}
		/>
		<SelectField
			label="Recipients"
			name="announcement-audience"
			bind:value={audience}
			disabled={busy}
			options={[
				{ value: 'all', label: 'All players' },
				{ value: 'team', label: 'Selected team' },
				{ value: 'player', label: 'Selected player' }
			]}
			onchange={() => (targetId = '')}
		/>
		{#if audience === 'team'}
			<SelectField
				label="Team"
				name="announcement-team"
				bind:value={targetId}
				required
				disabled={busy}
				options={[
					{ value: '', label: 'Choose a team' },
					...view.ruleset.teams.map((team) => ({ value: team.id, label: team.name }))
				]}
			/>
		{:else if audience === 'player'}
			<SelectField
				label="Player"
				name="announcement-player"
				bind:value={targetId}
				required
				disabled={busy}
				options={[
					{ value: '', label: 'Choose a player' },
					...participants.map((player) => ({
						value: player.id,
						label: `Player ${player.playerNumber} · ${player.gameAlias || player.displayNameSnapshot}`
					}))
				]}
			/>
		{/if}
		<p role="status">
			{recipients.length}
			{recipients.length === 1 ? 'player' : 'players'} will receive this announcement.
		</p>
		<details>
			<summary>Images, audio and sound cues</summary>
			<div class="attachments">
				{#each attachmentKinds as kind (kind)}
					<fieldset>
						<legend>{kind === 'image' ? 'Image (optional)' : 'Audio (optional)'}</legend>
						<SelectField
							label={kind === 'image' ? 'Image source' : 'Audio source'}
							name={`announcement-${kind}-source`}
							bind:value={media[kind].source}
							disabled={busy}
							options={[
								{ value: 'none', label: `No ${kind}` },
								{ value: 'ruleset', label: 'Choose from ruleset' },
								{ value: 'upload', label: 'Upload for this announcement' }
							]}
							onchange={() => {
								media[kind].key = '';
								media[kind].file = null;
								media[kind].description = '';
							}}
						/>
						{#if media[kind].source === 'ruleset'}
							<SelectField
								label={kind === 'image' ? 'Ruleset image' : 'Ruleset audio'}
								name={`announcement-${kind}`}
								bind:value={media[kind].key}
								required
								disabled={busy}
								options={[
									{ value: '', label: `Choose ${kind === 'image' ? 'an image' : 'audio'}` },
									...view.assets
										.filter((asset) => asset.kind === kind)
										.map((asset) => ({ value: asset.assetKey, label: asset.displayName }))
								]}
								onchange={() =>
									(media[kind].description =
										view.assets.find((asset) => asset.assetKey === media[kind].key)
											?.accessibilityText ?? '')}
							/>
						{:else if media[kind].source === 'upload'}
							<label class="form-field">
								<span>{kind === 'image' ? 'Image file' : 'Audio file'}</span>
								<input
									type="file"
									accept={kind === 'image'
										? 'image/jpeg,image/png,image/webp'
										: 'audio/mpeg,audio/mp4,audio/ogg,audio/wav'}
									required
									disabled={busy}
									onchange={(event) => (media[kind].file = event.currentTarget.files?.[0] ?? null)}
								/>
							</label>
						{/if}
						{#if media[kind].source !== 'none'}
							<Field
								label={kind === 'image' ? 'Image description' : 'Audio alternative'}
								name={`announcement-${kind}-description`}
								bind:value={media[kind].description}
								required
								disabled={busy}
							/>
						{/if}
					</fieldset>
				{/each}
				{#if view.ruleset.audioCues?.length > 0}
					<SelectField
						label="Sound cue (optional)"
						name="announcement-cue"
						bind:value={cueKey}
						disabled={busy}
						options={[
							{ value: '', label: 'No sound cue' },
							...view.ruleset.audioCues.map((cue) => ({ value: cue.id, label: cue.name }))
						]}
					/>
				{/if}
			</div>
		</details>
	</form>
	{#snippet actions()}
		<Button variant="ghost" disabled={busy} onclick={close}>Cancel</Button>
		<Button
			loading={busy}
			disabled={!canSend}
			onclick={() =>
				(document.getElementById('announcement-form') as HTMLFormElement)?.requestSubmit()}
		>
			{audienceName ? `Send to ${audienceName}` : 'Send announcement'}
		</Button>
	{/snippet}
</Sheet>

<style>
	form,
	.attachments,
	fieldset {
		display: grid;
		gap: var(--space-4);
		min-width: 0;
	}
	p {
		margin: 0;
		color: var(--text-secondary);
	}
	summary {
		min-height: var(--target-size);
		align-content: center;
		cursor: pointer;
		color: var(--action-dark);
	}
	.attachments {
		margin-block-start: var(--space-3);
	}
	fieldset {
		margin: 0;
		border: var(--border-subtle);
		padding: var(--space-3);
	}
	legend {
		font-weight: 700;
	}
	input[type='file'] {
		max-width: 100%;
	}
</style>
