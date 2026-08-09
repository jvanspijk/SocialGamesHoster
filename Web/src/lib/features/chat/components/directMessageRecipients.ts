import type { SelectionDialogEntry } from '$lib/components/SelectionDialog.svelte';

export function directMessageRecipient(
	id: string,
	label: string,
	seatNumber: number
): SelectionDialogEntry {
	const supportingLabel = `Seat ${seatNumber}`;
	return {
		id,
		label,
		accessibleLabel: `${label}, ${supportingLabel}`,
		supportingLabel,
		leadingText: label.slice(0, 1).toUpperCase()
	};
}
