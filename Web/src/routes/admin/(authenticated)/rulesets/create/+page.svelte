<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	
	import Drawer from '$lib/components/Drawer.svelte';
	import ActionCard from '$lib/components/ActionCard.svelte';
	import SecondaryButton from '$lib/components/SecondaryButton.svelte';
	import SelectionPool from '$lib/components/SelectionPool.svelte';
	import TextInput from '$lib/components/TextInput.svelte';
	import BackLink from '$lib/components/BackLink.svelte';

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

<div class="page-header">
	<BackLink href="/admin" pageName="Admin Overview"></BackLink>
	<h1>Ruleset creator</h1>
	<SecondaryButton onclick={() => console.log($state.snapshot(ruleset))}>
	Save Ruleset
	</SecondaryButton>
</div>

<main class="container">
	<section class="section header-section">
		<div class="input-group">
			<label for="rs-name">Ruleset Name</label>
			<TextInput 
				bind:value={ruleset.name} 
				placeholder="" 
				name="rs-name"
			/>
		</div>
		<div class="input-group">
			<label for="rs-desc">Description</label>
			<TextInput 
				bind:value={ruleset.description} 
				placeholder="Describe the general goals and ideas behind this ruleset..." 
				multiline={true} 
				rows={4} 
			/>
		</div>
	</section>

	<section class="section">
		<div class="section-header">
			<h2 class="title">Roles</h2>
			<SecondaryButton variant="secondary" onclick={() => openRoleDrawer()}>
				+ Add Role
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
			<TextInput 
				bind:value={roleForm.name} 
				placeholder="Role Name" 
				name="roleName"
			/>
			<TextInput 
				bind:value={roleForm.description} 
				placeholder="Description" 
				multiline={true} 
				rows={4} 
			/>
		</div>
		
		<div class="form-divider">
			<div class="divider-header">
				<span class="label">Abilities</span>
				<SecondaryButton onclick={() => abilityDrawerOpen = true}>
					<span class="text-xs font-bold">+ New Ability</span>
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
		<TextInput bind:value={abilityForm.name} placeholder="Ability Name"/>
		<TextInput bind:value={abilityForm.description} placeholder="Description" multiline={true} rows={4}/>
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
        background-color: #2d2319; 
        background-image: url("data:image/svg+xml,%3Csvg width='100' height='100' viewBox='0 0 100 100' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M11 18c3.866 0 7-3.134 7-7s-3.134-7-7-7-7 3.134-7 7 3.134 7 7 7zm48 25c3.866 0 7-3.134 7-7s-3.134-7-7-7-7 3.134-7 7 3.134 7 7 7zm-43-7c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zm63 31c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zM34 90c1.657 0 3-1.343 3-3s-1.343-3-3-3-3 1.343-3 3 1.343 3 3 3zm56-76c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zM12 86c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zm66 3c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zm-46-45c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zm54 0c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zM57 7c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zm-8 48c1.105 0 2-.895 2-2s-.895-2-2-2-2 .895-2 2 .895 2 2 2zM25 34c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm23 47c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm50 35c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM21 0c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm54 44c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM8 33c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm56 56c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM33 51c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm65 11c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM42 2c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-7 48c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm59 2c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM40 76c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm57 0c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM60 28c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm39 67c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM46 39c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM28 89c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM71 81c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm18-68c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM42 29c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm8 40c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-6 25c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm22-97c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM1 31c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm7 51c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm33 4c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm60-46c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-9 52c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm20-17c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-44-20c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-69 60c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm13 14c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm89-7c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM73 66c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM28 14c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm1 90c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm77-76c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-79 12c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm54 39c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm32-1c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM32 6c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm51 38c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-75 37c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm55 35c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-61-14c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM15 6c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm32 7c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm36 41c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM60 71c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM35 88c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM12 40c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm0 54c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM3 75c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm39-52c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm60-2c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-9 30c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm20 35c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-44-14c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-69 76c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm13-37c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm89-7c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zM73 67c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1zm-45-48c.552 0 1-.448 1-1s-.448-1-1-1-1 .448-1 1 .448 1 1 1z' fill='%233e322b' fill-opacity='0.2' fill-rule='evenodd'/%3E%3C/svg%3E");
        margin: 0;
        font-family: 'IM Fell English', serif;
        color: #3e322b;
    }

    .container {
        max-width: 60rem;
        margin: 2rem auto;
        padding: 0 1.5rem;
    }

    .header-section {
        background-color: #fff9e6;
        padding: 2.5rem;
        border-radius: 4px;
        border: 2px solid #5b4a3c;
        box-shadow: 8px 8px 0px rgba(0,0,0,0.2);
        display: flex;
        flex-direction: column;
        gap: 1.5rem;
        position: relative;
		margin: 0 auto;
		width: 100%;
		margin-bottom: 2rem;
    }

    .header-section::before {
        position: absolute;
        top: 10px;
        left: 10px;
        color: #5b4a3c;
        opacity: 0.3;
        font-size: 2rem;
    }

    .title {
        font-family: 'IM Fell English', serif;
        font-size: 2rem;
        font-weight: 700;
        color: #3e322b;
        border-bottom: 2px solid #5b4a3c;
        margin-bottom: 1rem;
		width: 100%;
    }

    .input-group label {
        font-family: 'IM Fell English', serif;
        font-size: 1.1rem;
        font-weight: bold;
        color: #5b4a3c;
        text-transform: uppercase;
        letter-spacing: 0.05em;
    }   

	.section {
		width: 100%;
		gap: 1.5rem;
	}

    .grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 2rem;
        margin-top: 1rem;
		width: 100%;
		margin: 0 auto;
		margin-top: 1rem;
    }

    .empty-state {
        grid-column: 1 / -1;
        
        display: flex;
        align-items: center;
        justify-content: center;
		justify-self: center;
		margin: 0 auto;
        
        width: 100%;
        max-width: 600px;
        
        background-color: rgba(255, 249, 230, 0.5);
        border: 2px dashed #5b4a3c;
        padding: 3rem;
        
        color: #5b4a3c; 
        font-style: italic;
        text-align: center;
    }

    .toast {
        position: fixed;
        bottom: 2rem;
        right: 2rem;
        z-index: 1000;
        background-color: #fff9e6;
        color: #3e322b;
        padding: 1rem 2rem;
        border: 2px solid #8c2a3e;
        border-radius: 4px;
        box-shadow: 0 10px 25px rgba(0,0,0,0.4);
        display: flex;
        align-items: center;
        gap: 1rem;
        font-weight: bold;
    }

    .form-divider {
        padding-top: 1.5rem;
        border-top: 2px solid #e8e0c5;
    }

    .label {
        font-family: 'IM Fell English', serif;
        font-size: 1rem;
        color: #5b4a3c;
        font-weight: bold;
    }
</style>