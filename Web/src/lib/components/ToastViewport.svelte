<script lang="ts">
	import { CircleCheck, Info, TriangleAlert, X } from '@lucide/svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { toasts, type ToastMessage } from '$lib/state/toasts.svelte';

	const timers = new SvelteMap<string, { started: number; remaining: number; handle: number }>();

	$effect(() => {
		for (const toast of toasts.items) {
			if (!toast.persistent && !timers.has(toast.id)) start(toast);
		}
		for (const [id, timer] of timers) {
			if (!toasts.items.some((item) => item.id === id)) {
				window.clearTimeout(timer.handle);
				timers.delete(id);
			}
		}
	});

	function duration(toast: ToastMessage) {
		return toast.tone === 'error' ? 8_000 : 4_000;
	}

	function start(toast: ToastMessage, remaining = duration(toast)) {
		const handle = window.setTimeout(() => {
			timers.delete(toast.id);
			toasts.dismiss(toast.id);
		}, remaining);
		timers.set(toast.id, { started: Date.now(), remaining, handle });
	}

	function pause(toast: ToastMessage) {
		const timer = timers.get(toast.id);
		if (!timer) return;
		window.clearTimeout(timer.handle);
		timer.remaining = Math.max(0, timer.remaining - (Date.now() - timer.started));
	}

	function resume(toast: ToastMessage) {
		const timer = timers.get(toast.id);
		if (!timer || toast.persistent || timer.remaining <= 0) return;
		start(toast, timer.remaining);
	}

	function runAction(toast: ToastMessage) {
		toast.action?.();
		toasts.dismiss(toast.id);
	}
</script>

<section class="toast-viewport" aria-label="Notifications" aria-live="polite">
	{#each toasts.items as toast (toast.id)}
		<article
			class:error={toast.tone === 'error'}
			class:success={toast.tone === 'success'}
			onmouseenter={() => pause(toast)}
			onmouseleave={() => resume(toast)}
			onfocusin={() => pause(toast)}
			onfocusout={() => resume(toast)}
		>
			<span class="status-icon" aria-hidden="true">
				{#if toast.tone === 'error'}
					<TriangleAlert size={20} />
				{:else if toast.tone === 'success'}
					<CircleCheck size={20} />
				{:else}
					<Info size={20} />
				{/if}
			</span>
			<p>{toast.message}</p>
			{#if toast.actionLabel}
				<button class="action" type="button" onclick={() => runAction(toast)}
					>{toast.actionLabel}</button
				>
			{/if}
			<button
				class="dismiss"
				type="button"
				aria-label="Dismiss notification"
				onclick={() => toasts.dismiss(toast.id)}
			>
				<X size={18} />
			</button>
		</article>
	{/each}
</section>

<style>
	.toast-viewport {
		position: fixed;
		z-index: var(--layer-toast);
		inset-block-start: max(var(--space-3), env(safe-area-inset-top));
		inset-inline-end: max(var(--space-3), env(safe-area-inset-right));
		display: grid;
		width: min(26rem, calc(100vw - 2rem));
		gap: var(--space-2);
		pointer-events: none;
	}

	article {
		display: grid;
		grid-template-columns: auto 1fr auto auto;
		align-items: center;
		gap: var(--space-2);
		border: 1px solid var(--information);
		background: var(--ink);
		box-shadow: var(--shadow);
		color: var(--paper-light);
		padding: var(--space-3);
		pointer-events: auto;
	}

	article.error {
		border-color: var(--danger);
	}

	article.success {
		border-color: var(--success);
	}

	p {
		margin: 0;
	}

	button {
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: inherit;
		cursor: pointer;
	}

	.action {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.dismiss {
		display: grid;
		width: var(--target-size);
		place-items: center;
	}
</style>
