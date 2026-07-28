import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import type { PlayerGameView } from '$lib/api/types';
import RoleReveal from './RoleReveal.svelte';

const view = {
	game: {
		id: 'game',
		name: 'Friday game',
		status: 'running',
		rulesetVersion: 'version',
		joiningOpen: false,
		rolesVisible: true,
		roleVisibilityRevision: 1,
		revision: 1,
		roundNumber: 1,
		phaseKey: 'night'
	},
	participant: {
		id: 'participant',
		displayName: 'Mira',
		gameAlias: '',
		seatNumber: 1,
		status: 'active'
	},
	ruleset: { name: 'Test', description: '' },
	roleAvailable: true,
	roleRevision: 1,
	role: {
		id: 'seer',
		name: 'Seer',
		description: 'Read one player.',
		winCondition: 'The village wins.',
		team: { id: 'village', name: 'Village', description: '' },
		abilities: [{ id: 'read', name: 'Read', description: 'Inspect one player.' }]
	},
	knowledge: [{ participantId: 'p2', seatNumber: 2, role: { name: 'Villager' } }],
	rooms: [],
	attentionItems: [],
	assets: [],
	party: []
} satisfies PlayerGameView;

describe('RoleReveal', () => {
	it('does not render role secrets until the player explicitly reveals them', async () => {
		const reveal = vi.fn();
		const hide = vi.fn();
		const back = vi.fn();
		const rendered = render(RoleReveal, {
			props: { view, revealed: false, reveal, hide, back }
		});

		expect(screen.getByRole('heading', { name: 'Your role is hidden' })).toBeInTheDocument();
		expect(screen.queryByRole('heading', { name: 'Seer' })).not.toBeInTheDocument();
		expect(screen.queryByText('Read one player.')).not.toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Reveal role' }));
		expect(reveal).toHaveBeenCalledOnce();

		await rendered.rerender({ view, revealed: true, reveal, hide, back });
		expect(screen.getByRole('heading', { name: 'Seer' })).toBeInTheDocument();
		expect(screen.getByText('Read one player.')).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Hide role' }));
		expect(hide).toHaveBeenCalledOnce();
	});

	it('shows the private activated state and undo control before phase lock', () => {
		const activated = {
			...view,
			role: {
				...view.role,
				abilities: [
					{
						...view.role.abilities[0],
						activationPhaseIds: ['night'],
						canCombineWithOtherAbilities: false
					}
				]
			},
			abilityChoices: [
				{
					id: 'choice',
					abilityId: 'read',
					abilityName: 'Read',
					status: 'Activated' as const,
					activatedAt: new Date().toISOString()
				}
			]
		} satisfies PlayerGameView;

		render(RoleReveal, {
			props: { view: activated, revealed: true, reveal: vi.fn(), hide: vi.fn(), back: vi.fn() }
		});

		expect(screen.getByText('Activated')).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Undo activation' })).toBeInTheDocument();
		expect(screen.queryByText('Pending')).not.toBeInTheDocument();
	});
});
