<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Gamepad2, LogOut, ScrollText, Settings, ShieldCheck, Swords } from '@lucide/svelte';
	import AppNav from '$lib/components/AppNav.svelte';
	import Button from '$lib/components/Button.svelte';
	import IconButton from '$lib/components/IconButton.svelte';
	import ConnectionBadge from '$lib/features/shell/components/ConnectionBadge.svelte';
	import Field from '$lib/components/Field.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { fieldErrorOrSummary, toFormError, type FormError } from '$lib/forms/errors';
	import type { AuthResponse } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let credentials = $state({ username: '', password: '' });
	let busy = $state(false);
	let loginError = $state<FormError | null>(null);
	let recoveryAvailable = $state(false);
	let recoveryOpen = $state(false);
	let recovery = $state({
		username: '',
		displayName: '',
		password: '',
		trustedLanAcknowledged: false,
		confirmation: ''
	});
	let recoveryError = $state<FormError | null>(null);

	const liveRoute = $derived(
		/^\/admin\/games\/[^/]+\/(overview|players|chat|activity|finish|summary)/.test(
			page.url.pathname
		)
	);
	const editorRoute = $derived(
		page.url.pathname === '/admin/rulesets/new' ||
			/^\/admin\/rulesets\/[^/]+\/edit\//.test(page.url.pathname)
	);
	const current = $derived.by(() => {
		if (page.url.pathname.startsWith('/admin/rulesets')) return 'rulesets';
		if (page.url.pathname.startsWith('/admin/approvals')) return 'approvals';
		if (page.url.pathname.startsWith('/admin/settings')) return 'settings';
		return 'games';
	});
	const navigation = [
		{ id: 'games', label: 'Games', href: resolve('/admin/games'), icon: Gamepad2 },
		{ id: 'rulesets', label: 'Rulesets', href: resolve('/admin/rulesets'), icon: ScrollText },
		{ id: 'approvals', label: 'Approvals', href: resolve('/admin/approvals'), icon: ShieldCheck },
		{ id: 'settings', label: 'Settings', href: resolve('/admin/settings/network'), icon: Settings }
	];

	onMount(() => {
		if (auth.isPlayer) {
			auth.clear();
			toasts.info('Sign in with a game-master account.');
		}
		void loadRecoveryAvailability();
	});

	async function loadRecoveryAvailability() {
		try {
			const status = await api<{ ownerRecoveryAvailable: boolean }>('/setup/status');
			recoveryAvailable = status.ownerRecoveryAvailable;
		} catch {
			recoveryAvailable = false;
		}
	}

	async function login(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		loginError = null;
		try {
			const response = await api<AuthResponse>('/auth/game-master/login', {
				method: 'POST',
				...jsonBody(credentials)
			});
			auth.save(response);
		} catch (caught) {
			loginError = toFormError(caught, 'Sign-in failed. Try again.');
		} finally {
			busy = false;
		}
	}

	async function logout() {
		try {
			await api('/auth/logout', { method: 'POST' });
		} finally {
			auth.clear();
			await goto(resolve('/'));
		}
	}

	async function recoverOwnership(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		recoveryError = null;
		try {
			const response = await api<AuthResponse>('/setup/owner-recovery', {
				method: 'POST',
				...jsonBody(recovery)
			});
			auth.clear();
			auth.save(response);
			recoveryOpen = false;
		} catch (caught) {
			recoveryError = toFormError(caught, 'Ownership could not be recovered.');
		} finally {
			busy = false;
		}
	}
</script>

