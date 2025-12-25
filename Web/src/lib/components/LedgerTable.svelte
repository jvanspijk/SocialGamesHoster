<script lang="ts">
    import SecondaryButton from '$lib/components/SecondaryButton.svelte';
    
    let { 
        data = [], 
        columns,
        rows,
        searchPlaceholder = "Search records...",
        enableFilter = false,
        filterOptions = []
    } = $props();

    let searchTerm = $state("");
    let filterValue = $state("All");
    let currentPage = $state(1);
    const itemsPerPage = 5;

    const filteredData = $derived(
        data.filter(item => {
            const matchesSearch = JSON.stringify(item).toLowerCase().includes(searchTerm.toLowerCase());
            const matchesFilter = filterValue === "All" || item.status === filterValue;
            return matchesSearch && matchesFilter;
        })
    );

    const paginatedData = $derived(
        filteredData.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage)
    );

    const totalPages = $derived(Math.ceil(filteredData.length / itemsPerPage));
</script>

<div class="ledger-container">
    <div class="ledger-tools">
        <div class="search-wrapper">
            <input 
                type="text" 
                bind:value={searchTerm} 
                placeholder={searchPlaceholder} 
                class="ledger-input"
            />
        </div>

        {#if enableFilter}
            <select bind:value={filterValue} class="ledger-select">
                <option value="All">All Statuses</option>
                {#each filterOptions as opt}
                    <option value={opt}>{opt}</option>
                {/each}
            </select>
        {/if}
    </div>

    <table class="ledger">
        <thead>
            {@render columns()}
        </thead>
        <tbody>
            {#each paginatedData as item}
                {@render rows(item)}
            {:else}
                <tr>
                    <td colspan="100" class="no-data">No scrolls found matching your criteria.</td>
                </tr>
            {/each}
        </tbody>
    </table>

    {#if totalPages > 1}
        <div class="pagination">
            <SecondaryButton onclick={() => currentPage--} enabled={currentPage != 1}>Previous</SecondaryButton>
            <span class="page-info">Page {currentPage} of {totalPages}</span>
            <SecondaryButton onclick={() => currentPage++} enabled={currentPage != totalPages}>Next</SecondaryButton>
        </div>
    {/if}
</div>

<style>
    .ledger-container {
        width: 100%;
        margin-top: 1rem;
    }

    .ledger-tools {
        display: flex;
        gap: 1rem;
        margin-bottom: 1.5rem;
        justify-content: space-between;
    }

    .search-wrapper {
        position: relative;
        flex-grow: 1;
    }

    .ledger-input, .ledger-select {
        background: #fcf5e5;
        border: 2px solid #5b4a3c;
        font-family: 'IM Fell English', serif;
        padding: 8px 8px 8px 35px;
        color: #3e322b;
        width: 100%;
    }

    .ledger-select {
        padding-left: 10px;
        width: auto;
        cursor: pointer;
    }

    .ledger {
        width: 100%;
        border-collapse: collapse;
    }

    .no-data {
        text-align: center;
        padding: 3rem;
        font-style: italic;
        opacity: 0.6;
    }

    .pagination {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 2rem;
        margin-top: 1.5rem;
        font-family: 'Cinzel', serif;
    }

    .page-info {
        font-size: 0.9rem;
        color: #5b4a3c;
    }
</style>