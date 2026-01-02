<script lang="ts">
	import BackLink from '$lib/components/BackLink.svelte';
    import type { PageProps } from './$types';

    interface Ability {
        name: string;
        description: string;
    }

    interface Role {
        name: string;
        description: string;
        abilities: Ability[];
    }

    interface Ruleset {
        name: string;
        description: string;
        roles: Role[];
        abilities: Ability[];
    }

    let { data }: PageProps = $props();
    let ruleset = $state<Ruleset>({
        name: '',
        description: '',
        roles: [],
        abilities: []
    });

    type Tab = 'config' | 'role' | 'ability';
    let activeTab = $state<Tab>('config');
    let currentRole = $state<Role | null>(null);

    function selectRole(role: Role) {
        currentRole = role;
        activeTab = 'role';
    }
</script>

<div class="create-container">
        <div class="page-header">
            <BackLink href="/admin" pageName="Admin Overview"></BackLink>
            <h1>Manage & create rulesets</h1>
        </div>

    <section class="basic-info">
        <input type="text" bind:value={ruleset.name} placeholder="Ruleset Name..." class="title-input" />
        <textarea bind:value={ruleset.description} placeholder="Describe the theme or laws..."></textarea>
    </section>

    <hr />

    <div class="editor-grid">
        <nav class="side-nav">
            <button class="active">Config</button>
            <button>Global Abilities ({ruleset.abilities.length})</button>
            <div class="nav-divider">Roles</div>
            {#each ruleset.roles as role}
                <button>{role.name || 'Unnamed Role'}</button>
            {/each}
            <button class="add-btn">+ Add Role</button>
        </nav>

        <main class="editor-content">
            <h2>Edit Role: {currentRole?.name || 'No Role Selected'}</h2>
        </main>
    </div>
</div>

<style>
    .editor-grid {
        display: grid;
        grid-template-columns: 250px 1fr;
        gap: 2rem;
        min-height: 60vh;
    }
    .side-nav {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        border-right: 1px solid #eee;
    }
    .title-input {
        font-size: 2rem;
        font-weight: bold;
        border: none;
        outline: none;
        width: 100%;
    }
</style>