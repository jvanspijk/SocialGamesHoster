<script lang="ts">
	import { Eye, EyeOff, ShieldCheck } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let {
		gameId,
		rolesVisible,
		presentation
	}: {
		gameId: string;
		rolesVisible: boolean;
		presentation: 'readiness' | 'heading';
	} = $props();

	let revealConfirmOpen = $state(false);
	let hideConfirmOpen = $state(false);
	let busy = $state(false);

	async function setRoleVisibility(visible: boolean) {
		busy = true;
		try {
			await api(`/games/${gameId}/role-visibility`, {
				method: 'PATCH',
				...jsonBody({ rolesVisible: visible })
			});
			await gameState.refreshAdmin(gameId);
			revealConfirmOpen = false;
			hideConfirmOpen = false;
			toasts.success(visible ? 'Roles are available.' : 'Roles are hidden.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Role visibility could not be changed.'));
		} finally {
			busy = false;
		}
	}
</script>

{#snippet status()}
	<strong>{rolesVisible ? 'Roles revealed' : 'Roles hidden'}</strong>
	<span class="visibility-description">
		{rolesVisible ? 'Players can see their own role.' : 'Roles are hidden from players.'}
	</span>
{/snippet}

{#snippet dialogs()}
	<Dialog
		open={revealConfirmOpen}
		title="Reveal roles?"
		description={presentation === 'readiness'
			? 'Players will be able to see their assigned role.'
			: ''}
		close={() => (revealConfirmOpen = false)}
	>
		{#if presentation === 'readiness'}
			<p>Players can still toggle the visibility of their role on their own screen.</p>
		{:else}
			<p>
				Players will be able to view their assigned roles. They can hide them again on their screen
				at any time.
			</p>
		{/if}
		{#snippet actions()}
			<Button variant="ghost" onclick={() => (revealConfirmOpen = false)}>Cancel</Button>
			<Button loading={busy} onclick={() => setRoleVisibility(true)}>Reveal roles</Button>
		{/snippet}
	</Dialog>

	<Dialog
		open={hideConfirmOpen}
		title="Hide roles?"
		description={presentation === 'readiness'
			? 'Players with the Role screen open will lose access immediately.'
			: ''}
		close={() => (hideConfirmOpen = false)}
	>
		{#if presentation === 'readiness'}
			<p>Role and knowledge data will be removed from player screens.</p>
		{:else}
			<p>Roles will be hidden on all player screens immediately.</p>
		{/if}
		{#snippet actions()}
			<Button variant="ghost" onclick={() => (hideConfirmOpen = false)}>Cancel</Button>
			<Button variant="danger" loading={busy} onclick={() => setRoleVisibility(false)}
				>Hide roles</Button
			>
		{/snippet}
	</Dialog>
{/snippet}

{#if presentation === 'readiness'}
	<li class="readiness-control">
		<ShieldCheck size={20} />
		<div>{@render status()}</div>
		{#if rolesVisible}
			<button class="readiness-action" type="button" onclick={() => (hideConfirmOpen = true)}
				>Hide</button
			>
		{:else}
			<button class="readiness-action" type="button" onclick={() => (revealConfirmOpen = true)}
				>Reveal roles</button
			>
		{/if}
		{@render dialogs()}
	</li>
{:else}
	<div class="heading-control">
		<div class="heading-status">
			{#if rolesVisible}<Eye size={20} />{:else}<EyeOff size={20} />{/if}
			<span>{@render status()}</span>
		</div>
		{#if rolesVisible}
			<Button variant="secondary" onclick={() => (hideConfirmOpen = true)}>Hide roles</Button>
		{:else}
			<Button onclick={() => (revealConfirmOpen = true)}>Reveal roles</Button>
		{/if}
		{@render dialogs()}
	</div>
{/if}

<style>
	.readiness-control {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-2);
		border: 0;
		padding: var(--space-3) 0;
	}

	.readiness-control strong,
	.visibility-description {
		display: block;
	}

	.visibility-description {
		color: var(--ink-soft);
		font-size: 0.82rem;
	}

	.readiness-action {
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
	}

	.heading-control {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		border: var(--border-subtle);
		background: rgb(255 249 230 / 58%);
		padding: var(--space-2);
	}

	.heading-status {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.heading-control strong {
		display: block;
	}

	@media (max-width: 63.99rem) {
		.heading-control {
			justify-content: space-between;
		}
	}

	@media (max-width: 47.99rem) {
		.heading-control {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
