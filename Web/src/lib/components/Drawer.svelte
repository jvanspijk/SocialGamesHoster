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
		z-index: 40;
		display: flex;
		justify-content: flex-end;
	}

	.backdrop {
		position: absolute;
		inset: 0;
		background-color: rgba(15, 23, 42, 0.4);
	}

	.panel {
		position: relative;
		width: 100%;
		background-color: white;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.panel.md { max-width: 28rem; }
	.panel.sm { max-width: 24rem; }

	.header {
		padding: 1.5rem;
		border-bottom: 1px solid #f1f5f9;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.title {
		font-size: 1.25rem;
		font-weight: 700;
		color: #1e293b;
		margin: 0;
	}

	.close-btn {
		padding: 0.5rem;
		border-radius: 9999px;
		border: none;
		background: transparent;
		cursor: pointer;
		color: #64748b;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background-color 0.2s;
	}

	.close-btn:hover {
		background-color: #f1f5f9;
	}

	.content {
		flex: 1;
		overflow-y: auto;
		padding: 1.5rem;
	}

	.footer {
		padding: 1.5rem;
		border-top: 1px solid #f1f5f9;
		background-color: #f8fafc;
		display: flex;
		gap: 0.75rem;
	}
</style>