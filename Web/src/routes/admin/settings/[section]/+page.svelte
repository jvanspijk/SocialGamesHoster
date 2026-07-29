<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		DatabaseBackup,
		MonitorCog,
		Network,
		Plus,
		QrCode,
		RefreshCw,
		Shield,
		UserCog
	} from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import DisplaySettings from '$lib/components/DisplaySettings.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import { auth } from '$lib/state/auth.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type HostSettings = {
		port: number;
		bindAddress: string;
		preferredAdapter: string;
		trustedLanAcknowledged: boolean;
		automaticBackups: boolean;
		privateAddresses: Array<{ adapter: string; address: string }>;
		restartRequired: boolean;
	};
	type Backup = { id: string; size: number; modifiedAt: string; automatic: boolean };
	type GameMaster = {
		id: string;
		username: string;
		displayName: string;
		isOwner: boolean;
		active: boolean;
		lastLoginAt?: string;
	};

	const sections = [
		{ id: 'network', label: 'Network', icon: Network },
		{ id: 'phone-join', label: 'Phone join', icon: QrCode },
		{ id: 'game-masters', label: 'Game masters', icon: Shield },
		{ id: 'backups', label: 'Backups', icon: DatabaseBackup },
		{ id: 'diagnostics', label: 'Diagnostics', icon: MonitorCog },
		{ id: 'display', label: 'Display', icon: UserCog },
		{ id: 'account', label: 'Account', icon: UserCog }
	];

	let settings = $state<HostSettings | null>(null);
	let backups = $state<Backup[]>([]);
	let gameMasters = $state<GameMaster[]>([]);
	let diagnostics = $state<Record<string, unknown> | null>(null);
	let loading = $state(true);
	let busy = $state(false);
	let addMasterOpen = $state(false);
	let restoreBackupTarget = $state<Backup | null>(null);
	let restoreConfirmation = $state('');
	let accountForm = $state({ username: '', displayName: '', password: '' });

	const section = $derived(page.params.section ?? 'network');

	onMount(load);

	async function load() {
		loading = true;
		try {
			if (auth.isOwner && ['network', 'phone-join'].includes(section)) {
				settings = await api<HostSettings>('/owner/settings');
			} else if (auth.isOwner && section === 'backups') {
				[settings, backups] = await Promise.all([
					api<HostSettings>('/owner/settings'),
					api<Backup[]>('/owner/backups')
				]);
			} else if (auth.isOwner && section === 'game-masters') {
				gameMasters = await api<GameMaster[]>('/owner/game-masters');
			} else if (section === 'diagnostics') {
				diagnostics = await api<Record<string, unknown>>('/diagnostics/resources');
			}
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Settings could not be loaded.'), {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function saveSettings() {
		if (!settings) return;
		busy = true;
		try {
			settings = await api<HostSettings>('/owner/settings', {
				method: 'PATCH',
				...jsonBody(settings)
			});
			toasts.success(
				settings.restartRequired
					? 'Settings saved. Restart the host to apply them.'
					: 'Settings saved.'
			);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Settings could not be saved.'));
		} finally {
			busy = false;
		}
	}

	async function createBackup() {
		busy = true;
		try {
			await api('/owner/backups', { method: 'POST' });
			backups = await api<Backup[]>('/owner/backups');
			toasts.success('Backup created.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The backup could not be created.'));
		} finally {
			busy = false;
		}
	}

	async function addGameMaster(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		try {
			await api('/owner/game-masters', { method: 'POST', ...jsonBody(accountForm) });
			gameMasters = await api<GameMaster[]>('/owner/game-masters');
			accountForm = { username: '', displayName: '', password: '' };
			addMasterOpen = false;
			toasts.success('Game master added.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game master could not be added.'));
		} finally {
			busy = false;
		}
	}

	async function restoreBackup(event: SubmitEvent) {
		event.preventDefault();
		if (!restoreBackupTarget) return;
		busy = true;
		try {
			await api(`/owner/backups/${restoreBackupTarget.id}/restore`, {
				method: 'POST',
				...jsonBody({ confirmation: restoreConfirmation })
			});
			restoreBackupTarget = null;
			restoreConfirmation = '';
			toasts.info('Restore scheduled. The host will restart when it is ready.', {
				persistent: true
			});
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The backup could not be restored.'));
		} finally {
			busy = false;
		}
	}
</script>

<header class="page-heading">
	<p class="eyebrow">Host configuration</p>
	<h1>Settings</h1>
</header>

