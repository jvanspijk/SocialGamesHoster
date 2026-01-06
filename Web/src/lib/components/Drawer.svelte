<script lang="ts">
	import { fade, fly } from 'svelte/transition';

	let { 
		open, 
		title, 
		size = 'md', 
		onclose, 
		children, 
		footer 
	} = $props();
</script>

{#if open}
	<div class="drawer-container">
		<!-- Backdrop -->
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div 
			transition:fade={{ duration: 200 }} 
			class="backdrop" 
			onclick={onclose}
		></div>
		
		<!-- Panel -->
		<div 
			transition:fly={{ x: '100%', duration: 300 }} 
			class="panel {size}"
		>
			<header class="header">
				<h3 class="title">{title}</h3>
				<button onclick={onclose} class="close-btn" aria-label="Close">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</header>
			
			<div class="content">
				{@render children()}
			</div>

			{#if footer}
				<footer class="footer">
					{@render footer()}
				</footer>
			{/if}
		</div>
	</div>
{/if}

<style>
    .drawer-container {
        position: fixed;
        inset: 0;
        z-index: 100;
        display: flex;
        justify-content: flex-end;
    }

    .backdrop {
        position: absolute;
        inset: 0;
        background-color: rgba(45, 35, 25, 0.6);
        backdrop-filter: blur(2px);
    }

    .panel {
        position: relative;
        width: 100%;
        height: 100%;
        display: flex;
        flex-direction: column;
        
        background-color: #5b4a3c;
        border-left: 4px solid #5b4a3c;
        box-shadow: -10px 0 30px rgba(0, 0, 0, 0.3);
    }

    .panel.md { max-width: 32rem; }
    .panel.sm { max-width: 24rem; }

    .header {
        padding: 1.5rem;
        border-bottom: 2px solid #e8e0c5;
        display: flex;
        justify-content: space-between;
        align-items: center;
        background-color: #f3edd7;
    }

    .title {
        font-family: 'IM Fell English', serif;
        font-size: 1.75rem;
        font-weight: 700;
        color: #3e322b;
        margin: 0;
    }

    .close-btn {
        padding: 0.5rem;
        border: 2px solid transparent;
        background: transparent;
        cursor: pointer;
        color: #5b4a3c;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all 0.2s;
        border-radius: 4px;
    }

    .close-btn:hover {
        color: #8c2a3e;
        background-color: #e8e0c5;
        border-color: #5b4a3c;
    }

    .content {
        flex: 1;
        overflow-y: auto;
        padding: 2rem;
        background-image: radial-gradient(#5b4a3c11 1px, transparent 1px);
        background-size: 20px 20px;
    }

    .footer {
        padding: 1.5rem;
        border-top: 2px solid #e8e0c5;
        background-color: #f3edd7;
        display: flex;
        gap: 1rem;
        justify-content: flex-end;
    }
</style>