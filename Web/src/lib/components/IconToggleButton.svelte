<script>
let { 
	pressed = $bindable(false), 
	onIconSrc, 
	offIconSrc, 
	label = "Toggle", 
	size = 42 
} = $props();
</script>

<button 
type="button"
class="icon-toggle" 
onclick={() => pressed = !pressed}
aria-pressed={pressed}
aria-label={label}
title={label}
style="width: {size}px; height: {size}px;"
>
{#if pressed}
	<div class="icon-wrapper"><img src={onIconSrc} alt="" /></div>
{:else}
	<div class="icon-wrapper"><img src={offIconSrc} alt="" /></div>
{/if}
</button>

<style>
.icon-toggle {
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	background: linear-gradient(180deg, var(--color-surface-soft) 0%, var(--color-surface-alt) 100%);
	border: 2px solid var(--color-border);
	border-radius: 2px;
	padding: 0;
	font-family: var(--font-heading);
	transition:
		transform 0.12s ease,
		box-shadow 0.12s ease,
		border-color 0.12s ease,
		background-color 0.12s ease;
	overflow: hidden;
	box-shadow:
		0 3px 0 var(--color-border-dark),
		0 5px 10px rgba(0, 0, 0, 0.28),
		inset 0 1px 0 rgba(255, 255, 255, 0.35);
}

.icon-toggle:hover:enabled {
	border-color: var(--color-accent);
	background: linear-gradient(180deg, var(--color-surface) 0%, var(--color-surface-soft) 100%);
	transform: translateY(-1px);
	box-shadow:
		0 4px 0 var(--color-border-dark),
		0 7px 12px rgba(0, 0, 0, 0.34),
		inset 0 1px 0 rgba(255, 255, 255, 0.45);
}

.icon-toggle[aria-pressed="true"] {
	border-color: var(--color-accent-strong);
	background: linear-gradient(180deg, #e9d5b2 0%, var(--color-surface) 100%);
	box-shadow:
		0 2px 0 var(--color-border-dark),
		0 4px 8px rgba(0, 0, 0, 0.25),
		inset 0 0 0 1px rgba(166, 42, 42, 0.2);
}

.icon-toggle:active:enabled,
.icon-toggle[aria-pressed="true"]:active:enabled {
	transform: translateY(2px);
	box-shadow:
		0 1px 0 var(--color-border-dark),
		0 2px 5px rgba(0, 0, 0, 0.28),
		inset 0 1px 2px rgba(0, 0, 0, 0.16);
}

.icon-toggle:focus-visible {
	outline: 3px solid var(--color-focus);
	outline-offset: 2px;
	border-color: var(--color-accent-strong);
}

.icon-wrapper {
	width: 62%;
	height: 62%;
	display: flex;
	align-items: center;
	justify-content: center;
	filter: drop-shadow(0 1px 0 rgba(62, 50, 43, 0.35));
}

.icon-wrapper img {
	width: 100%;
	height: 100%;
	display: block;
	object-fit: contain;
}
</style>
