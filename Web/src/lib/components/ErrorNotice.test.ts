import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import ErrorNotice from './ErrorNotice.svelte';

describe('ErrorNotice', () => {
	it('shows actionable form guidance without technical details for expected input errors', () => {
		render(ErrorNotice, {
			props: { message: 'Please correct the highlighted details.' }
		});

		expect(screen.getByRole('alert')).toHaveAttribute('aria-live', 'assertive');
		expect(screen.getByRole('alert')).toHaveTextContent('Please correct the highlighted details.');
		expect(screen.queryByText('Technical details')).not.toBeInTheDocument();
		expect(screen.queryByText(/validation error/i)).not.toBeInTheDocument();
	});

	it('keeps technical details optional and separate from the recovery message', async () => {
		render(ErrorNotice, {
			props: { message: 'The profile could not be saved.', traceId: 'trace-123' }
		});

		expect(screen.getByText('Technical details')).toBeVisible();
		await fireEvent.click(screen.getByText('Technical details'));
		expect(screen.getByText('trace-123')).toBeVisible();
	});
});
