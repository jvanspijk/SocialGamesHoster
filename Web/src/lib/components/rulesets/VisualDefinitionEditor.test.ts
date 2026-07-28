import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import VisualDefinitionHarness from '../../../test/VisualDefinitionHarness.svelte';
import ChatDefinitionHarness from '../../../test/ChatDefinitionHarness.svelte';
import AbilityDefinitionHarness from '../../../test/AbilityDefinitionHarness.svelte';

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

	it('creates an emoji-only channel with phase visibility rules', async () => {
		render(ChatDefinitionHarness);

		await fireEvent.click(screen.getByRole('button', { name: 'Add channel' }));
		expect(screen.getByLabelText('Custom channel count')).toHaveTextContent('1');

		await fireEvent.change(screen.getByLabelText('Allowed messages'), {
			target: { value: 'emoji_only' }
		});
		await fireEvent.change(screen.getByLabelText('New channel visibility during Night'), {
			target: { value: 'no' }
		});

		expect(screen.getByLabelText('Custom channel restriction')).toHaveTextContent('emoji_only');
		expect(screen.getByLabelText('Custom channel night visibility')).toHaveTextContent('hidden');
	});

	it('configures an ability phase and combination rule', async () => {
		render(AbilityDefinitionHarness);

		await fireEvent.click(screen.getByRole('button', { name: 'Add ability' }));
		await fireEvent.click(screen.getByLabelText('May combine with other combinable abilities'));
		await fireEvent.click(screen.getByLabelText('Night'));

		expect(screen.getByLabelText('Ability phases')).toHaveTextContent('night');
		expect(screen.getByLabelText('Ability combination')).toHaveTextContent('combinable');
	});
});