{#if !auth.isGameMaster}
	<main class="login-page">
		<section class="login-panel" aria-labelledby="login-heading">
			<div class="seal" aria-hidden="true"><Swords size={32} /></div>
			<p class="eyebrow">Game master</p>
			<h1 id="login-heading">Sign in</h1>
			<p class="muted">Manage games on this host.</p>
			<form onsubmit={login}>
				<Field
					label="Username"
					name="username"
					bind:value={credentials.username}
					autocomplete="username"
					required
				/>
				<Field
					label="Password"
					name="password"
					type="password"
					bind:value={credentials.password}
					autocomplete="current-password"
					error={fieldErrorOrSummary(loginError, 'password')}
					required
				/>
				<Button type="submit" loading={busy}>Sign in</Button>
			</form>
			{#if recoveryAvailable}
				<Button variant="ghost" onclick={() => (recoveryOpen = true)}
					>Recover local ownership</Button
				>
			{/if}
			<a href={resolve('/')}>Return to join page</a>
		</section>
	</main>
{:else if liveRoute || editorRoute}
	{@render children()}
{:else}
	<div class="management-shell">
		<AppNav items={navigation} {current} label="Management" />
		<header class="management-header">
			<a class="product" href={resolve('/admin/games')}><Swords size={23} /> Social Games Hoster</a>
			<div class="header-actions">
				<ConnectionBadge />
				<IconButton label="Sign out" variant="ghost" onclick={logout}>
					{#snippet icon()}<LogOut size={20} />{/snippet}
				</IconButton>
			</div>
		</header>
		<main class="management-content">
			{@render children()}
		</main>
	</div>
{/if}

<Dialog
	open={recoveryOpen}
	title="Recover local ownership"
	description="This replaces obsolete owner accounts on this computer. A recovery backup is created before any change."
	close={() => (recoveryOpen = false)}
>
	<form id="owner-recovery-form" onsubmit={recoverOwnership}>
		<Field
			label="New owner username"
			name="recovery-username"
			bind:value={recovery.username}
			required
		/>
		<Field
			label="New owner display name"
			name="recovery-display-name"
			bind:value={recovery.displayName}
			required
		/>
		<Field
			label="New owner password"
			name="recovery-password"
			type="password"
			bind:value={recovery.password}
			autocomplete="new-password"
			required
		/>
		<label
			><input type="checkbox" bind:checked={recovery.trustedLanAcknowledged} required /> I understand
			and trust this local network.</label
		>
		<Field
			label="Type &quot;RECOVER OWNERSHIP&quot; to confirm"
			name="recovery-confirmation"
			bind:value={recovery.confirmation}
			required
		/>
		{#if recoveryError}<p class="field-error" role="alert">{recoveryError.message}</p>{/if}
	</form>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (recoveryOpen = false)}>Cancel</Button>
		<Button
			variant="danger"
			loading={busy}
			onclick={() =>
				(document.getElementById('owner-recovery-form') as HTMLFormElement)?.requestSubmit()}
			>Recover ownership</Button
		>
	{/snippet}
</Dialog>

<style>
	.login-page {
		display: grid;
		min-height: 100dvh;
		place-items: center;
		padding: var(--space-4);
	}

	.login-panel {
		width: min(100%, 27rem);
		border: var(--border-strong);
		background: var(--paper);
		box-shadow: var(--shadow);
		padding: var(--space-6);
		text-align: center;
	}

	.seal {
		display: grid;
		width: 4rem;
		height: 4rem;
		place-items: center;
		border: 3px double var(--gold-light);
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--paper-light);
		margin: 0 auto var(--space-3);
	}

	form {
		display: grid;
		gap: var(--space-4);
		margin-block: var(--space-5);
		text-align: start;
	}

	form :global(button) {
		width: 100%;
	}

	.field-error {
		color: var(--danger);
		margin: 0;
	}

	.management-shell {
		min-height: 100dvh;
		padding-block-end: calc(4rem + env(safe-area-inset-bottom));
	}

	.management-header {
		position: sticky;
		z-index: var(--layer-sticky);
		inset-block-start: 0;
		display: flex;
		min-height: 4rem;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: rgb(247 231 196 / 94%);
		padding: var(--space-2) max(var(--space-4), env(safe-area-inset-right)) var(--space-2)
			max(var(--space-4), env(safe-area-inset-left));
		backdrop-filter: blur(8px);
	}

	.product {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--ink);
		font-family: var(--font-display);
		font-size: 0.8rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.management-content {
		width: min(100%, 76rem);
		margin-inline: auto;
		padding: clamp(var(--space-4), 4vw, var(--space-7));
	}

	@media (min-width: 64rem) {
		.management-shell {
			padding-block-end: 0;
			padding-inline-start: 13rem;
		}

		.management-header {
			padding-inline: var(--space-5);
		}
	}
</style>
