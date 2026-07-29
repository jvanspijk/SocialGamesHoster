import { describe, expect, it } from 'vitest';
import { AppApiError } from './client';
import { errorMessage } from './errors';

describe('errorMessage', () => {
	it('returns a standard Error message', () => {
		expect(errorMessage(new Error('Request failed.'), 'Fallback.')).toBe('Request failed.');
	});

	it('returns an AppApiError message', () => {
		const caught = new AppApiError({ code: 'request.failed', message: 'Already decided.' }, 409);

		expect(errorMessage(caught, 'Fallback.')).toBe('Already decided.');
	});

	it.each([
		new Error(''),
		new Error('   '),
		'Thrown string',
		{ message: 'Thrown object' },
		null,
		undefined
	])('returns the fallback for an absent or untrusted message', (caught) => {
		expect(errorMessage(caught, 'Fallback.')).toBe('Fallback.');
	});
});
