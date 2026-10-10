<script lang="ts">
	import { tick } from 'svelte';
	import BellRing from '@lucide/svelte/icons/bell-ring';
	import Megaphone from '@lucide/svelte/icons/megaphone';
	import RotateCcw from '@lucide/svelte/icons/rotate-ccw';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import RecordItem from '$lib/components/RecordItem.svelte';
	import RecordList from '$lib/components/RecordList.svelte';
	import AnnouncementSheet from './AnnouncementSheet.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { AdminGameView } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { view, disabled = false }: { view: AdminGameView; disabled?: boolean } = $props();
	let open = $state(false);
	let pending = $state('');
	let error = $state('');
	let sendAction: HTMLDivElement;
	const latest = $derived(view.attentionItems[0]);
	const announcement = $derived(latest?.trackingCleared ? undefined : latest);
	const players = $derived(announcement?.unacknowledgedPlayers ?? []);
	const busy = $derived(disabled || pending !== '');
	const hasPlayers = $derived(
		view.participants.some((player) => !['kicked', 'left'].includes(player.status))
	);

	async function command(action: 'remind' | 'clear') {
		if (!announcement || busy) return;
		pending = action;
		error = '';
		const gameId = view.game.id;
		let cleared = false;
		try {
			await api(`/games/${gameId}/announcements/${announcement.id}/${action}`, {
				method: 'POST',
				...jsonBody({})
			});
			cleared = action === 'clear';
			toasts.success(action === 'remind' ? 'Reminder sent.' : 'Announcement cleared.');
			try {
				await gameState.refreshAdmin(gameId);
			} catch (caught) {
				error = errorMessage(caught, 'The game could not be refreshed. Reload the page.');
			}
		} catch (caught) {
			error = errorMessage(
				caught,
				action === 'remind'
					? 'The reminder could not be sent.'
					: 'The announcement could not be cleared.'
			);
		} finally {
			pending = '';
			if (cleared) {
				await tick();
				sendAction?.querySelector('button')?.focus({ preventScroll: true });
			}
		}
	}
</script>

<aside aria-label="Announcements">
	<Panel title="Announcements" fill>
		<div class="announcement-panel">
			<div class="send-action" bind:this={sendAction}>
				<Button
					variant="secondary"
					disabled={busy || !hasPlayers}
					onclick={() => {
						error = '';
						open = true;
					}}
				>
					<Megaphone size={18} /> Send announcement
				</Button>
			</div>
			<!-- svelte-ignore a11y_no_noninteractive_tabindex (Keyboard users need to scroll the pending-player list.) -->
			<div
				class="announcement-state"
				aria-live="polite"
				role="region"
				aria-label="Announcement acknowledgement status"
				tabindex={players.length ? 0 : undefined}
			>
				<ErrorNotice message={error} />
				{#if announcement}
					<p class="message"><strong>Message:</strong> <em>"{announcement.content}"</em></p>
					{#if players.length}
						<h2>Not acknowledged ({players.length})</h2>
						<RecordList>
							{#each players as player (player.participantId)}
								<RecordItem
									><span>Player {player.playerNumber} · {player.displayName}</span></RecordItem
								>
							{/each}
						</RecordList>
					{:else}
						<p>
							{announcement.acknowledgementCount === announcement.recipientTotal
								? 'All players acknowledged.'
								: 'No players are waiting to acknowledge.'}
						</p>
					{/if}
				{/if}
			</div>
			<div class="panel-actions">
				{#if announcement}
					<Button
						variant="secondary"
						disabled={busy || players.length === 0}
						loading={pending === 'remind'}
						onclick={() => command('remind')}
					>
						<BellRing size={18} /> Send reminder
					</Button>
					<Button
						variant="ghost"
						disabled={busy}
						loading={pending === 'clear'}
						onclick={() => command('clear')}
					>
						<RotateCcw size={18} /> Clear
					</Button>
				{/if}
			</div>
		</div>
	</Panel>
</aside>
<AnnouncementSheet {open} {view} close={() => (open = false)} />

<style>
	aside {
		min-width: 0;
		min-height: 0;
	}
	.announcement-panel {
		height: 24rem;
		display: grid;
		grid-template-rows: auto minmax(0, 1fr) 7rem;
		gap: var(--space-3);
	}
	.announcement-state {
		overflow-y: auto;
		overflow-wrap: anywhere;
		min-height: 0;
	}
	.message {
		white-space: pre-wrap;
		margin-block: 0 var(--space-4);
	}
	h2 {
		font-size: var(--font-size-lg);
		margin-block: 0 var(--space-2);
	}
	.panel-actions {
		display: grid;
		align-content: end;
		gap: var(--space-2);
	}
	.panel-actions > :global(*) {
		width: 100%;
	}
	.send-action > :global(*) {
		width: 100%;
	}
</style>
