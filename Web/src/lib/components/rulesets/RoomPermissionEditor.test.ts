import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import RoomPermissionHarness from '../../../test/RoomPermissionHarness.svelte';

describe('RoomPermissionEditor', () => {
	it('represents a phase override as inherit, yes, or no', async () => {
		render(RoomPermissionHarness);

		const visibility = screen.getByLabelText('Players can see the room');
		expect(visibility).toHaveValue('inherit');
		expect(screen.getByLabelText('Visibility result')).toHaveTextContent('inherit');

		await fireEvent.change(visibility, { target: { value: 'no' } });
		expect(screen.getByLabelText('Visibility result')).toHaveTextContent('no');

		await fireEvent.change(visibility, { target: { value: 'inherit' } });
		expect(screen.getByLabelText('Visibility result')).toHaveTextContent('inherit');
	});
});
