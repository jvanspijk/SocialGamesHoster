import { render, screen } from '@testing-library/svelte';
import { expect, it } from 'vitest';
import ChatUnreadBadge from './ChatUnreadBadge.svelte';
import type { UnreadCounts } from '../unread.svelte';

it('shows nothing until its promise resolves, then displays the count', async () => {
	let resolve!: (value: UnreadCounts) => void;
	const counts = new Promise<UnreadCounts>((settle) => {
		resolve = settle;
	});
	const { container } = render(ChatUnreadBadge, { counts });
	expect(container.textContent).toBe('');
	resolve({ counts: { general: 3 }, total: 3 });
	expect(await screen.findByText('3')).toBeInTheDocument();
});

it('hides empty counts and pending replacements, then displays the new count', async () => {
	const { container, rerender } = render(ChatUnreadBadge, {
		counts: Promise.resolve({ counts: {}, total: 0 })
	});
	await rerender({ counts: null });
	expect(container.textContent).toBe('');
	await rerender({ counts: Promise.resolve({ counts: {}, total: 99 }) });
	expect(await screen.findByText('99+')).toBeInTheDocument();
	await rerender({ counts: new Promise<UnreadCounts>(() => {}) });
	expect(container.textContent).toBe('');
});

it('shows nothing when the optional count request fails', async () => {
	const { container } = render(ChatUnreadBadge, { counts: Promise.reject(new Error('offline')) });
	await Promise.resolve();
	expect(container.textContent).toBe('');
});
