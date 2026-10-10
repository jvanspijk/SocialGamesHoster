<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { AdminGameView, Participant } from '$lib/api/types';
	import { playerActionOptions, type PlayerAction } from '../playerActions';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let {
		open,
		view,
		participantId,
		close
	}: {
		open: boolean;
		view: AdminGameView;
		participantId: string;
		close: () => void;
	} = $props();

	let achievementId = $state('');
	let action = $state<PlayerAction | ''>('');
	let note = $state('');
	let outcome = $state<Participant['outcome']>('unset');
	let busy = $state(false);
	let error = $state('');
	$effect(() => {
		if (open) {
			action = '';
			achievementId = '';
			note = '';
			outcome = selected?.outcome ?? 'unset';
			error = '';
		}
	});
	const selected = $derived(view.participants.find((player) => player.id === participantId));
	const actionOptions = $derived(selected ? playerActionOptions(view, selected) : []);
	const achievement = $derived(
		view.ruleset.achievements?.find((item) => item.id === achievementId)
	);
	const existingAward = $derived(
		view.awards?.find(
			(award) => award.profileId === selected?.profileId && award.achievementId === achievementId
		)
	);
	const selectedName = $derived(selected?.gameAlias || selected?.displayNameSnapshot || '');
	const allowed = $derived(!!selected && actionOptions.some((option) => option.value === action));
	const canSubmit = $derived(
		!!action &&
			allowed &&
			!!selected &&
			(action !== 'achievement' || !!achievement) &&
			(action !== 'eliminate' || selected.status === 'active') &&
			(action !== 'reinstate' || selected.status === 'eliminated')
	);
	const submitLabel = $derived(
		action === 'achievement'
			? existingAward
				? 'Revoke achievement'
				: 'Award achievement'
			: action === 'outcome'
				? 'Save outcome'
				: action
					? `${action === 'eliminate' ? 'Eliminate' : 'Reinstate'} ${selectedName || 'player'}`
					: 'Choose an action'
	);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (busy || !canSubmit || !selected) return;
		busy = true;
		error = '';
		const name = selectedName;
		const revoking = !!existingAward;
		try {
			if (action === 'achievement') {
				if (existingAward) {
					await api(`/games/${view.game.id}/achievement-awards/${existingAward.id}`, {
						method: 'DELETE'
					});
				} else {
					await api(`/games/${view.game.id}/achievement-awards`, {
						method: 'POST',
						...jsonBody({ participantId: selected.id, achievementId, note })
					});
				}
			} else if (action === 'outcome') {
				await api(`/games/${view.game.id}/outcomes`, {
					method: 'PUT',
					...jsonBody({ outcomes: [{ participantId: selected.id, outcome }] })
				});
			} else {
				await api(`/games/${view.game.id}/participants/${selected.id}/${action}`, {
					method: 'POST',
					...jsonBody({})
				});
			}
			toasts.success(
				action === 'achievement'
					? `Achievement ${revoking ? 'revoked for' : 'awarded to'} ${name}.`
					: action === 'outcome'
						? `Outcome saved for ${name}.`
						: `${name} ${action === 'eliminate' ? 'eliminated' : 'reinstated'}.`
			);
			close();
			void gameState
				.refreshAdmin(view.game.id)
				.catch((caught) =>
					toasts.error(
						errorMessage(caught, 'The action succeeded, but the game could not be refreshed.')
					)
				);
		} catch (caught) {
			error = errorMessage(caught, 'The player could not be updated.');
		} finally {
			busy = false;
		}
	}
</script>

<Sheet
	{open}
	title="Player actions"
	description={selected ? `Player ${selected.playerNumber} · ${selectedName}` : ''}
	{close}
>
	<form id="player-action-form" onsubmit={submit}>
		<ErrorNotice message={error} />
		{#if !selected || actionOptions.length === 0 || (action && !allowed)}
			<p>This action is unavailable in the current game state.</p>
		{:else if selected}
			<SelectField
				label="Action"
				name="player-action"
				bind:value={action}
				options={[{ value: '', label: 'Choose an action' }, ...actionOptions]}
				disabled={busy}
				onchange={() => (error = '')}
			/>
			<p>
				<strong>{selectedName}</strong> · {selected.status === 'eliminated'
					? 'Eliminated'
					: 'In play'}
			</p>
			{#if action === 'eliminate'}
				<p>
					{selectedName} will be eliminated but remain in the game. You can reinstate them later.
				</p>
			{:else if action === 'reinstate'}
				<p>{selectedName} will return to active play.</p>
			{:else if action === 'achievement'}
				<SelectField
					label="Achievement"
					name="action-achievement"
					bind:value={achievementId}
					required
					disabled={busy}
					options={[
						{ value: '', label: 'Choose an achievement' },
						...view.ruleset.achievements.map((item) => ({ value: item.id, label: item.name }))
					]}
				/>
				{#if achievement}
					<p>{achievement.description}</p>
					{#if existingAward}<p>
							{selectedName} already has this achievement. You can revoke it.
						</p>
					{:else}<Field
							label="Note (optional)"
							name="achievement-note"
							bind:value={note}
							disabled={busy}
						/>{/if}
				{/if}
			{:else if action === 'outcome'}
				<SelectField
					label="Outcome"
					name="action-outcome"
					bind:value={outcome}
					disabled={busy}
					options={[
						{ value: 'unset', label: 'Unset' },
						{ value: 'win', label: 'Win' },
						{ value: 'loss', label: 'Loss' },
						{ value: 'draw', label: 'Draw' }
					]}
				/>
			{/if}
		{/if}
	</form>
	{#snippet actions()}
		<Button variant="ghost" disabled={busy} onclick={close}>Cancel</Button>
		<Button
			loading={busy}
			disabled={!canSubmit}
			onclick={() =>
				(document.getElementById('player-action-form') as HTMLFormElement)?.requestSubmit()}
			>{submitLabel}</Button
		>
	{/snippet}
</Sheet>

<style>
	form {
		display: grid;
		gap: var(--space-4);
	}
	p {
		margin: 0;
		color: var(--text-secondary);
		overflow-wrap: anywhere;
	}
</style>
