<script lang="ts">
    import { fade, scale } from 'svelte/transition';
    import { onMount } from 'svelte';
    import TimeInput from './TimeInput.svelte';
    import SecondaryButton from './SecondaryButton.svelte';
    import { AdjustTimer } from '$lib/client/Timers/AdjustTimer';

    let { onclose }: { onclose: () => void } = $props();

    let totalSeconds = $state(0);
    let containerEl: HTMLDivElement | undefined = $state();

    onMount(() => {
        containerEl?.focus();
    });

    const adjust = async (multiplier: 1 | -1) => {
        const secondsToAdjust = totalSeconds * multiplier;
        if (secondsToAdjust === 0) return;

        const result = await AdjustTimer(fetch, { deltaSeconds: secondsToAdjust });

        if (result.ok) {
            onclose();
        } else {
            alert(result.error.detail ?? 'Failed to adjust timer');
        }
    };

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === 'Escape') {
            onclose(); // 2. Call the prop directly
        }
    }
</script>

<div
    bind:this={containerEl}
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50"
    transition:fade={{ duration: 150 }}
    role="dialog"
    aria-modal="true"
    aria-labelledby="modal-title"
    tabindex="-1"
    onclick={(e) => e.target === e.currentTarget && onclose()}
    onkeydown={handleKeydown}
>
    <div
        class="w-full max-w-md rounded-lg bg-slate-800 p-6 shadow-xl"
        transition:scale={{ duration: 150, start: 0.95 }}
    >
        <h2 id="modal-title" class="mb-4 text-xl font-bold text-slate-100">Adjust Timer</h2>
        <p class="mb-6 text-sm text-slate-400">
            Enter an amount of time to add to or remove from the timer.
        </p>

        <div class="flex justify-center">
            <TimeInput bind:value={totalSeconds} placeholder="00:00" />
        </div>

        <div class="mt-8 flex justify-end gap-3">
            <SecondaryButton 
                variant="danger" 
                onclick={() => adjust(-1)}
            >
                Remove Time
            </SecondaryButton>
            
            <SecondaryButton 
                onclick={() => adjust(1)}
            >
                Add Time
            </SecondaryButton>
        </div>
    </div>
</div>