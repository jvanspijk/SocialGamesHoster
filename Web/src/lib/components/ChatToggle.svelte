<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	const dispatch = createEventDispatcher();

	export let hasNewMessage = false;

	function handleToggle() {
		dispatch('click');
	}
</script>

<div class="chat-toggle-container">
	<button class="toggle-button" on:click={handleToggle} aria-label="Open Chat">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="24"
			height="24"
			viewBox="0 0 24 24"
			fill="none"
			stroke="#f7e7c4"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
		</svg>

		{#if hasNewMessage}
			<span class="indicator"></span>
		{/if}
	</button>
</div>

<style>
	.chat-toggle-container {
		position: relative; 
        display: flex;
        flex-direction: column;
        align-items: center;
		z-index: 1000;
	}	

	.toggle-button {
		position: relative;
		background: #5b4a3c;
		color: white;
		border: 2px solid #5b4a3c;
		border-radius: 50%;
		width: 56px;
		height: 56px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s ease;
		box-shadow: 0 0 5px rgba(255, 255, 255, 0.1), 
                    inset 0 0 15px rgba(0, 0, 0, 0.8);
	}

	.toggle-button:hover {
		background: #6d5a4a;
		transform: scale(1.05);
	}

	.toggle-button:active {
		transform: scale(0.95);
	}

	.indicator {
		position: absolute;
		top: 2px;
		right: 2px;
		width: 14px;
		height: 14px;
		background-color: #ff4444;
		border: 2px solid white;
		border-radius: 50%;
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0% { transform: scale(1); box-shadow: 0 0 0 0 rgba(255, 68, 68, 0.7); }
		70% { transform: scale(1.1); box-shadow: 0 0 0 6px rgba(255, 68, 68, 0); }
		100% { transform: scale(1); box-shadow: 0 0 0 0 rgba(255, 68, 68, 0); }
	}
</style>