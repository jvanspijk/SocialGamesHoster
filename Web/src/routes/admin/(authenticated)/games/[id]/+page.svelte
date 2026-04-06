<script lang="ts">
	import { enhance } from '$app/forms';
	import LedgerTable from '$lib/components/LedgerTable.svelte';
	import SecondaryButton from '$lib/components/SecondaryButton.svelte';
	import AdminTimer from '$lib/components/AdminTimer.svelte';

	import type { GetGamePlayersResponse } from '$lib/client/GameSessions/GetGamePlayers.js';
	import { StartNewRound } from '$lib/client/GameSessions/StartNewRound.js';
	import { invalidateAll } from '$app/navigation';
	import HUDFooter from '$lib/components/HUDFooter.svelte';
	import BackLink from '$lib/components/BackLink.svelte';
	import { authHub } from '$lib/client/Auth/AuthHub.svelte';
	//import { FinishGameSession } from '$lib/client/GameSessions/FinishGameSession';
	import { StartGameSession } from '$lib/client/GameSessions/StartGameSession';
	import ConfirmationModal from '$lib/components/ConfirmationModal.svelte';
	import TimerControlModal from '$lib/components/TimerControlModal.svelte';
	import { DeletePlayer } from '$lib/client/Players/DeletePlayer';
	import ChatToggle from '$lib/components/ChatToggle.svelte';
	import AdminChatChannel from '$lib/components/AdminChatChannel.svelte';

	let { data } = $props();

	let pendingRoles = $state<Record<number, number | null>>({});
	let winnerIds = $state<Set<number>>(new Set());
	let isSaving = $state(false);

	let showKickConfirm = $state(false);
	let playerToKick: GetGamePlayersResponse | null = $state(null);
	let showTimerModal = $state(false);
	let isChatOpen = $state(false);
	let hasNewMessages = $state(false);

	const isDirty = $derived(Object.keys(pendingRoles).length > 0);

	function openChat() {
		isChatOpen = true;
		hasNewMessages = false;
	}

	function closeChat() {
		isChatOpen = false;
	}

	authHub.onEvent('PlayerLoggedIn', (event) => {
		console.debug(`Player ${event.playerId} logged in.`);
		const isPlayerPresent = data.players.some((p) => p.id === event.playerId);
		if (!isPlayerPresent) {
			invalidateAll();
		}
	});

	function toggleWinner(id: number) {
		if (winnerIds.has(id)) winnerIds.delete(id);
		else winnerIds.add(id);
	}

	function handleKickClick(player: GetGamePlayersResponse) {
		playerToKick = player;
		showKickConfirm = true;
	}

	async function confirmKick() {
		if (!playerToKick) return;

		await DeletePlayer(fetch, { id: playerToKick.id });
		showKickConfirm = false;
		playerToKick = null;
		invalidateAll(); // Refresh data
	}
</script>

