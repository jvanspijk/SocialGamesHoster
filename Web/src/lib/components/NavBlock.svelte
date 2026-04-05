<script lang="ts">
	let { href, title, description, icon = '📜' } = $props();
</script>

<a {href} class="nav-block">
	<div class="icon">{icon}</div>
	<div class="text-content">
		<h3>{title}</h3>
		<p>{description}</p>
	</div>
</a>

<style>
	.nav-block {
		display: flex;
		align-items: center;
		text-decoration: none;
		background-color: var(--color-surface-soft);
		border: 2px solid var(--color-border);
		padding: 1.5rem;
		margin: 1rem 0;
		transition: all 0.2s ease-in-out;
		box-shadow: 3px 3px 0px var(--color-border);
		cursor: pointer;
		text-align: left;
	}

	.nav-block:hover {
		transform: translate(-2px, -2px);
		box-shadow: 6px 6px 0px var(--color-accent-strong);
		background-color: var(--color-on-accent) cf5;
		border-color: var(--color-accent-strong);
	}

	.icon {
		font-size: 2.5rem;
		margin-right: 1.5rem;
		filter: sepia(1);
	}

	h3 {
		margin: 0;
		color: var(--color-accent-strong) !important;
		font-family: var(--font-heading);
		font-style: normal !important;
		text-transform: uppercase;
		font-size: 1.2rem;
	}

	p {
		margin: 4px 0 0 0;
		color: var(--color-border);
		font-size: 0.95rem;
		line-height: 1.2;
	}
</style>
