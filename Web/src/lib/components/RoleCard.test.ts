import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import RoleCard from './RoleCard.svelte';

describe('RoleCard', () => {
	it('conceals and reveals every secret with one accessible control', async () => {
		render(RoleCard, {
			props: {
				role: {
					id: 'seer',
					name: 'Seer',
					description: 'Read one player.',
					winCondition: 'The village wins.',
					team: { id: 'village', name: 'Village', description: '' },
					abilities: [{ id: 'read', name: 'Read', description: 'Inspect one player.' }]
				},
				knowledge: [{ participantId: 'p2', seatNumber: 2, role: { name: 'Villager' } }]
			}
		});

		expect(screen.getByRole('heading', { name: 'Seer' })).toBeInTheDocument();
		await fireEvent.click(screen.getByRole('button', { name: /conceal/i }));
		expect(screen.getByText('Role concealed')).toBeInTheDocument();
		expect(screen.getByText('Safe to pass the phone')).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: /reveal/i }));
		expect(screen.queryByText('Role concealed')).not.toBeInTheDocument();
	});
});
