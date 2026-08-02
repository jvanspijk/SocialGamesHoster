import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Composer from './Composer.svelte';
import MessageList from './MessageList.svelte';
import ChatRail from '$lib/features/chat/components/ChatRail.svelte';

afterEach(cleanup);

describe('chat presentation primitives', () => {
	it('exposes a semantic loading state for the conversation rail', () => {
		render(ChatRail, { props: { entries: [], loading: true, onselect: vi.fn() } });

		expect(screen.getByRole('status')).toHaveTextContent('Loading conversations…');
	});

	it('renders a searchable rich rail with selected and unread conversations', async () => {
		const onselect = vi.fn();
		render(ChatRail, {
			props: {
				entries: [
					{
						id: 'general',
						label: 'General',
						accessibleLabel: 'General, New messages',
						description: 'Alex: Hello',
						supportingLabel: 'Game',
						metaLabel: '12:00',
						leadingText: 'G',
						leadingVariant: 'hash',
						unread: true
					}
				],
				selectedId: 'general',
				onselect
			}
		});

		expect(screen.getByRole('heading', { name: 'Conversations' })).toBeInTheDocument();
		expect(screen.getByRole('searchbox', { name: 'Search conversations' })).toBeInTheDocument();
		const conversation = screen.getByRole('button', { name: 'General, New messages' });
		expect(conversation).toHaveAttribute('aria-current', 'page');
		expect(screen.getByText('Alex: Hello')).toBeInTheDocument();

		await fireEvent.click(conversation);
		expect(onselect).toHaveBeenCalledWith('general');
	});

	it('uses the shared empty state for an empty or filtered rail', () => {
		const onnewmessage = vi.fn();
		render(ChatRail, {
			props: { entries: [], search: '', onselect: vi.fn(), onnewmessage }
		});

		expect(screen.getByRole('heading', { name: 'No conversations' })).toBeInTheDocument();
		expect(screen.getByText('Conversations appear when chat is available.')).toBeInTheDocument();
		expect(screen.getAllByRole('button', { name: 'New message' })).toHaveLength(2);

		return fireEvent.click(screen.getAllByRole('button', { name: 'New message' })[1]).then(() => {
			expect(onnewmessage).toHaveBeenCalledOnce();
		});
	});

	it('renders message grouping, unread placement, pagination, and removal actions semantically', async () => {
		const onloadEarlier = vi.fn();
		const onremove = vi.fn();
		render(MessageList, {
			props: {
				messages: [
					{
						id: 'message-1',
						senderLabel: 'Alex',
						timeLabel: '10:00',
						dayKey: '2025-7-27',
						dayLabel: 'Sunday, July 27',
						content: 'Earlier',
						canRemove: true,
						removeLabel: 'Remove message from Alex'
					},
					{
						id: 'message-2',
						senderLabel: 'Blake',
						timeLabel: '11:00',
						dayKey: '2031-7-27',
						dayLabel: 'Sunday, July 27',
						content: 'Latest',
						canRemove: true,
						removeLabel: 'Remove message from Blake'
					}
				],
				firstUnreadId: 'message-2',
				hasEarlierMessages: true,
				emptyDescription: 'Start the conversation.',
				onloadEarlier,
				onremove
			}
		});

		expect(screen.getAllByRole('separator')).toHaveLength(3);
		expect(screen.getAllByRole('separator', { name: 'Sunday, July 27' })).toHaveLength(2);
		expect(screen.getByRole('separator', { name: 'New messages' })).toBeInTheDocument();
		await fireEvent.click(screen.getByRole('button', { name: 'Load earlier messages' }));
		await fireEvent.click(screen.getByRole('button', { name: 'Remove message from Blake' }));
		expect(onloadEarlier).toHaveBeenCalledOnce();
		expect(onremove).toHaveBeenCalledWith('message-2');
	});

	it('renders loading and empty message states', () => {
		const { rerender } = render(MessageList, {
			props: { messages: [], loading: true, emptyDescription: 'Start the conversation.' }
		});

		expect(screen.getByRole('status')).toHaveTextContent('Loading messages…');
		return rerender({ loading: false }).then(() => {
			expect(screen.getByRole('heading', { name: 'No messages yet' })).toBeInTheDocument();
			expect(screen.getByText('Start the conversation.')).toBeInTheDocument();
		});
	});

	it('submits the composer on its named send action and Enter', async () => {
		const onsubmit = vi.fn();
		render(Composer, { props: { value: '', onsubmit } });

		const message = screen.getByRole('textbox', { name: 'Message' });
		const send = screen.getByRole('button', { name: 'Send' });
		expect(send).toBeDisabled();
		message.focus();
		await fireEvent.input(message, { target: { value: 'Hello' } });
		expect(send).toBeEnabled();
		await fireEvent.keyDown(message, { key: 'Enter' });
		expect(onsubmit).toHaveBeenCalledOnce();
		expect(message).toHaveFocus();
	});
});
