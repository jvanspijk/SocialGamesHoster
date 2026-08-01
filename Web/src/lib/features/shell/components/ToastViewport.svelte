<script lang="ts">
	import { onDestroy } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import Toast from '$lib/components/Toast.svelte';
	import { toasts, type ToastMessage } from '$lib/state/toasts.svelte';

	type Timer = {
		started: number;
		remaining: number;
		handle: number;
		hovered: boolean;
		focused: boolean;
	};

	const timers = new SvelteMap<string, Timer>();

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

	onDestroy(() => {
		for (const timer of timers.values()) window.clearTimeout(timer.handle);
	});

	function duration(toast: ToastMessage) {
		return toast.tone === 'error' ? 8_000 : 4_000;
	}

	function start(toast: ToastMessage, remaining = duration(toast)) {
		const previous = timers.get(toast.id);
		if (previous) window.clearTimeout(previous.handle);
		const handle = window.setTimeout(() => {
			timers.delete(toast.id);
			toasts.dismiss(toast.id);
		}, remaining);
		timers.set(toast.id, {
			started: Date.now(),
			remaining,
			handle,
			hovered: previous?.hovered ?? false,
			focused: previous?.focused ?? false
		});
	}

	function pause(toast: ToastMessage, interaction: 'hovered' | 'focused') {
		const timer = timers.get(toast.id);
		if (!timer) return;
		if (timer[interaction]) return;
		const alreadyPaused = timer.hovered || timer.focused;
		timer[interaction] = true;
		if (alreadyPaused) return;
		window.clearTimeout(timer.handle);
		timer.remaining = Math.max(0, timer.remaining - (Date.now() - timer.started));
	}

	function resume(toast: ToastMessage, interaction: 'hovered' | 'focused') {
		const timer = timers.get(toast.id);
		if (!timer || !timer[interaction]) return;
		timer[interaction] = false;
		if (timer.hovered || timer.focused || toast.persistent || timer.remaining <= 0) return;
		start(toast, timer.remaining);
	}

	function focusOut(toast: ToastMessage, event: FocusEvent) {
		if (
			event.currentTarget instanceof HTMLElement &&
			event.relatedTarget instanceof Node &&
			event.currentTarget.contains(event.relatedTarget)
		)
			return;
		resume(toast, 'focused');
	}

	function runAction(toast: ToastMessage) {
		toast.action?.();
		toasts.dismiss(toast.id);
	}
</script>

<section class="toast-viewport" aria-label="Notifications">
	{#each toasts.items as toast (toast.id)}
		<Toast
			tone={toast.tone}
			message={toast.message}
			actionLabel={toast.actionLabel}
			onaction={() => runAction(toast)}
			ondismiss={() => toasts.dismiss(toast.id)}
			onmouseenter={() => pause(toast, 'hovered')}
			onmouseleave={() => resume(toast, 'hovered')}
			onfocusin={() => pause(toast, 'focused')}
			onfocusout={(event) => focusOut(toast, event)}
		/>
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

	:global(.toast-viewport > article) {
		pointer-events: auto;
	}
</style>
