import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import DisplayPreferencesSettings from './DisplayPreferencesSettings.svelte';

beforeEach(() => {
	localStorage.clear();
	document.documentElement.removeAttribute('data-text-size');
	document.documentElement.removeAttribute('data-contrast');
});

afterEach(cleanup);

describe('DisplayPreferencesSettings', () => {
	it('persists display preferences through the composed setting controls', async () => {
		render(DisplayPreferencesSettings);

		const largeText = screen.getByRole('checkbox', { name: /large text/i });
		await fireEvent.click(largeText);

		expect(largeText).toBeChecked();
		expect(JSON.parse(localStorage.getItem('sgh.display-preferences.v1') ?? '{}')).toEqual({
			largeText: true,
			highContrast: false
		});
		expect(document.documentElement).toHaveAttribute('data-text-size', 'large');
	});
});
