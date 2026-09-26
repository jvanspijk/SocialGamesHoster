<script module lang="ts">
	import type { Snippet } from 'svelte';

	export type ManagementTableColumn = {
		label: string;
		align?: 'start' | 'end';
	};
</script>

<script lang="ts">
	let {
		caption,
		columns,
		children,
		controls,
		empty = false,
		emptyMessage = 'No matching records.'
	}: {
		caption: string;
		columns: ManagementTableColumn[];
		children: Snippet;
		controls?: Snippet;
		empty?: boolean;
		emptyMessage?: string;
	} = $props();
</script>

<div class="management-table">
	{#if controls}
		<div class="controls">{@render controls()}</div>
	{/if}

	{#if empty}
		<p class="empty" role="status">{emptyMessage}</p>
	{:else}
		<div class="table-frame">
			<table>
				<caption>{caption}</caption>
				<thead>
					<tr>
						{#each columns as column (column.label)}
							<th scope="col" class:align-end={column.align === 'end'}>
								{#if column.label}{column.label}{:else}<span class="sr-only">Actions</span>{/if}
							</th>
						{/each}
					</tr>
				</thead>
				<tbody>{@render children()}</tbody>
			</table>
		</div>
	{/if}
</div>

<style>
	.management-table {
		min-width: 0;
	}

	.controls {
		display: flex;
		flex-wrap: wrap;
		align-items: end;
		gap: var(--space-3);
		margin-block-end: var(--space-4);
	}

	.controls :global(.search-field) {
		width: min(100%, 28rem);
		flex: 1 1 16rem;
		margin: 0;
	}

	.controls > :global(label) {
		margin: 0;
	}

	.table-frame {
		border-block: var(--border-subtle);
	}

	table {
		width: 100%;
		border-collapse: collapse;
	}

	caption {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
		white-space: nowrap;
	}

	thead th,
	tbody :global(th),
	tbody :global(td) {
		border-block-end: var(--border-subtle);
		padding: var(--space-3) var(--space-2);
		text-align: left;
		vertical-align: middle;
	}

	thead th {
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
	}

	thead th.align-end {
		text-align: right;
	}

	tbody :global(tr:last-child > *) {
		border-block-end: 0;
	}

	tbody :global(.row-actions) {
		text-align: right;
	}

	tbody :global(.row-actions a),
	tbody :global(.row-actions button) {
		display: inline-flex;
		min-inline-size: var(--target-size);
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		gap: var(--space-1);
		border: 0;
		background: transparent;
		color: var(--action-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		padding-inline: var(--space-1);
		text-decoration: none;
	}

	tbody :global(.row-actions .danger) {
		color: var(--danger);
	}

	.empty {
		border-block: var(--border-subtle);
		color: var(--ink-soft);
		margin: 0;
		padding: var(--space-5) var(--space-2);
	}

	@media (max-width: 47.99rem) {
		.controls {
			align-items: stretch;
		}

		.controls :global(.search-field) {
			width: 100%;
			max-width: none;
			flex-basis: 100%;
		}

		.table-frame {
			border-block-end: 0;
		}

		table,
		tbody :global(tr),
		tbody :global(th),
		tbody :global(td) {
			display: block;
		}

		thead {
			position: absolute;
			width: 1px;
			height: 1px;
			overflow: hidden;
			clip: rect(0 0 0 0);
		}

		tbody :global(tr) {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			gap: var(--space-2) var(--space-4);
			border-block-end: var(--border-subtle);
			padding: var(--space-4) 0;
		}

		tbody :global(tr:last-child) {
			border-block-end: 0;
		}

		tbody :global(th),
		tbody :global(td) {
			border: 0;
			padding: 0;
		}

		tbody :global(th) {
			grid-column: 1 / -1;
		}

		tbody :global(td[data-label])::before {
			display: block;
			margin-block-end: var(--space-1);
			color: var(--ink-soft);
			content: attr(data-label);
			font-family: var(--font-display);
			font-size: var(--font-size-xs);
			font-weight: 700;
		}

		tbody :global(.row-actions) {
			display: flex;
			grid-column: 1 / -1;
			justify-content: flex-start;
			text-align: left;
		}

		tbody :global(.row-actions)::before {
			flex-basis: 100%;
		}
	}
</style>
