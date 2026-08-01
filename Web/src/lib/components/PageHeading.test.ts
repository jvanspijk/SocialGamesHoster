import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import PageHeading from './PageHeading.svelte';
import PageHeadingHarness from '../../test/PageHeadingHarness.svelte';

describe('PageHeading', () => {
	it('uses one semantic page heading and omits optional copy when it is not supplied', () => {
		render(PageHeading, { props: { title: 'Games' } });

		expect(screen.getByRole('banner')).toContainElement(
			screen.getByRole('heading', { level: 1, name: 'Games' })
		);
		expect(screen.getByRole('heading', { level: 1, name: 'Games' })).toBeVisible();
		expect(screen.queryByText('Manage your games.')).not.toBeInTheDocument();
	});

	it('renders eyebrow, description, and caller-provided actions', () => {
		render(PageHeadingHarness);

		expect(screen.getByText('Account')).toHaveClass('eyebrow');
		expect(screen.getByText('Manage your player identity.')).toBeVisible();
		expect(screen.getByRole('button', { name: 'Open settings' })).toBeVisible();
	});
});
