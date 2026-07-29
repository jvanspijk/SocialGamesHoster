import { describe, expect, it } from 'vitest';
import { AppApiError } from '$lib/api/client';
import { fieldError, toFormError } from './errors';

describe('form error normalization', () => {
	it('turns API field errors into one reusable validation state', () => {
		const error = toFormError(
			new AppApiError(
				{
					code: 'profile.invalid_name',
					message: 'Enter a valid profile name.',
					fieldErrors: { displayName: ['Use a different name.', 'Try again.'] },
					traceId: 'must-not-be-used'
				},
				422
			),
			'The profile could not be saved.'
		);

		expect(error).toEqual({
			kind: 'validation',
			message: 'Please correct the highlighted details.',
			fieldErrors: { displayName: 'Use a different name.' }
		});
		expect(fieldError(error, 'displayName')).toBe('Use a different name.');
	});

	it('keeps application traces and uses a safe fallback for non-API failures', () => {
		const application = toFormError(
			new AppApiError(
				{ code: 'game.conflict', message: 'The game is no longer editable.', traceId: 'trace-123' },
				409
			),
			'The game could not be updated.'
		);
		const network = toFormError(
			new Error('raw transport detail'),
			'The host could not be reached.'
		);

		expect(application).toMatchObject({
			kind: 'application',
			message: 'The game is no longer editable.',
			traceId: 'trace-123'
		});
		expect(network).toEqual({
			kind: 'network',
			message: 'The host could not be reached.',
			fieldErrors: {}
		});
	});
});
