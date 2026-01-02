<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	
	import Drawer from '$lib/components/Drawer.svelte';
	import ActionCard from '$lib/components/ActionCard.svelte';
	import SecondaryButton from '$lib/components/SecondaryButton.svelte';
	import SelectionPool from '$lib/components/SelectionPool.svelte';

	interface Ability {
		id: string;
		name: string;
		description: string;
	}

	interface Role {
		id: string;
		name: string;
		description: string;
		abilityIds: string[];
	}

	interface Ruleset {
		name: string;
		description: string;
		roles: Role[];
		abilities: Ability[];
	}

	let ruleset = $state<Ruleset>({
		name: '',
		description: '',
		roles: [],
		abilities: []
	});

	let roleDrawerOpen = $state(false);
	let abilityDrawerOpen = $state(false);
	let editingRoleId = $state<string | null>(null);
	let message = $state({ text: '', visible: false });

	let roleForm = $state<Omit<Role, 'id'>>({ name: '', description: '', abilityIds: [] });
	let abilityForm = $state<Omit<Ability, 'id'>>({ name: '', description: '' });

	const rolesDisplay = $derived(
		ruleset.roles.map(role => ({
			...role,
			tags: role.abilityIds
				.map(id => ruleset.abilities.find(a => a.id === id)?.name)
				.filter((name): name is string => !!name)
		}))
	);

	function showMessage(text: string) {
		message.text = text;
		message.visible = true;
		setTimeout(() => message.visible = false, 3000);
	}

	function openRoleDrawer(id: string | null = null) {
		if (id) {
			const role = ruleset.roles.find(r => r.id === id);
			if (role) {
				editingRoleId = id;
				roleForm = { 
					name: role.name, 
					description: role.description, 
					abilityIds: [...role.abilityIds] 
				};
			}
		} else {
			editingRoleId = null;
			roleForm = { name: '', description: '', abilityIds: [] };
		}
		roleDrawerOpen = true;
	}

	function saveRole() {
		if (!roleForm.name.trim()) return showMessage("Name required");
		
		if (editingRoleId) {
			const idx = ruleset.roles.findIndex(r => r.id === editingRoleId);
			ruleset.roles[idx] = { ...roleForm, id: editingRoleId };
		} else {
			ruleset.roles.push({ ...roleForm, id: 'role-' + Date.now() });
		}
		roleDrawerOpen = false;
		showMessage("Role saved");
	}

	function saveAbility() {
		if (!abilityForm.name.trim()) return showMessage("Name required");
		const newAb = { ...abilityForm, id: 'ab-' + Date.now() };
		ruleset.abilities.push(newAb);
		
		if (roleDrawerOpen) {
			roleForm.abilityIds = [...roleForm.abilityIds, newAb.id];
		}
		
		abilityDrawerOpen = false;
		abilityForm = { name: '', description: '' };
		showMessage("Ability created");
	}

	function toggleAbility(id: string) {
		const list = roleForm.abilityIds;
		roleForm.abilityIds = list.includes(id) 
			? list.filter(i => i !== id) 
			: [...list, id];
	}

	function deleteRole(id: string) {
		ruleset.roles = ruleset.roles.filter(r => r.id !== id);
	}
</script>

<nav class="nav-bar">
	<div class="nav-brand">
		<div class="brand-icon">
			<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
			</svg>
		</div>
		<h1 class="brand-text">Ruleset Studio</h1>
	</div>
	<SecondaryButton onclick={() => console.log($state.snapshot(ruleset))}>
		Save Ruleset
	</SecondaryButton>
</nav>

