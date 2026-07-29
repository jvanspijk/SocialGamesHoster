import { AppApiError } from '$lib/api/client';

export type FormErrorKind = 'validation' | 'application' | 'network';

export type FormError = {
	kind: FormErrorKind;
	message: string;
	fieldErrors: Record<string, string>;
	traceId?: string;
};

const validationSummary = 'Please correct the highlighted details.';

export function toFormError(caught: unknown, fallback: string): FormError {
	if (!(caught instanceof AppApiError)) {
		return { kind: 'network', message: fallback, fieldErrors: {} };
	}

	const fieldErrors = firstFieldErrors(caught.body.fieldErrors);
	const isValidation = caught.status === 422 || Object.keys(fieldErrors).length > 0;
	return {
		kind: isValidation ? 'validation' : 'application',
		message:
			isValidation && Object.keys(fieldErrors).length > 0 ? validationSummary : caught.body.message,
		fieldErrors,
		traceId: isValidation ? undefined : caught.body.traceId
	};
}

export function fieldError(error: FormError | null, field: string): string {
	return error?.fieldErrors[field] ?? '';
}

export function fieldErrorOrSummary(error: FormError | null, field: string): string {
	return (
		fieldError(error, field) ||
		(error && Object.keys(error.fieldErrors).length === 0 ? error.message : '')
	);
}

function firstFieldErrors(fields: Record<string, string[]> | undefined): Record<string, string> {
	return Object.fromEntries(
		Object.entries(fields ?? {})
			.map(([field, messages]) => [field, messages[0] ?? ''] as const)
			.filter(([, message]) => message !== '')
	);
}