<div class="settings-layout">
	<nav aria-label="Settings sections">
		{#each sections as item (item.id)}
			{@const Icon = item.icon}
			<a class:active={section === item.id} href={resolve(`/admin/settings/${item.id}`)}>
				<Icon size={18} />
				{item.label}
			</a>
		{/each}
	</nav>

	<div class="settings-content">
		{#if loading}
			<p role="status">Loading settings…</p>
		{:else if !auth.isOwner && ['network', 'phone-join', 'game-masters', 'backups'].includes(section)}
			<Panel title="Owner access required" variant="focal">
				<p>Only the host owner can change this section.</p>
			</Panel>
		{:else if section === 'network' && settings}
			<Panel title="Network" description="Changes apply after restarting the host.">
				<div class="form-stack">
					<label>
						<span>Port</span>
						<input type="number" min="1" max="65535" bind:value={settings.port} />
					</label>
					<Field
						label="Bind address"
						name="bind-address"
						bind:value={settings.bindAddress}
						help="Leave blank to use the recommended private network address."
					/>
					<Field
						label="Preferred adapter"
						name="preferred-adapter"
						bind:value={settings.preferredAdapter}
					/>
					<label class="check"
						><input type="checkbox" bind:checked={settings.trustedLanAcknowledged} /> I trust this local
						network.</label
					>
					<Button loading={busy} onclick={saveSettings}>Save network settings</Button>
				</div>
			</Panel>
		{:else if section === 'phone-join' && settings}
			<Panel
				title="Phone join"
				description="Players scan this code while connected to the same private network."
				variant="focal"
			>
				<div class="join-layout">
					<img src="/api/app/v1/setup/join-qr" alt="QR code for the player join page" />
					<div>
						<h3>Available addresses</h3>
						{#each settings.privateAddresses as address (`${address.adapter}:${address.address}`)}
							<p>
								<strong>{address.adapter}</strong><br />http://{address.address}:{settings.port}
							</p>
						{/each}
					</div>
				</div>
			</Panel>
		{:else if section === 'game-masters'}
			<div class="section-header">
				<div>
					<h2>Game masters</h2>
					<p>People who can prepare and run games.</p>
				</div>
				<Button onclick={() => (addMasterOpen = true)}><Plus size={17} /> Add game master</Button>
			</div>
			<div class="record-list">
				{#each gameMasters as master (master.id)}
					<article>
						<div>
							<h3>{master.displayName}</h3>
							<p>@{master.username}</p>
						</div>
						<span>{master.isOwner ? 'Owner' : master.active ? 'Active' : 'Disabled'}</span>
					</article>
				{/each}
			</div>
		{:else if section === 'backups' && settings}
			<div class="section-header">
				<div>
					<h2>Backups</h2>
					<p>Create and restore recoverable host backups.</p>
				</div>
				<Button loading={busy} onclick={createBackup}
					><DatabaseBackup size={17} /> Create backup</Button
				>
			</div>
			<label class="check backup-toggle">
				<input type="checkbox" bind:checked={settings.automaticBackups} onchange={saveSettings} />
				Create automatic backups
			</label>
			<div class="record-list">
				{#each backups as backup (backup.id)}
					<article>
						<div>
							<h3>{backup.id}</h3>
							<p>
								{new Date(backup.modifiedAt).toLocaleString()} · {Math.ceil(backup.size / 1024)} KB
							</p>
						</div>
						<button
							type="button"
							onclick={() => {
								restoreBackupTarget = backup;
								restoreConfirmation = '';
							}}
						>
							Restore
						</button>
					</article>
				{/each}
			</div>
		{:else if section === 'diagnostics'}
			<div class="section-header">
				<div>
					<h2>Diagnostics</h2>
					<p>Reader-safe host resource information.</p>
				</div>
				<Button variant="secondary" onclick={load}><RefreshCw size={17} /> Refresh</Button>
			</div>
			<pre>{JSON.stringify(diagnostics, null, 2)}</pre>
		{:else if section === 'display'}
			<Panel
				title="Display"
				description="These preferences apply only on this device."
				variant="focal"
			>
				<DisplaySettings />
			</Panel>
		{:else if section === 'account'}
			<Panel title="Account" variant="focal">
				<dl>
					<div>
						<dt>Display name</dt>
						<dd>{auth.actor?.displayName}</dd>
					</div>
					<div>
						<dt>Access</dt>
						<dd>{auth.isOwner ? 'Host owner' : 'Game master'}</dd>
					</div>
				</dl>
			</Panel>
		{/if}
	</div>
</div>

<Dialog open={addMasterOpen} title="Add game master" close={() => (addMasterOpen = false)}>
	<form id="add-master-form" class="form-stack" onsubmit={addGameMaster}>
		<Field label="Username" name="master-username" bind:value={accountForm.username} required />
		<Field
			label="Display name"
			name="master-display-name"
			bind:value={accountForm.displayName}
			required
		/>
		<Field
			label="Temporary password"
			name="master-password"
			type="password"
			bind:value={accountForm.password}
			required
		/>
	</form>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (addMasterOpen = false)}>Cancel</Button>
		<Button
			loading={busy}
			onclick={() =>
				(document.getElementById('add-master-form') as HTMLFormElement)?.requestSubmit()}
		>
			Add game master
		</Button>
	{/snippet}
</Dialog>

<Dialog
	open={Boolean(restoreBackupTarget)}
	title="Restore backup"
	close={() => (restoreBackupTarget = null)}
>
	{#if restoreBackupTarget}
		<p>
			The host will create a rollback backup, restore <strong>{restoreBackupTarget.id}</strong>, and
			restart. Everyone will be disconnected.
		</p>
		<form id="restore-backup-form" class="form-stack" onsubmit={restoreBackup}>
			<Field
				label={`Type RESTORE ${restoreBackupTarget.id} to confirm`}
				name="restore-confirmation"
				bind:value={restoreConfirmation}
				autocomplete="off"
				required
			/>
		</form>
	{/if}
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (restoreBackupTarget = null)}>Cancel</Button>
		<Button
			variant="danger"
			loading={busy}
			onclick={() =>
				(document.getElementById('restore-backup-form') as HTMLFormElement)?.requestSubmit()}
		>
			Restore and restart
		</Button>
	{/snippet}
</Dialog>

<style>
	.page-heading {
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

	.settings-layout {
		display: grid;
		grid-template-columns: 13rem minmax(0, 1fr);
		gap: var(--space-6);
	}

	nav {
		display: grid;
		align-content: start;
		gap: var(--space-1);
	}

	nav a {
		display: flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
		border-inline-start: 3px solid transparent;
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.75rem;
		font-weight: 700;
		padding-inline: var(--space-3);
		text-decoration: none;
	}

	nav a.active {
		border-color: var(--crimson);
		background: color-mix(in srgb, var(--crimson) 8%, transparent);
		color: var(--crimson-dark);
	}

	.form-stack {
		display: grid;
		gap: var(--space-4);
	}

	.form-stack > label:not(.check) {
		display: grid;
		gap: var(--space-1);
	}

	.form-stack label > span {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
	}

	input[type='number'] {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-2);
	}

	.check {
		display: flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
	}

	.join-layout {
		display: grid;
		grid-template-columns: minmax(10rem, 16rem) 1fr;
		gap: var(--space-5);
	}

	.join-layout img {
		width: 100%;
		border: var(--border-strong);
		background: white;
		padding: var(--space-2);
	}

	.section-header {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-strong);
		padding-block-end: var(--space-3);
	}

	.section-header h2,
	.section-header p {
		margin: 0;
	}

	.record-list article {
		display: flex;
		min-height: 4.5rem;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
	}

	.record-list h3,
	.record-list p {
		margin: 0;
	}

	.record-list button {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-weight: 700;
	}

	.backup-toggle {
		margin-block: var(--space-3);
	}

	pre {
		overflow: auto;
		border: var(--border-subtle);
		background: var(--ink);
		color: var(--paper-light);
		padding: var(--space-4);
	}

	dl div {
		display: grid;
		grid-template-columns: 10rem 1fr;
		border-block-end: var(--border-subtle);
		padding: var(--space-2) 0;
	}

	dt {
		font-family: var(--font-display);
		font-weight: 700;
	}

	dd {
		margin: 0;
	}

	@media (max-width: 63.99rem) {
		.settings-layout {
			grid-template-columns: 1fr;
			gap: var(--space-4);
		}

		nav {
			display: flex;
			overflow-x: auto;
			padding-block-end: var(--space-2);
		}

		nav a {
			flex: 0 0 auto;
			border-inline-start: 0;
			border-block-end: 3px solid transparent;
		}

		.join-layout {
			grid-template-columns: 1fr;
		}

		.join-layout img {
			max-width: 18rem;
		}
	}

	@media (max-width: 47.99rem) {
		.section-header {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
