<script lang="ts">
	import '../app.css';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { LogOut, Shield, UserRound, Swords } from '@lucide/svelte';
	import ConnectionBadge from '$lib/components/ConnectionBadge.svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/state/auth.svelte';
	import { displayPreferences } from '$lib/state/display.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();

	onMount(() => displayPreferences.init());

	async function logout() {
		try {
			await api('/auth/logout', { method: 'POST' });
		} finally {
			auth.clear();
			await goto(resolve('/'));
		}
	}
</script>

<svelte:head>
	<title>Social Games Hoster</title>
</svelte:head>

{#if page.url.pathname.startsWith('/play') && auth.isPlayer}
	<div class="immersive-shell">
		<main class="immersive-page">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="sheet">
		<header>
			<a class="brand" href={resolve('/')} aria-label="Social Games Hoster home">
				<Swords size={25} strokeWidth={1.7} />
				<span>Social Games Hoster</span>
			</a>
			<nav aria-label="Main navigation">
				<a class:active={page.url.pathname.startsWith('/play')} href={resolve('/play')}>
					<UserRound size={17} /> Play
				</a>
				<a class:active={page.url.pathname.startsWith('/admin')} href={resolve('/admin')}>
					<Shield size={17} /> Host
				</a>
				{#if auth.authenticated}
					<button aria-label="Sign out" onclick={logout}><LogOut size={18} /></button>
				{/if}
			</nav>
			<ConnectionBadge />
		</header>
		<main class="page">
			{@render children()}
		</main>
	</div>
{/if}

<style>
	header {
		position: relative;
		z-index: var(--layer-navigation);
		display: grid;
		grid-template-columns: 1fr auto auto;
		align-items: center;
		gap: 1rem;
		border-bottom: var(--border-subtle);
		padding: 0.75rem clamp(1rem, 4vw, 2.5rem);
	}

	.brand {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: 0.6rem;
		color: var(--ink);
		font-family: var(--font-display);
		font-size: clamp(0.72rem, 2.8vw, 1rem);
		font-weight: 700;
		letter-spacing: 0.07em;
		text-decoration: none;
		text-transform: uppercase;
	}

	nav {
		display: flex;
		align-items: center;
		gap: 0.2rem;
	}

	nav a,
	nav button {
		display: inline-flex;
		min-width: var(--target-size);
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		gap: 0.3rem;
		border: 0;
		border-bottom: 2px solid transparent;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-decoration: none;
		text-transform: uppercase;
	}

	nav a:hover,
	nav a.active,
	nav button:hover {
		border-color: var(--crimson);
		color: var(--crimson-dark);
	}

	@media (max-width: 650px) {
		header {
			grid-template-columns: 1fr auto;
		}

		header > :global(span:last-child) {
			grid-column: 1 / -1;
			justify-self: end;
		}

		.brand span {
			display: none;
		}

		nav a {
			font-size: 0;
		}

	}
</style>
