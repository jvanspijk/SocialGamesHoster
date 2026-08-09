<script lang="ts">
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';

	export type RoleHeroPresentation = {
		name: string;
		description: string;
		teamName?: string;
	};

	let { role, asset }: { role: RoleHeroPresentation; asset?: string } = $props();
</script>

<header class:has-art={asset}>
	{#if asset}
		<div class="role-art"><ProtectedMedia src={asset} kind="image" alt="" /></div>
	{:else}
		<div class="role-fallback" aria-hidden="true">
			<span>{role.name.slice(0, 1).toUpperCase()}</span>
		</div>
	{/if}
	<div class="hero-gradient"></div>
	<div class="hero-copy">
		{#if role.teamName}<p>{role.teamName}</p>{/if}
		<h1>{role.name}</h1>
		<p class="description">{role.description}</p>
	</div>
</header>

<style>
	header {
		position: relative;
		display: grid;
		min-height: clamp(20rem, 54vh, 34rem);
		overflow: hidden;
		background: linear-gradient(145deg, var(--navy), var(--ink));
		color: var(--paper-light);
		isolation: isolate;
	}

	.role-art,
	.role-fallback,
	.hero-gradient,
	.hero-copy {
		grid-area: 1 / 1;
	}

	.role-art {
		position: absolute;
		inset: 0;
	}

	.role-art :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.role-fallback {
		display: grid;
		place-items: center;
		background:
			radial-gradient(circle at 50% 40%, rgb(223 189 101 / 20%), transparent 26%),
			linear-gradient(135deg, var(--navy), var(--crimson-dark) 62%, var(--ink));
	}

	.role-fallback::before,
	.role-fallback::after {
		position: absolute;
		inset: 2rem;
		border: 2px solid rgb(223 189 101 / 35%);
		content: '';
		transform: rotate(4deg);
	}

	.role-fallback::after {
		transform: rotate(-4deg);
	}

	.role-fallback span {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: clamp(9rem, 35vw, 17rem);
		opacity: 0.55;
	}

	.hero-gradient {
		z-index: 1;
		background: linear-gradient(
			to bottom,
			transparent 25%,
			rgb(16 10 7 / 28%) 52%,
			rgb(16 10 7 / 94%) 100%
		);
	}

	.hero-copy {
		z-index: 2;
		align-self: end;
		padding: clamp(var(--space-5), 7vw, var(--space-7));
	}

	.hero-copy h1,
	.hero-copy p {
		margin: 0;
		color: var(--paper-light);
		text-shadow: 0 2px 6px rgb(0 0 0 / 70%);
	}

	.hero-copy h1 {
		font-size: clamp(2.5rem, 10vw, 5.5rem);
	}

	.hero-copy > p:first-child {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	.description {
		max-width: 40rem;
		font-size: clamp(1rem, 3vw, 1.25rem);
	}
</style>
