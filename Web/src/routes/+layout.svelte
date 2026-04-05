<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<div class="page-main">
	<div class="parchment-container">
		{@render children?.()}
	</div>
</div>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<style>
	/* 2. Page & Background Styles */
	:global(body) {
		background-color: var(--color-bg-canvas);
		margin: 0;
		padding: 0;
		min-height: 100vh;
	}

	.page-main {
		background-color: var(--color-bg-canvas);
		display: flex;
		justify-content: center;
		padding: 10px 14px;
		margin-top: 10px;
		text-align: center;
	}

	.parchment-container {
		background-color: var(--color-surface);
		color: var(--color-text);
		font-family: var(--font-body);
		width: 100%;
		max-width: 700px;
		box-shadow:
			0 4px 10px rgba(0, 0, 0, 0.4),
			inset 0 0 50px rgba(100, 80, 50, 0.3),
			inset 0 0 10px rgba(100, 80, 50, 0.5);
		padding: 20px 40px;
		border: 8px solid var(--color-border);
		border-radius: 4px;
		line-height: 1.6;
		min-height: 85vh;
	}

	/* for mobile */
	@media (max-width: 650px) {
		.parchment-container {
			padding: 15px 10px !important;
			border-width: 5px;
		}
	}

	/* 3. Common Typography*/

	:global(h1),
	:global(h2) {
		font-family: var(--font-heading);
		color: var(--color-border);
		text-transform: uppercase;
		letter-spacing: 2px;
		margin-bottom: 10px;
		text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.5);
	}

	:global(h1) {
		font-size: 2.5em;
		text-align: center;
		border-bottom: 3px double var(--color-border);
		padding-bottom: 15px;
		margin-bottom: 20px;
	}

	:global(h2) {
		font-size: 1.75em;
		border-bottom: 1px solid var(--color-border);
		padding-bottom: 5px;
		margin-top: 25px;
	}

	:global(h3) {
		font-size: 1.25em;
		text-align: center;
		color: var(--color-text-muted);
		font-style: italic;
		margin-top: -8px;
		padding-bottom: 12px;
	}

	:global(p) {
		margin-top: -2px;
		margin-bottom: 4px;
		font-size: 1.1em;
	}

	:global(.container) {
		display: flex;
		flex-direction: column;
		align-items: center;
		max-width: 80%;
		margin: 0 auto;
	}

	:global(.error-message) {
		font-family: var(--font-heading);
		color: var(--color-accent-strong);
		text-align: center;
		font-size: 1.5em;
		padding: 50px;
	}

	:global(.success-message) {
		font-family: var(--font-heading);
		color: var(--color-success);
		border-color: var(--color-success);
		background-color: var(--color-success-soft);
		text-align: center;
		font-size: 1.5em;
		padding: 50px;
	}
</style>
