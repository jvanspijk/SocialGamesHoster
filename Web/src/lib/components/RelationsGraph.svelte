<script lang="ts">
    let { 
        title, 
        items = [], 
        icon = '✒️',
        accentColor = '#5b4a3c' 
    } = $props<{
        title: string,
        items: string[],
        icon?: string,
        accentColor?: string
    }>();
</script>

<div class="ledger-entry" style="--accent: {accentColor}">
    <span class="icon">{icon}</span>
    <span class="title">{title}:</span>
    <div class="ink-container">
        <span class="count">{items.length}</span>
        {#if items.length > 0}
            <div class="scroll-popover">
                <div class="scroll-content">
                    <h4>{title}</h4>
                    <ul>
                        {#each items as item}
                            <li>{item}</li>
                        {/each}
                    </ul>
                </div>
            </div>
        {/if}
    </div>
</div>

<style>
    .ledger-entry {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        font-family: 'IM Fell English', serif;
        font-size: 1rem;
        color: #3e322b;
        position: relative;
    }

    .icon { font-size: 0.9em; opacity: 0.8; }
    .title { font-style: italic; font-weight: bold; }

    .ink-container {
        position: relative;
        cursor: pointer;
        display: inline-block;
    }

    .count {
        border-bottom: 1px double var(--accent);
        padding: 0 4px;
        font-weight: bold;
        transition: color 0.2s;
    }

    .ink-container:hover .count {
        color: var(--accent);
    }

    /* Popover styled like a small scrap of paper */
    .scroll-popover {
        visibility: hidden;
        position: absolute;
        bottom: 130%;
        left: 50%;
        transform: translateX(-50%);
        background-color: #fcf5e5; /* Lighter parchment */
        border: 2px solid #5b4a3c;
        box-shadow: 3px 3px 15px rgba(0,0,0,0.2);
        z-index: 100;
        min-width: 150px;
        padding: 10px;
        border-radius: 2px;
    }

    .scroll-popover::after {
        content: '';
        position: absolute;
        top: 100%;
        left: 50%;
        margin-left: -8px;
        border-width: 8px;
        border-style: solid;
        border-color: #5b4a3c transparent transparent transparent;
    }

    .ink-container:hover .scroll-popover {
        visibility: visible;
    }

    h4 {
        font-family: 'Cinzel', serif;
        font-size: 0.8rem;
        margin: 0 0 5px 0;
        border-bottom: 1px solid rgba(0,0,0,0.1);
        text-align: center;
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
        max-height: 200px;
        overflow-y: auto;
    }

    li {
        font-size: 0.9rem;
        padding: 2px 0;
        border-bottom: 1px dashed rgba(0,0,0,0.05);
    }
</style>