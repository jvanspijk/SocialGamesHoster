<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Archive, Copy, History, Play, Plus, Trash2 } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
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
	let busy = $state(false);
	let form = $state({ name: '', rulesetVersionId: '' });
	let formError = $state<FormError | null>(null);

	const visibleGames = $derived(games.filter((game) => showArchived || game.status !== 'archived'));
	const readyRulesets = $derived(
		rulesets.filter((ruleset) => ruleset.latestPublishedVersion && !ruleset.archived)
	);

	onMount(load);

	async function load() {
		loading = true;
		try {
			[games, rulesets] = await Promise.all([
				api<Game[]>('/games'),
				api<RulesetSummary[]>('/rulesets')
			]);
			form.rulesetVersionId ||= readyRulesets[0]?.latestPublishedVersion ?? '';
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
			toasts.success('Game created.');
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

	async function duplicate(game: Game) {
		try {
			const created = await api<Game>(`/games/${game.id}/duplicate`, {
				method: 'POST',
				...jsonBody({})
			});
			games = [created, ...games];
			toasts.success('Game duplicated.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game could not be duplicated.'));
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

<header class="page-heading">
	<div>
		<p class="eyebrow">Management</p>
		<h1>Games</h1>
		<p>Create a game, resume play, or revisit a finished game.</p>
	</div>
	<Button onclick={() => (createOpen = true)}><Plus size={18} /> New game</Button>
</header>

<label class="archive-filter">
	<input type="checkbox" bind:checked={showArchived} />
	Show archived games
</label>

{#if loading}
	<p role="status">Loading games…</p>
{:else if visibleGames.length === 0}
	<Panel variant="focal">
		<div class="empty">
			<Play size={38} strokeWidth={1.5} aria-hidden="true" />
			<h2>No games yet</h2>
			<p>Choose a ready ruleset to prepare your first game.</p>
			<Button onclick={() => (createOpen = true)}>New game</Button>
		</div>
	</Panel>
{:else}
	<div class="game-list">
		{#each visibleGames as game (game.id)}
			<article>
				<div class="game-mark" class:archived={game.status === 'archived'} aria-hidden="true">
					{#if game.status === 'archived'}<Archive size={23} />{:else}<Play size={23} />{/if}
				</div>
				<div class="game-copy">
					<div class="title-line">
						<h2>{game.name}</h2>
						<span class:live={['lobby', 'running', 'paused'].includes(game.status)}
							>{gameStatusLabel(game.status)}</span
						>
					</div>
					<p>
						{game.startedAt
							? `Started ${new Date(game.startedAt).toLocaleDateString()}`
							: 'Not started'}
					</p>
				</div>
				<div class="row-actions">
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
					<a class="open-link" href={openGame(game)}>
						{#if game.status === 'archived'}<History size={17} /> View summary{:else}<Play
								size={17}
							/> Open{/if}
					</a>
					<button type="button" onclick={() => duplicate(game)}><Copy size={17} /> Duplicate</button
					>
					{#if ['draft', 'archived'].includes(game.status)}
						<button class="danger" type="button" onclick={() => (deleteTarget = game)}>
							<Trash2 size={17} /> Delete
						</button>
					{/if}
				</div>
			</article>
		{/each}
	</div>
{/if}

<Dialog
	open={createOpen}
	title="New game"
	description="Choose a ready ruleset. The game keeps its own frozen copy."
	close={() => (createOpen = false)}
>
	<form id="new-game-form" class="dialog-form" onsubmit={createGame}>
		<ErrorNotice error={formError} />
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
				bind:value={form.rulesetVersionId}
				aria-invalid={fieldError(formError, 'rulesetVersionId') ? 'true' : undefined}
				required
			>
				<option value="" disabled>Choose a ruleset</option>
				{#each readyRulesets as ruleset (ruleset.id)}
					<option value={ruleset.latestPublishedVersion}>{ruleset.name}</option>
				{/each}
			</select>
			{#if fieldError(formError, 'rulesetVersionId')}<small
					>{fieldError(formError, 'rulesetVersionId')}</small
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
	.page-heading {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-4);
		margin-block-end: var(--space-5);
	}

	.page-heading h1,
	.page-heading p {
		margin: 0;
	}

	.eyebrow {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	.archive-filter {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
		margin-block-end: var(--space-4);
	}

	.game-list {
		border-block: var(--border-subtle);
	}

	article {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-4);
		border-block-end: var(--border-subtle);
		padding: var(--space-4) 0;
	}

	article:last-child {
		border-block-end: 0;
	}

	.game-mark {
		display: grid;
		width: 3rem;
		height: 3rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--paper-light);
	}

	.game-mark.archived {
		background: var(--ink-soft);
	}

	.title-line {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	h2,
	.game-copy p {
		margin: 0;
	}

	.title-line span {
		border: 1px solid var(--ink-faint);
		color: var(--ink-soft);
		font-size: 0.72rem;
		padding: 0.15rem 0.45rem;
	}

	.title-line span.live {
		border-color: var(--success);
		color: var(--success);
	}

	.row-actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-1);
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
		.page-heading {
			align-items: stretch;
			flex-direction: column;
		}

		.page-heading :global(button) {
			width: 100%;
		}

		article {
			grid-template-columns: auto minmax(0, 1fr);
			align-items: start;
		}

		.row-actions {
			grid-column: 1 / -1;
			justify-content: flex-start;
		}
	}
</style>
