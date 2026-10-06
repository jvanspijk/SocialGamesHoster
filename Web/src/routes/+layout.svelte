<script lang="ts">
	import '../app.css';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Shield from '@lucide/svelte/icons/shield';
	import UserRound from '@lucide/svelte/icons/user-round';
	import SignOutButton from '$lib/features/shell/components/SignOutButton.svelte';
	import ToastViewport from '$lib/features/shell/components/ToastViewport.svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/state/auth.svelte';
	import { displayPreferences } from '$lib/state/display.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let applicationVersion = $state('');

	onMount(() => {
		displayPreferences.init();
		void loadApplicationVersion();
	});

	async function loadApplicationVersion() {
		try {
			const status = await api<{ version: string }>('/setup/status');
			applicationVersion = status.version;
		} catch {
			applicationVersion = '';
		}
	}
</script>

<svelte:head>
	<title>Social Games Hoster</title>
	<link rel="icon" type="image/png" href="/icons/logo.webp" />
</svelte:head>

{#if page.url.pathname.startsWith('/play') || page.url.pathname.startsWith('/admin')}
	<div class="immersive-shell">
		<div class="canvas-blots-layer" aria-hidden="true"></div>
		<div class="canvas-fractal-layer" aria-hidden="true"></div>
		<div class="canvas-noise-layer" aria-hidden="true"></div>
		<main class="immersive-page">
			{@render children()}
		</main>
	</div>
{:else}
	<div class="sheet">
		<div class="canvas-blots-layer" aria-hidden="true"></div>
		<div class="canvas-fractal-layer" aria-hidden="true"></div>
		<div class="canvas-noise-layer" aria-hidden="true"></div>
		<header>
			<a class="brand" href={resolve('/')} aria-label="Social Games Hoster home">
				<img src="/icons/logo.webp" alt="" width="48" height="48" />
				<span>Social Games Hoster</span>
			</a>
			<nav aria-label="Main navigation">
				<a class:active={page.url.pathname.startsWith('/play')} href={resolve('/play')}>
					<UserRound size={17} /> Play
				</a>
				<a class:active={page.url.pathname.startsWith('/admin')} href={resolve('/admin')}>
					<Shield size={17} /> Manage
				</a>
				{#if auth.authenticated}
					<SignOutButton />
				{/if}
			</nav>
		</header>
		<main class="page">
			{@render children()}
		</main>
		{#if applicationVersion}
			<footer>Version {applicationVersion}</footer>
		{/if}
	</div>
{/if}

<ToastViewport />

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
		color: var(--text-primary);
		font-family: var(--font-display);
		font-size: clamp(var(--font-size-sm), 2.8vw, var(--font-size-base));
		font-weight: 700;
		text-decoration: none;
	}

	nav {
		display: flex;
		align-items: center;
		gap: clamp(0.45rem, 1.5vw, 1rem);
	}

	nav a {
		display: inline-flex;
		min-width: var(--target-size);
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		gap: 0.3rem;
		border: 0;
		border-bottom: 2px solid transparent;
		background: transparent;
		color: var(--text-secondary);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		text-decoration: none;
	}

	nav a:hover,
	nav a.active {
		border-color: var(--action);
		color: var(--action-dark);
	}

	footer {
		border-top: var(--border-subtle);
		color: var(--text-secondary);
		font-size: var(--font-size-sm);
		margin: 2rem clamp(1rem, 4vw, 2.5rem) 0;
		padding-block: 1rem;
		text-align: center;
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
