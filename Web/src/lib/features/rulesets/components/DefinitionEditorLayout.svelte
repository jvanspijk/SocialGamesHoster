<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="definition-editor">
	{@render children()}
</div>

<style>
	.definition-editor {
		display: grid;
		gap: var(--space-4);
	}

	.definition-editor :global(.inline-row) {
		display: flex;
		align-items: end;
		justify-content: space-between;
		gap: var(--space-3);
	}

	.definition-editor :global(.subsection) {
		margin-top: var(--space-5);
		border-top: var(--border-strong);
		padding-top: var(--space-4);
	}

	.definition-editor :global(.cards) {
		display: grid;
		gap: var(--space-3);
	}

	.definition-editor :global(.item-card) {
		display: grid;
		gap: var(--space-3);
		border: var(--border-subtle);
		background: rgb(255 249 230 / 34%);
		padding: var(--space-3);
	}

	.definition-editor :global(.item-card h3) {
		font-size: 1rem;
	}

	.definition-editor :global(.form-grid) {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
	}

	.definition-editor :global(.form-grid.thirds) {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	.definition-editor :global(label) {
		display: grid;
		gap: var(--space-1);
	}

	.definition-editor :global(label > span),
	.definition-editor :global(.choice-block > strong),
	.definition-editor :global(.nested-heading strong),
	.definition-editor :global(.override-room strong) {
		font-family: var(--font-display);
		font-size: 0.67rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.definition-editor :global(input:not([type='checkbox'])) {
		width: 100%;
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-2);
	}

	.definition-editor :global(.remove),
	.definition-editor :global(.add-small) {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.62rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		padding: var(--space-1) var(--space-2);
		text-transform: uppercase;
	}

	.definition-editor :global(.add-small) {
		color: var(--ink);
	}

	.definition-editor :global(.nested) {
		display: grid;
		gap: var(--space-3);
		border-inline-start: var(--border-strong);
		padding: var(--space-3);
	}

	.definition-editor :global(.inline-row > label) {
		flex: 1;
	}

	.definition-editor :global(.selector-grid),
	.definition-editor :global(.audience-grid),
	.definition-editor :global(.phase-permissions) {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
	}

	.definition-editor :global(.phase-permissions > div) {
		display: grid;
		gap: var(--space-2);
		border: var(--border-subtle);
		padding: var(--space-2);
	}

	.definition-editor :global(.phase-permissions b) {
		font-family: var(--font-display);
		font-size: 0.78rem;
	}

	.definition-editor :global(.hint.compact) {
		border: 0;
		background: transparent;
		padding: var(--space-1) 0 0;
	}

	.definition-editor :global(.override-room) {
		display: grid;
		gap: var(--space-2);
		border-top: var(--border-subtle);
		padding-top: var(--space-2);
	}

	.definition-editor :global(.empty),
	.definition-editor :global(.hint) {
		border: 1px dashed var(--gold-dark);
		background: rgb(255 249 230 / 30%);
		margin: 0;
		padding: var(--space-3);
	}

	.definition-editor :global(.hint) {
		color: var(--ink-soft);
		font-size: 0.82rem;
	}

	@media (max-width: 720px) {
		.definition-editor :global(.form-grid),
		.definition-editor :global(.form-grid.thirds),
		.definition-editor :global(.selector-grid),
		.definition-editor :global(.audience-grid),
		.definition-editor :global(.phase-permissions) {
			grid-template-columns: 1fr;
		}

		.definition-editor :global(.inline-row) {
			align-items: stretch;
			flex-direction: column;
		}
	}
</style>
