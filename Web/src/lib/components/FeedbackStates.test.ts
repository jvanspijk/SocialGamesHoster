import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Alert from './Alert.svelte';
import EmptyState from './EmptyState.svelte';
import LoadingState from './LoadingState.svelte';

afterEach(cleanup);

describe('Alert', () => {
	it.each([
		['error', 'alert'],
		['success', 'status'],
		['warning', 'status'],
		['info', 'status']
	] as const)('renders the %s tone with semantic live feedback', (tone, role) => {
		render(Alert, { props: { tone, title: `${tone} title`, message: `${tone} message` } });

		const alert = screen.getByRole(role);
		expect(alert).toHaveTextContent(`${tone} title`);
		expect(alert).toHaveTextContent(`${tone} message`);
		expect(alert).toHaveAttribute('aria-live', tone === 'error' ? 'assertive' : 'polite');
	});

	it('offers an optional recovery action', async () => {
		const onaction = vi.fn();
		render(Alert, {
			props: {
				tone: 'error',
				message: 'The requests could not be loaded.',
				actionLabel: 'Retry',
				onaction
			}
		});

		await fireEvent.click(screen.getByRole('button', { name: 'Retry' }));
		expect(onaction).toHaveBeenCalledOnce();
	});
});

describe('EmptyState', () => {
	it('explains an empty result without offering an action', () => {
		render(EmptyState, {
			props: {
				title: 'No conversations',
				description: 'Conversations appear when chat is available.'
			}
		});

		expect(screen.getByRole('heading', { name: 'No conversations' })).toBeVisible();
		expect(screen.getByText('Conversations appear when chat is available.')).toBeVisible();
		expect(screen.queryByRole('button')).not.toBeInTheDocument();
	});

	it('offers one primary action when recovery is available', async () => {
		const onaction = vi.fn();
		render(EmptyState, {
			props: {
				title: 'No players available',
				description: 'Invite a player before starting a conversation.',
				actionLabel: 'Invite player',
				onaction
			}
		});

		await fireEvent.click(screen.getByRole('button', { name: 'Invite player' }));
		expect(onaction).toHaveBeenCalledOnce();
	});
});

describe('LoadingState', () => {
	it('announces a literal loading status without fake content', () => {
		render(LoadingState, { props: { label: 'Loading conversations…' } });

		expect(screen.getByRole('status')).toHaveTextContent('Loading conversations…');
		expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
	});

	it('exposes supplied progress as a native progress bar', () => {
		render(LoadingState, { props: { label: 'Uploading image', progress: 42 } });

		expect(screen.getByRole('progressbar', { name: 'Uploading image: 42%' })).toHaveValue(42);
	});
});
