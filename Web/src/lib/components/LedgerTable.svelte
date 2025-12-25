<script lang="ts">
    import SecondaryButton from '$lib/components/SecondaryButton.svelte';
    
    let { 
        data = [], 
        columns,
        rows,
        searchPlaceholder = "Search...",
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

    <div class="table-scroll-shield">
        <table class="ledger">
            <thead>
                {@render columns()}
            </thead>
            <tbody>
                {#each paginatedData as item}
                    {@render rows(item)}
                {:else}
                    <tr>
                        <td colspan="100" class="no-data">Nothing found matching your criteria.</td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>

    {#if totalPages > 1}
        <div class="pagination">
            <SecondaryButton onclick={() => currentPage--} enabled={currentPage != 1}>Previous</SecondaryButton>
            <span class="page-info">Page {currentPage} of {totalPages}</span>
            <SecondaryButton onclick={() => currentPage++} enabled={currentPage != totalPages}>Next</SecondaryButton>
        </div>
    {/if}
</div>

<style>    
    .search-wrapper {
        position: relative;
        flex-grow: 1;
    }

    .pagination {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        align-items: center;
        gap: 1rem;
        margin-top: 1.5rem;
        font-family: 'Cinzel', serif;
    }

    .page-info {
        font-size: 0.9rem;
        color: #5b4a3c;
    }

    .ledger-container {
        width: 100%;
        margin-top: 1rem;
        border-radius: 2px;
        overflow: hidden;
        border: 1px solid rgba(0, 0, 0, 0);          
    }

    .ledger-tools {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        margin-bottom: 1.5rem;
    }
    
    .ledger-input, .ledger-select {
        background: #fcf5e5;
        border: 2px solid #5b4a3c;
        font-family: 'IM Fell English', serif;
        padding: 8px 8px 8px 35px;
        color: #3e322b;
        width: 100%;
        box-sizing: border-box;
        font-size: 16px;
    }

    .ledger-select {
        padding-left: 10px;
        width: auto;
    }

    .ledger {
        width: 100%;
        border-collapse: collapse;
    }

    .ledger :global(th), .ledger :global(td) {
        padding: 12px 10px;
        text-align: left;
        border-bottom: 1px solid rgba(91, 74, 60, 0.2);
    }

    .ledger :global(th) {
        font-family: 'Cinzel', serif;
        border-bottom: 2px solid #5b4a3c;
        color: #5b4a3c;
        text-transform: uppercase;
        letter-spacing: 1px;
        white-space: nowrap;
        font-size: 0.85rem;
    }

    .ledger :global(tr) {
        background: rgba(91, 74, 60, 0.16);
    }   

    .ledger :global(thead) {
        background-color: rgba(91, 74, 60, 0.08);
    }

    /* desktop */
    @media (min-width: 650px) {
        .ledger-tools {
            flex-direction: row;
        }
        .ledger :global(td:not(:last-child)), 
        .ledger :global(th:not(:last-child)) {
            border-right: 1px solid rgba(91, 74, 60, 0.1);
        }
        .ledger :global(tbody tr:last-child td) {
            border-bottom: none; 
        }
        .ledger {
            border: 2px double #3e322b;
            border-radius: 2px;
        }
    }

    /* mobile */
    @media (max-width: 650px) {
        .ledger :global(thead) { display: none; }

        .ledger :global(tr) {
            display: block;
            
            margin-bottom: 2rem;
            border: 2px solid #5b4a3c;
            padding: 15px;
            box-shadow: 2px 2px 0px rgba(91, 74, 60, 0.1);
        }

        .ledger :global(td) {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px 0;
            border-bottom: 1px double rgba(91, 74, 60, 0.1);
        }

        .ledger :global(td:last-child) {
            border-bottom: none;
            padding-top: 1.5rem;
            flex-direction: column; 
            gap: 2px;
        }

        .ledger :global(td::before) {
            content: attr(data-label);
            font-family: 'Cinzel', serif;
            font-weight: bold;
            font-size: 0.75rem;
            color: #5b4a3c;
            text-transform: uppercase;
        }
    }
</style>