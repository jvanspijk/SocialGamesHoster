<script lang="ts">
	import { page } from '$app/state';
	import type { ApiError } from '$lib/client/ApiError';

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function toApiError(value: unknown): ApiError | null {
		if (!isRecord(value)) {
			return null;
		}

		const status = value.status;
		const title = value.title;

		if (typeof status !== 'number' || typeof title !== 'string') {
			return null;
		}

		const detail = typeof value.detail === 'string' ? value.detail : undefined;
		const errors = isRecord(value.errors)
			? Object.fromEntries(
					Object.entries(value.errors).map(([key, messages]) => [
						key,
						Array.isArray(messages)
							? messages.filter((message): message is string => typeof message === 'string')
							: []
					])
				)
			: undefined;

		return { status, title, detail, errors };
	}

	function getErrorMessage(value: unknown): string | null {
		if (!isRecord(value) || typeof value.message !== 'string') {
			return null;
		}

		return value.message;
	}

	const apiError = $derived(toApiError(page.error));
	const validationEntries = $derived(
		Object.entries(apiError?.errors ?? {}).filter(([, messages]) => messages.length > 0)
	);
	const fallbackMessage = $derived(getErrorMessage(page.error));
</script>

<div class="error-sheet">
	<h1>An Error Occurred</h1>
	<p class="intro">Something unexpected happened. Here is the full report.</p>

	<section class="error-block">
		<h2>Status</h2>
		<p class="error-value">{apiError?.status ?? page.status}</p>
	</section>

	<section class="error-block">
		<h2>Title</h2>
		<p class="error-value">{apiError?.title ?? 'Unexpected Error'}</p>
	</section>

	<section class="error-block">
		<h2>Details</h2>
		<p class="error-value">{apiError?.detail ?? fallbackMessage ?? 'No additional details were provided.'}</p>
	</section>

	{#if validationEntries.length > 0}
		<section class="error-block">
			<h2>Validation Issues</h2>
			<div class="validation-list">
				{#each validationEntries as [field, messages] (field)}
					<div class="validation-item">
						<h3>{field}</h3>
						<ul>
							{#each messages as message (message)}
								<li>{message}</li>
							{/each}
						</ul>
					</div>
				{/each}
			</div>
		</section>
	{/if}
</div>

<style>
	.error-sheet {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.intro {
		font-size: 1.15rem;
		font-style: italic;
		margin-bottom: 0.25rem;
		color: #5b4a3c;
	}

	.error-block {
		text-align: left;
		background: #fff9e6;
		border: 2px solid #5b4a3c;
		border-radius: 4px;
		padding: 0.85rem 1rem;
		box-shadow: 2px 2px 0 rgba(91, 74, 60, 0.2);
	}

	.error-block h2 {
		font-size: 1.2rem;
		margin: 0 0 0.45rem;
		color: #8c2a3e;
		border-bottom: 1px solid #8c2a3e;
		padding-bottom: 0.35rem;
	}

	.error-value {
		margin: 0;
		font-size: 1.05rem;
		word-break: break-word;
	}

	.validation-list {
		display: flex;
		flex-direction: column;
		gap: 0.7rem;
	}

	.validation-item {
		background: #f7f3e8;
		border: 1px solid #d4c5a1;
		padding: 0.6rem 0.75rem;
		border-radius: 3px;
	}

	.validation-item h3 {
		margin: 0 0 0.35rem;
		font-family: 'Cinzel', serif;
		font-size: 0.95rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: #5b4a3c;
	}

	.validation-item ul {
		margin: 0;
		padding-left: 1.2rem;
	}

	.validation-item li {
		margin: 0.2rem 0;
	}

	@media (max-width: 650px) {
		.intro {
			font-size: 1.05rem;
		}

		.error-block {
			padding: 0.75rem;
		}
	}
</style>
