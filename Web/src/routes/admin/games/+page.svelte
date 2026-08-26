<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { DoorOpen, Flag, Plus, Trash2, XCircle } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import { fieldError, toFormError, type FormError } from '$lib/forms/errors';
	import { gameStatusLabel } from '$lib/gamePresentation';
	import type { Game, RulesetSummary } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let games = $state<Game[]>([]);
	let rulesets = $state<RulesetSummary[]>([]);
	let loading = $state(true);
	let showArchived = $state(false);
	let createOpen = $state(false);
	let deleteTarget = $state<Game | null>(null);
	let cancelTarget = $state<Game | null>(null);
	let busy = $state(false);
	let form = $state({ name: '', rulesetId: '' });
	let formError = $state<FormError | null>(null);

	const visibleGames = $derived(games.filter((game) => showArchived || game.status !== 'archived'));
	const readyRulesets = $derived(rulesets.filter((ruleset) => ruleset.status === 'valid'));

	onMount(load);

	async function load() {
		loading = true;
		try {
			[games, rulesets] = await Promise.all([
				api<Game[]>('/games'),
				api<RulesetSummary[]>('/rulesets')
			]);
			form.rulesetId ||= readyRulesets[0]?.id ?? '';
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Games could not be loaded.'), {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function createGame(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		formError = null;
		try {
			const created = await api<Game>('/games', {
				method: 'POST',
				...jsonBody(form)
			});
			createOpen = false;
			toasts.success('Game created and lobby opened.');
			await goto(resolve(`/admin/games/${created.id}/overview`));
		} catch (caught) {
			const nextError = toFormError(caught, 'The game could not be created.');
			if (nextError.kind === 'validation') {
				formError = nextError;
				return;
			}
			toasts.error(nextError.message);
		} finally {
			busy = false;
		}
	}

	async function openLobby(game: Game) {
		try {
			const updated = await api<Game>(`/games/${game.id}/open-lobby`, {
				method: 'POST',
				...jsonBody({})
			});
			games = games.map((item) => (item.id === game.id ? { ...item, ...updated } : item));
			toasts.success('Lobby opened.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The lobby could not be opened.'));
		}
	}

	async function cancelGame() {
		if (!cancelTarget) return;
		busy = true;
		try {
			await api(`/games/${cancelTarget.id}/cancel`, {
				method: 'POST',
				...jsonBody({})
			});
			games = games.filter((game) => game.id !== cancelTarget?.id);
			cancelTarget = null;
			toasts.success('Game cancelled.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game could not be cancelled.'));
		} finally {
			busy = false;
		}
	}

	async function endGame(game: Game) {
		try {
			await api(`/games/${game.id}/completion/start`, { method: 'POST', ...jsonBody({}) });
			await goto(resolve(`/admin/games/${game.id}/finish/outcomes`));
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The completion flow could not be started.'));
		}
	}

	async function remove() {
		if (!deleteTarget) return;
		busy = true;
		try {
			await api(`/games/${deleteTarget.id}`, {
				method: 'DELETE',
				...jsonBody({ confirmation: `DELETE ${deleteTarget.id}` })
			});
			games = games.filter((game) => game.id !== deleteTarget?.id);
			deleteTarget = null;
			toasts.success('Game deleted.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game could not be deleted.'));
		} finally {
			busy = false;
		}
	}

	function openGame(game: Game) {
		return game.status === 'archived'
			? resolve(`/admin/games/${game.id}/summary`)
			: resolve(`/admin/games/${game.id}/overview`);
	}
</script>

<PageHeading
	eyebrow="Management"
	title="Games"
	description="Create a game, resume play, or revisit a finished game."
