import { cleanup, fireEvent, render, screen, within } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import RoleVisibilityControl from './RoleVisibilityControl.svelte';

afterEach(cleanup);

describe('RoleVisibilityControl', () => {
	it.each([
		['readiness', false, 'Roles hidden', 'Roles are hidden from players.', 'Reveal roles'],
		['readiness', true, 'Roles revealed', 'Players can see their own role.', 'Hide'],
		['heading', false, 'Roles hidden', 'Roles are hidden from players.', 'Reveal roles'],
		['heading', true, 'Roles revealed', 'Players can see their own role.', 'Hide roles']
	] as const)(
		'presents the %s control when role visibility is %s',
		(presentation, rolesVisible, status, description, action) => {
			render(RoleVisibilityControl, {
				props: { gameId: 'game-1', rolesVisible, presentation }
			});

			expect(screen.getByText(status)).toBeVisible();
			expect(screen.getByText(description)).toBeVisible();
			expect(screen.getByRole('button', { name: action })).toBeVisible();
		}
	);

	it.each(['readiness', 'heading'] as const)(
		'opens the %s reveal confirmation',
		async (presentation) => {
			render(RoleVisibilityControl, {
				props: { gameId: 'game-1', rolesVisible: false, presentation }
			});

			await fireEvent.click(screen.getByRole('button', { name: 'Reveal roles' }));
			const dialog = screen.getByRole('dialog', { name: 'Reveal roles?' });

			expect(within(dialog).getByRole('button', { name: 'Cancel' })).toBeVisible();
			expect(within(dialog).getByRole('button', { name: 'Reveal roles' })).toBeVisible();
		}
	);
});
