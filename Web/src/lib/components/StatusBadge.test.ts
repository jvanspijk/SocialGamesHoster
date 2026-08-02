import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import StatusBadge from './StatusBadge.svelte';
import StatusBadgeIconFixture from './StatusBadgeIconFixture.svelte';

afterEach(cleanup);

describe('StatusBadge', () => {
	it.each([
		['success', 'tone-success'],
		['warning', 'tone-warning'],
		['danger', 'tone-danger'],
		['info', 'tone-info']
	] as const)('presents the %s tone without application state', (tone, className) => {
		const view = render(StatusBadge, { props: { label: 'Waiting for players', tone, dot: false } });

		expect(screen.getByText('Waiting for players')).toBeVisible();
		expect(view.container.querySelector('.status-badge')).toHaveClass(className);
		expect(view.container.querySelector('i')).not.toBeInTheDocument();
	});

	it('renders an optional icon instead of the status dot', () => {
		const view = render(StatusBadgeIconFixture);

		expect(screen.getByTestId('connection-icon')).toBeVisible();
		expect(view.container.querySelector('i')).not.toBeInTheDocument();
	});
});
