import { afterEach, describe, expect, it, vi } from 'vitest';

afterEach(() => {
	localStorage.clear();
	document.documentElement.removeAttribute('data-text-size');
	document.documentElement.removeAttribute('data-contrast');
	vi.resetModules();
});

describe('displayPreferences', () => {
	it('persists device-local display settings and restores them on initialization', async () => {
		const { displayPreferences } = await import('./display.svelte');
		displayPreferences.init();
		displayPreferences.setLargeText(true);
		displayPreferences.setHighContrast(true);

		expect(JSON.parse(localStorage.getItem('sgh.display-preferences.v1') ?? '{}')).toEqual({
			largeText: true,
			highContrast: true
		});

		vi.resetModules();
		const { displayPreferences: restored } = await import('./display.svelte');
		restored.init();
		expect(restored.largeText).toBe(true);
		expect(restored.highContrast).toBe(true);
		expect(document.documentElement).toHaveAttribute('data-text-size', 'large');
		expect(document.documentElement).toHaveAttribute('data-contrast', 'high');
	});
});
