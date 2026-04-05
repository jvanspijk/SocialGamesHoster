<script lang="ts">
	export let text: string = '';
	export let isHidden: boolean = false;
</script>

{#if !isHidden}
	<p class="description">
		{text}
	</p>
{/if}

<style>
	.description {
		background-color: var(--color-surface-alt);
		border: 1px solid var(--color-border-soft);
		padding: 15px;
		margin-top: 15px;
		border-radius: 3px;
		line-height: 1.5;
		font-size: 1.1em;
		text-align: left;
		color: var(--color-text);
	}
</style>
