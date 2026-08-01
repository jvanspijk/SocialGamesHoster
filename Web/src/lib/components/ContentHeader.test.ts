import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import ContentHeaderHarness from '../../test/ContentHeaderHarness.svelte';

describe('ContentHeader', () => {
	it('renders optional copy and actions without changing the caller-owned heading level', () => {
		render(ContentHeaderHarness);

		expect(screen.getByText('Ruleset')).toHaveClass('eyebrow');
		expect(screen.getByRole('heading', { level: 3, name: 'Teams' })).toBeVisible();
		expect(screen.getByText('Configure the game.')).toHaveClass('description');
		expect(screen.getByRole('button', { name: 'Add team' })).toBeVisible();
	});
});
