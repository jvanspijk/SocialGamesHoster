import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import VisualDefinitionHarness from '../../../test/VisualDefinitionHarness.svelte';

describe('VisualDefinitionEditor', () => {
	it('creates spoiler-safe point achievements without JSON editing', async () => {
		render(VisualDefinitionHarness);

		expect(screen.getByLabelText('Achievement count')).toHaveTextContent('0');
		await fireEvent.click(screen.getByRole('button', { name: 'Add achievement' }));
		expect(screen.getByLabelText('Achievement count')).toHaveTextContent('1');
		await fireEvent.input(screen.getByLabelText('Achievement points'), { target: { value: '75' } });
		await fireEvent.click(screen.getByLabelText('Hide from players until the game ends'));

		expect(screen.getByLabelText('Saved achievement points')).toHaveTextContent('75');
		expect(screen.getByLabelText('Achievement spoiler visibility')).toHaveTextContent(
			'hidden until complete'
		);
	});
});