>
	{#snippet actions()}
		<div class="game-heading-actions">
			<Button onclick={() => (createOpen = true)}><Plus size={18} /> New game</Button>
		</div>
	{/snippet}
</PageHeading>

<label class="archive-filter">
	<input type="checkbox" bind:checked={showArchived} />
	Show archived games
</label>

{#if loading}
	<p role="status">Loading games…</p>
{:else if visibleGames.length === 0}
	<Panel variant="focal">
		<div class="empty">
			<DoorOpen size={38} strokeWidth={1.5} aria-hidden="true" />
			<h2>No games yet</h2>
			<p>Choose a ready ruleset to prepare your first game.</p>
			<Button onclick={() => (createOpen = true)}>New game</Button>
		</div>
	</Panel>
{:else}
	<div class="table-frame">
		<table>
			<caption>Games and their current state</caption>
			<thead>
				<tr>
					<th scope="col">Game</th>
					<th scope="col">Players</th>
					<th scope="col">Status</th>
					<th scope="col"><span class="sr-only">Actions</span></th>
				</tr>
			</thead>
			<tbody>
				{#each visibleGames as game (game.id)}
					<tr>
						<th scope="row" data-label="Game">
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
							<a class="game-name" href={openGame(game)}>{game.name}</a>
						</th>
						<td data-label="Players">{game.playerCount ?? 0}/{game.maxPlayers ?? '—'}</td>
						<td data-label="Status">
							<span
								class:live={['lobby', 'running', 'paused'].includes(game.status)}
								class="status"
							>
								{gameStatusLabel(game.status)}
							</span>
						</td>
						<td class="row-actions" data-label="Actions">
							{#if game.status === 'draft'}
								<button type="button" onclick={() => openLobby(game)}
									><DoorOpen size={17} /> Open lobby</button
								>
							{:else if game.status === 'lobby'}
								<button class="danger" type="button" onclick={() => (cancelTarget = game)}>
									<XCircle size={17} /> Cancel game
								</button>
							{:else if ['running', 'paused'].includes(game.status)}
								<button type="button" onclick={() => endGame(game)}
									><Flag size={17} /> End game</button
								>
							{:else if game.status === 'review'}
								<a href={resolve(`/admin/games/${game.id}/finish/outcomes`)}>Continue finishing</a>
							{:else if game.status === 'archived'}
								<a href={resolve(`/admin/games/${game.id}/summary`)}>View summary</a>
							{/if}
							{#if ['draft', 'archived'].includes(game.status)}
								<button class="danger" type="button" onclick={() => (deleteTarget = game)}>
									<Trash2 size={17} /> Delete
								</button>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}

<Dialog
	open={createOpen}
	title="New game"
	description="Choose a ready ruleset. The game keeps its own frozen copy."
	close={() => (createOpen = false)}
>
	<form id="new-game-form" class="dialog-form" onsubmit={createGame}>
		<ErrorNotice message={formError?.message} traceId={formError?.traceId} />
		<Field
			label="Game name"
			name="game-name"
			bind:value={form.name}
			error={fieldError(formError, 'name')}
			required
		/>
		<label>
			<span>Ruleset</span>
			<select
				bind:value={form.rulesetId}
				aria-invalid={fieldError(formError, 'rulesetId') ? 'true' : undefined}
				required
			>
				<option value="" disabled>Choose a ruleset</option>
				{#each readyRulesets as ruleset (ruleset.id)}
					<option value={ruleset.id}>{ruleset.name}</option>
				{/each}
			</select>
			{#if fieldError(formError, 'rulesetId')}<small>{fieldError(formError, 'rulesetId')}</small
				>{/if}
			{#if readyRulesets.length === 0}
				<small>No ready rulesets. Save a valid ruleset first.</small>
			{/if}
		</label>
	</form>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (createOpen = false)}>Cancel</Button>
		<Button
			type="submit"
			loading={busy}
			disabled={readyRulesets.length === 0}
			onclick={() => (document.getElementById('new-game-form') as HTMLFormElement)?.requestSubmit()}
		>
			Create game
		</Button>
	{/snippet}
</Dialog>

<Dialog
	open={cancelTarget !== null}
	title="Cancel game?"
	description={cancelTarget ? `"${cancelTarget.name}" will be permanently removed.` : ''}
	close={() => (cancelTarget = null)}
>
	<p>The player roster and lobby chat will also be deleted. This action cannot be undone.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (cancelTarget = null)}>Keep game</Button>
		<Button variant="danger" loading={busy} onclick={cancelGame}>Cancel game</Button>
	{/snippet}
</Dialog>

<Dialog
	open={deleteTarget !== null}
	title="Delete game?"
	description={deleteTarget ? `"${deleteTarget.name}" will be permanently deleted.` : ''}
	close={() => (deleteTarget = null)}
>
	<p>This action cannot be undone.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (deleteTarget = null)}>Cancel</Button>
		<Button variant="danger" loading={busy} onclick={remove}>Delete game</Button>
	{/snippet}
</Dialog>

<style>
	.archive-filter {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
		margin-block-end: var(--space-4);
	}

	@media (max-width: 47.99rem) {
		.game-heading-actions,
		.game-heading-actions :global(button) {
			width: 100%;
		}
	}

	.table-frame {
		border-block: var(--border-subtle);
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	caption {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
		white-space: nowrap;
	}

	th,
	td {
		border-block-end: var(--border-subtle);
		padding: var(--space-3) var(--space-2);
		text-align: left;
		vertical-align: middle;
	}

	thead th {
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.72rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	tbody tr:last-child > * {
		border-block-end: 0;
	}

	.game-name {
		color: var(--ink);
		font-family: var(--font-display);
		font-size: 1rem;
		text-decoration-thickness: 1px;
		text-underline-offset: 0.2em;
	}

	.status {
		display: inline-block;
		border: 1px solid var(--ink-faint);
		color: var(--ink-soft);
		font-size: 0.72rem;
		padding: 0.15rem 0.45rem;
	}

	.status.live {
		border-color: var(--success);
		color: var(--success);
	}

	.row-actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-1);
		text-align: right;
	}

	.row-actions a,
	.row-actions button {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-decoration: none;
	}

	.row-actions .danger {
		color: var(--danger);
	}

	.empty {
		padding: var(--space-6);
		text-align: center;
	}

	.dialog-form {
		display: grid;
		gap: var(--space-4);
	}

	.dialog-form label {
		display: grid;
		gap: var(--space-1);
	}

	.dialog-form label > span {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	select {
		width: 100%;
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		padding: var(--space-2);
	}

	small {
		color: var(--danger);
	}

	@media (max-width: 47.99rem) {
		.table-frame {
			border-block-end: 0;
		}

		table,
		tbody,
		tr,
		th,
		td {
			display: block;
		}

		thead {
			position: absolute;
			width: 1px;
			height: 1px;
			overflow: hidden;
			clip: rect(0 0 0 0);
		}

		tbody tr {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			gap: var(--space-2) var(--space-4);
			border-block-end: var(--border-subtle);
			padding: var(--space-4) 0;
		}

		tbody tr > * {
			border: 0;
			padding: 0;
		}

		tbody th {
			grid-column: 1 / -1;
		}

		td[data-label]::before {
			display: block;
			margin-block-end: var(--space-1);
			color: var(--ink-soft);
			content: attr(data-label);
			font-family: var(--font-display);
			font-size: 0.65rem;
			font-weight: 700;
			letter-spacing: 0.08em;
			text-transform: uppercase;
		}

		.row-actions {
			display: flex;
			grid-column: 1 / -1;
			justify-content: flex-start;
			text-align: left;
		}

		.row-actions::before {
			flex-basis: 100%;
		}
	}
</style>
