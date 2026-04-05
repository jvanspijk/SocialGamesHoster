<script lang="ts">
	interface Props {
		value?: number;
		placeholder?: string;
		disabled?: boolean;
		onchange?: (seconds: number) => void;
	}

	let {
		value = $bindable(0),
		placeholder = '00:00',
		disabled = false,
		onchange
	}: Props = $props();

	let displayValue = $state('');

	// Initialize display value from seconds
	$effect(() => {
		if (value > 0 && !displayValue) {
			const mins = Math.floor(value / 60);
			const secs = value % 60;
			displayValue = `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
		}
	});

	function handleInput(e: Event) {
		const input = e.target as HTMLInputElement;
		let raw = input.value.replace(/[^0-9]/g, '');

		// Limit to 4 digits (MM:SS)
		if (raw.length > 4) {
			raw = raw.slice(0, 4);
		}

		// Auto-insert colon
		if (raw.length > 2) {
			displayValue = `${raw.slice(0, 2)}:${raw.slice(2)}`;
		} else {
			displayValue = raw;
		}

		// Parse to seconds
		updateValue();
	}

	function handleBlur() {
		// Normalize on blur (pad with zeros)
		const raw = displayValue.replace(/[^0-9]/g, '');
		if (raw.length === 0) {
			displayValue = '';
			value = 0;
			return;
		}

		let mins = 0;
		let secs = 0;

		if (raw.length <= 2) {
			// Treat as minutes only
			mins = parseInt(raw, 10) || 0;
		} else {
			mins = parseInt(raw.slice(0, 2), 10) || 0;
			secs = parseInt(raw.slice(2, 4), 10) || 0;
		}

		// Clamp seconds to 59
		if (secs > 59) secs = 59;

		displayValue = `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
		value = mins * 60 + secs;
		onchange?.(value);
	}

	function updateValue() {
		const raw = displayValue.replace(/[^0-9]/g, '');
		if (raw.length === 0) {
			value = 0;
			return;
		}

		let mins = 0;
		let secs = 0;

		if (raw.length <= 2) {
			mins = parseInt(raw, 10) || 0;
		} else {
			mins = parseInt(raw.slice(0, 2), 10) || 0;
			secs = parseInt(raw.slice(2, 4), 10) || 0;
		}

		value = mins * 60 + secs;
	}

	function handleKeyDown(e: KeyboardEvent) {
		// Allow: backspace, delete, tab, escape, enter, arrows
		const allowedKeys = ['Backspace', 'Delete', 'Tab', 'Escape', 'Enter', 'ArrowLeft', 'ArrowRight', 'Home', 'End'];
		if (allowedKeys.includes(e.key)) return;

		// Block non-numeric
		if (!/^\d$/.test(e.key)) {
			e.preventDefault();
		}
	}
</script>

<input
	type="text"
	inputmode="numeric"
	class="time-input"
	{placeholder}
	{disabled}
	bind:value={displayValue}
	oninput={handleInput}
	onblur={handleBlur}
	onkeydown={handleKeyDown}
	maxlength="5"
	aria-label="Duration in MM:SS format"
/>

<style>
	.time-input {
		font-family: var(--font-heading);
		font-size: 1.5rem;
		text-align: center;
		width: 7rem;
		padding: 0.5rem 0.75rem;
		border: 2px solid var(--color-border);
		border-radius: 3px;
		background: var(--color-surface);
		color: var(--color-text);
		letter-spacing: 0.1em;
		box-shadow: inset 1px 1px 3px rgba(0, 0, 0, 0.1);
		transition: border-color 0.15s ease, box-shadow 0.15s ease;
	}

	.time-input:focus {
		outline: none;
		border-color: var(--color-accent-strong);
		box-shadow:
			inset 1px 1px 3px rgba(0, 0, 0, 0.1),
			0 0 0 2px rgba(166, 42, 42, 0.2);
	}

	.time-input:disabled {
		background: var(--color-surface-alt);
		color: var(--color-text-muted);
		cursor: not-allowed;
	}

	.time-input::placeholder {
		color: var(--color-text-muted);
		opacity: 0.6;
	}
</style>
