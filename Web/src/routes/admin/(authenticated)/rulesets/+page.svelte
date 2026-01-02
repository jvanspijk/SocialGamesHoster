<script lang="ts">
	import BackLink from '$lib/components/BackLink.svelte';
    import type { PageProps } from './$types';
    import LedgerTable from '$lib/components/LedgerTable.svelte';
    import SecondaryButton from '$lib/components/SecondaryButton.svelte';
    import type { GetAllRulesetsResponse } from '$lib/client/Rulesets/GetAllRulesets';
	import { goto } from '$app/navigation';


    let { data }: PageProps = $props();
</script>

<div class="main-container">
    <div class="page-header">
        <BackLink href="/admin" pageName="Admin Overview"></BackLink>
        <h1>Manage & create rulesets</h1>
    </div>

    <LedgerTable 
        data={data.rulesets}
    >
        {#snippet columns()}
            <tr>
                <th>Id</th>
                <th>Ruleset</th>
                <th>Actions</th>
            </tr>
        {/snippet}

        {#snippet rows(ruleset: GetAllRulesetsResponse)}
            <tr>
                <td data-label="Id" style="font-weight: bold;">{ruleset.id}</td>
                <td data-label="Ruleset">{ruleset.name}</td>
                <td>
                    <div class="actions-wrapper">
                        <SecondaryButton onclick={() => goto(`/admin/rulesets/${ruleset.id}`)}>Edit</SecondaryButton>
                        <SecondaryButton variant="danger" onclick={() => {}}>Delete</SecondaryButton>
                    </div>
                </td>
            </tr>
        {/snippet}
    </LedgerTable>

    <div class="actions-footer">
        <SecondaryButton variant="primary" onclick={() => {}}>Create new ruleset</SecondaryButton>
    </div>
</div>


<style>
    .main-container {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
    }

    .actions-footer {
        display: flex;
        justify-content: center;
    }

</style>
