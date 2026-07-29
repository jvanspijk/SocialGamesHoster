import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import ErrorNotice from './ErrorNotice.svelte';

describe('ErrorNotice', () => {
	it('shows actionable form guidance without technical details for expected input errors', () => {
		render(ErrorNotice, {
			error: {
				kind: 'validation',
				message: 'Please correct the highlighted details.',
				fieldErrors: { displayName: 'Enter a valid profile name.' }
			}
		});

		expect(screen.getByRole('alert')).toHaveTextContent('Please correct the highlighted details.');
		expect(screen.queryByText('Technical details')).not.toBeInTheDocument();
		expect(screen.queryByText(/validation error/i)).not.toBeInTheDocument();
	});
});
