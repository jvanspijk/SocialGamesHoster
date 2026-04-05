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
		background: var(--color-text);

		border-top: 3px solid var(--color-border);
		box-shadow: 0 -4px 15px rgba(0, 0, 0, 0.7);
		padding: 0;

		font-family: var(--font-heading);
		color: var(--color-surface);
		text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.9);
		letter-spacing: 1px;
		transition: transform 0.2s ease-out;
		box-sizing: border-box;
	}

	footer.collapsed {
		transform: translateY(calc(100% - 40px));
	}

	.hud {
		display: flex;
		justify-content: flex-start;
		align-items: center;
		padding: 5px 20px 10px 20px;
		gap: 25px;

		flex-wrap: nowrap;
		max-width: 100vw;
		box-sizing: border-box;
		overflow: hidden;
	}

	.hud > :global(*) {
		flex-shrink: 0;
	}

	.toggle-control {
		display: flex;
		justify-content: center;
		width: 100%;
		background-color: var(--color-text);
		border-bottom: 1px solid var(--color-border);
		height: 40px;
	}

	.toggle-control button {
		font-family: var(--font-heading);
		font-size: 1.4em;
		color: var(--color-highlight);
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
		color: var(--color-surface);
	}

	:global(footer a),
	:global(footer button) {
		font-family: var(--font-heading);
		text-transform: uppercase;
		color: var(--color-surface);
		background: none;
		border: none;
		padding: 5px 10px;
		cursor: pointer;
		transition: color 0.2s;
		font-size: 0.9em;
	}

	:global(footer a:hover),
	:global(footer button:hover) {
		color: var(--color-highlight);
	}
</style>