<div class="page-main">
	{#if showKickConfirm && playerToKick}
		<ConfirmationModal
			title="Confirm Kick"
			message={`Are you sure you want to remove ${playerToKick.name} from the game? This action is immediate and cannot be undone.`}
			onConfirm={confirmKick}
			onCancel={() => (showKickConfirm = false)}
		/>
	{/if}
	<div class="parchment-container">
		<header class="admin-header">
			<BackLink href="/admin/games" pageName="Games Overview"></BackLink>
			<h1>Game Management</h1>
		</header>

		<section class="player-management">
			<h2>Player Roster</h2>
			{#await data.streamed.fullPlayers}
				<div class="loading-overlay">Retrieving player details...</div>
			{:then fullPlayers}
				{@const playerLookup = new Map(
					fullPlayers.filter((p) => p !== null).map((p) => [p!.id, p!])
				)}

				<LedgerTable data={data.players} enableFilter={false}>
					{#snippet columns()}
						<tr>
							<th>Name</th>
							<th>Role</th>
							<th>Actions</th>
						</tr>
					{/snippet}

					{#snippet rows(player: GetGamePlayersResponse)}
						{@const fullData = playerLookup.get(player.id)}
						<tr>
							<td data-label="Name">
								<span class={winnerIds.has(player.id) ? 'winner-highlight' : ''}>
									{player.name}
								</span>
							</td>
							<td data-label="Role">
								{#if fullData}
									<select
										class="ledger-select-inline"
										value={pendingRoles[player.id] ?? fullData.role?.id ?? ''}
										onchange={(e) => (pendingRoles[player.id] = Number(e.currentTarget.value))}
									>
										<option value="">No Role</option>
										{#each data.roles as role (role.id)}
											<option value={role.id}>{role.name}</option>
										{/each}
									</select>
								{:else}
									<span class="status-error">Load Error</span>
								{/if}
							</td>
							<td data-label="Fate">
								<div class="action-cell">
									<button
										type="button"
										class="action-btn {winnerIds.has(player.id) ? 'active' : ''}"
										title="Mark as Winner"
										onclick={() => toggleWinner(player.id)}>🏆</button
									>
									<button
										type="button"
										class="action-btn-danger"
										title="Remove Player from game"
										onclick={() => handleKickClick(player)}>❌</button
									>
								</div>
							</td>
						</tr>
					{/snippet}
				</LedgerTable>
			{/await}
		</section>

		<footer class="admin-footer">
			{#if isDirty}
				<div class="sticky-action-bar">
					<p>You have unsaved role assignments.</p>
					<form
						method="POST"
						action="?/saveRoles"
						use:enhance={() => {
							isSaving = true;
							return async ({ update }) => {
								await update({ reset: false });
								pendingRoles = {};
								isSaving = false;
							};
						}}
					>
						<input
							type="hidden"
							name="updates"
							value={JSON.stringify(
								Object.entries(pendingRoles).map(([id, roleId]) => ({ id: Number(id), roleId }))
							)}
						/>
						<SecondaryButton type="submit" enabled={!isSaving}>
							{isSaving ? 'Saving...' : 'Save All Roles'}
						</SecondaryButton>
						<button type="button" class="text-link" onclick={() => (pendingRoles = {})}
							>Discard</button
						>
					</form>
				</div>
			{/if}

			{#if winnerIds.size > 0}
				<form method="POST" action="?/declareWinners" use:enhance>
					{#each Array.from(winnerIds) as id (id)}
						<input type="hidden" name="winnerIds" value={id} />
					{/each}
					<button type="submit" class="btn-finalize">Declare {winnerIds.size} Winner(s)</button>
				</form>
			{/if}
		</footer>
	</div>

	<div class="game-controls">
		<div class="button-group">
			{#if data.gameSession.status === 'Not started'}
				<SecondaryButton onclick={() => StartGameSession(fetch, { gameId: data.gameSession.id })}
					>Start game</SecondaryButton
				>
			{:else if data.gameSession.status === 'Running'}
				<SecondaryButton variant="danger" onclick={() => alert('not implemented')}
					>Stop game</SecondaryButton
				>
				<SecondaryButton
					onclick={() => StartNewRound(fetch, { gameId: data.gameSession.id, newPhaseId: 1 })}
					>Start new round</SecondaryButton
				>
			{:else if data.gameSession.status === 'Finished'}
				<p>Winners:</p>
			{/if}
		</div>

		<h3>Game Status: {data.gameSession.status}</h3>
	</div>

	<HUDFooter>
		<ChatToggle hasNewMessage={hasNewMessages} onclick={openChat} />
		<span>Round #{data.currentRound?.id}</span>
		{#if data.timer}
			<AdminTimer timer={data.timer} />
		{/if}
		<SecondaryButton onclick={() => (showTimerModal = true)}>Adjust Timer</SecondaryButton>
	</HUDFooter>

	{#if showTimerModal}
		<TimerControlModal onclose={() => (showTimerModal = false)} />
	{/if}

	<AdminChatChannel gameId={data.gameSession.id} isOpen={isChatOpen} onClose={closeChat} />
</div>

<style>
	.admin-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;
	}

	.loading-overlay {
		padding: 40px;
		text-align: center;
		font-style: italic;
		color: var(--color-text-muted);
	}

	.action-btn {
		background: none;
		border: 1px solid var(--color-border);
		cursor: pointer;
		font-size: 1.2rem;
		padding: 5px 10px;
		filter: grayscale(1);
	}
	.action-btn.active {
		filter: grayscale(0);
		background: var(--color-surface-soft);
	}
	.action-btn-danger {
		background: none;
		border: 1px solid var(--color-accent-strong);
		cursor: pointer;
		font-size: 1.2rem;
		padding: 5px 10px;
		color: var(--color-accent-strong);
		margin-left: 5px;
	}
	.action-btn-danger:hover {
		background: var(--color-accent-strong);
		color: var(--color-on-accent);
	}
	.action-cell {
		display: flex;
	}

	.winner-highlight {
		font-weight: bold;
		color: var(--color-warning-text);
		text-decoration: underline;
	}

	.sticky-action-bar {
		position: sticky;
		bottom: 0;
		background: var(--color-surface);
		border: 2px solid var(--color-border);
		padding: 15px;
		margin-top: 20px;
		text-align: center;
		box-shadow: 0 -5px 15px rgba(0, 0, 0, 0.1);
	}

	.btn-finalize {
		background: var(--color-border);
		color: var(--color-surface);
		font-family: var(--font-heading);
		padding: 15px 30px;
		border: none;
		cursor: pointer;
		width: 100%;
		margin-top: 20px;
	}

	.text-link {
		background: none;
		border: none;
		color: var(--color-accent-strong);
		text-decoration: underline;
		cursor: pointer;
		margin-left: 10px;
	}

	.ledger-select-inline {
		background: var(--color-surface-soft);
		border: 1px solid var(--color-border);
		font-family: inherit;
		padding: 4px;
        min-width: fit-content;
	}

	.game-controls {
		margin: 10px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
	}

    option {
        color: var(--color-on-accent);
    }
</style>
