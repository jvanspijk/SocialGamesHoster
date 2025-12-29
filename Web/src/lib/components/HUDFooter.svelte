<script>
    let isCollapsed = true; 

    function toggleFooter() {
        isCollapsed = !isCollapsed;
    }
</script>

<footer class={isCollapsed ? 'collapsed' : ''}>
    <div class="toggle-control">
        <button on:click={toggleFooter} title={isCollapsed ? 'Expand Menu' : 'Collapse Menu'}>
            {isCollapsed ? '▲' : '▼'}
        </button>
    </div>
    <div class="hud">
        <slot></slot>
    </div>
</footer>

<style>
    footer {
        position: fixed; 
        bottom: 0;
        left: 0;
        width: 100%;
        z-index: 100;
        background: #3e322b;
        
        border-top: 4px solid #5b4a3c; 
        box-shadow: 0 -4px 15px rgba(0, 0, 0, 0.7); 
        padding: 0;
        
        font-family: 'Cinzel', serif;
        color: #f7e7c4;
        text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.9);
        letter-spacing: 1px;
        transition: transform 0.3s ease-in-out;
        box-sizing: border-box;
    }

    footer.collapsed {
        transform: translateY(calc(100% - 38px));
    }

    .hud {
        display: flex;
        justify-content: flex-end; 
        align-items: center;
        padding: 5px 20px 10px 20px;
        gap: 25px;

        flex-wrap: wrap; 
        max-width: 100vw; 
        box-sizing: border-box;
        overflow: hidden;
    }

    .hud > :global(*) {
        max-width: fit-content;
        flex-shrink: 0; /* Prevents them from squishing since you have "plenty of space" */
    }

    .toggle-control {
        display: flex;
        justify-content: center;
        width: 100%;
        background-color: #3e322b;
        border-bottom: 1px solid #5b4a3c;
        height: 38px;
    }

    .toggle-control button {
        font-family: 'Cinzel', serif;
        font-size: 1.4em;
        color: #e7c58d;
        background: none;
        border: none;
        cursor: pointer;
        padding: 0 10px;
        line-height: 1;
        transition: color 0.2s;
        height: 100%;
        width: 100%;
        border-radius: 5px;
    }

    .toggle-control button:hover {
        color: #f7e7c4;
    }
    
    :global(footer a), :global(footer button) {
        font-family: 'Cinzel', serif;
        text-transform: uppercase;
        color: #f7e7c4;
        background: none;
        border: none;
        padding: 5px 10px;
        cursor: pointer;
        transition: color 0.2s;
        font-size: 0.9em;
    }

    :global(footer a:hover), :global(footer button:hover) {
        color: #e7c58d;
    }
</style>