<main class="container">
	<section class="section header-section">
		<div class="input-group">
			<label for="rs-name">Ruleset Name</label>
			<input id="rs-name" bind:value={ruleset.name} type="text" placeholder="e.g. Tournament 2024">
		</div>
		<div class="input-group">
			<label for="rs-desc">Description</label>
			<textarea id="rs-desc" bind:value={ruleset.description} rows="2" placeholder="Describe the purpose of this ruleset..."></textarea>
		</div>
	</section>

	<section class="section">
		<div class="section-header">
			<h2 class="title">Roles</h2>
			<SecondaryButton variant="secondary" onclick={() => openRoleDrawer()}>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
				Add Role
			</SecondaryButton>
		</div>

		<div class="grid">
			{#each rolesDisplay as role (role.id)}
				<ActionCard 
					title={role.name}
					description={role.description}
					tags={role.tags}
				>
					{#snippet actions()}
						<SecondaryButton onclick={() => openRoleDrawer(role.id)}>
							<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
						</SecondaryButton>
						<SecondaryButton variant="danger" onclick={() => deleteRole(role.id)}>
							<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
						</SecondaryButton>
					{/snippet}
				</ActionCard>
			{:else}
				<div class="empty-state">
					No roles created yet.
				</div>
			{/each}
		</div>
	</section>
</main>

<!-- DRAWER: ROLE -->
<Drawer 
	open={roleDrawerOpen} 
	title={editingRoleId ? 'Edit Role' : 'Add Role'} 
	onclose={() => roleDrawerOpen = false}
>
	<div class="form-container">
		<div class="form-section">
			<input bind:value={roleForm.name} placeholder="Role Name" class="field">
			<textarea bind:value={roleForm.description} placeholder="Description" class="field" rows="3"></textarea>
		</div>
		
		<div class="form-divider">
			<div class="divider-header">
				<span class="label">Abilities</span>
				<SecondaryButton onclick={() => abilityDrawerOpen = true}>
					<span class="text-xs font-bold">+ NEW</span>
				</SecondaryButton>
			</div>
			
			<SelectionPool 
				items={ruleset.abilities} 
				selectedIds={roleForm.abilityIds} 
				onToggle={toggleAbility} 
			/>
		</div>
	</div>
	
	{#snippet footer()}
		<SecondaryButton onclick={saveRole}>Save Role</SecondaryButton>
		<SecondaryButton variant="secondary" onclick={() => roleDrawerOpen = false}>Cancel</SecondaryButton>
	{/snippet}
</Drawer>

<Drawer 
	open={abilityDrawerOpen} 
	title="New Ability" 
	size="sm" 
	onclose={() => abilityDrawerOpen = false}
>
	<div class="form-section">
		<input bind:value={abilityForm.name} placeholder="Ability Name" class="field">
		<textarea bind:value={abilityForm.description} placeholder="Description" class="field" rows="4"></textarea>
	</div>

	{#snippet footer()}
		<SecondaryButton onclick={saveAbility}>Create Ability</SecondaryButton>
	{/snippet}
</Drawer>

{#if message.visible}
	<div transition:fly={{ y: 20 }} class="toast">
		<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
			<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
			<polyline points="22 4 12 14.01 9 11.01"></polyline>
		</svg>
		<span>{message.text}</span>
	</div>
{/if}

<style>
	:global(body) {
		background-color: #f8fafc;
		margin: 0;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.nav-bar {
		background-color: white;
		border-bottom: 1px solid #e2e8f0;
		padding: 1rem 2rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		position: sticky;
		top: 0;
		z-index: 10;
	}

	.nav-brand {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.brand-icon {
		background-color: #4f46e5;
		padding: 0.5rem;
		border-radius: 0.5rem;
		display: flex;
	}

	.brand-text {
		font-size: 1.125rem;
		font-weight: 700;
		letter-spacing: -0.025em;
		margin: 0;
	}

	.container {
		max-width: 56rem;
		margin: 0 auto;
		padding: 3rem 1.5rem;
	}

	.section {
		margin-bottom: 3rem;
	}

	.header-section {
		background-color: white;
		padding: 2rem;
		border-radius: 0.75rem;
		border: 1px solid #e2e8f0;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.title {
		font-size: 1.5rem;
		font-weight: 700;
		margin: 0;
	}

	.input-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.input-group label {
		font-size: 0.875rem;
		font-weight: 600;
		color: #334155;
	}

	.input-group input, .input-group textarea, .field {
		width: 100%;
		padding: 0.625rem 1rem;
		border-radius: 0.5rem;
		border: 1px solid #e2e8f0;
		font-size: 1rem;
		transition: ring 0.2s;
		box-sizing: border-box;
	}

	.input-group input:focus, .input-group textarea:focus, .field:focus {
		outline: none;
		border-color: #4f46e5;
		box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.1);
	}

	.grid {
		display: grid;
		gap: 1rem;
	}

	.empty-state {
		background-color: #f1f5f9;
		border: 2px dashed #cbd5e1;
		border-radius: 0.75rem;
		padding: 3rem;
		text-align: center;
		color: #94a3b8;
	}

	.form-container {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.form-section {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-divider {
		padding-top: 1.5rem;
		border-top: 1px solid #f1f5f9;
	}

	.divider-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.label {
		font-size: 0.75rem;
		font-weight: 700;
		color: #64748b;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.toast {
		position: fixed;
		bottom: 1.5rem;
		right: 1.5rem;
		z-index: 100;
		background-color: #0f172a;
		color: white;
		padding: 0.75rem 1.5rem;
		border-radius: 0.75rem;
		box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 0.875rem;
		font-weight: 500;
	}
</style>