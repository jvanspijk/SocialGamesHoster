import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import PanelHarness from '../../test/PanelHarness.svelte';

describe('Panel', () => {
	it('keeps its section heading while rendering shared description and actions', () => {
		render(PanelHarness);

		expect(screen.getByRole('heading', { level: 2, name: 'Players' })).toBeVisible();
		expect(screen.getByText('People approved for this game.')).toHaveClass('description');
		expect(screen.getByRole('button', { name: 'Invite player' })).toBeVisible();
	});
});
