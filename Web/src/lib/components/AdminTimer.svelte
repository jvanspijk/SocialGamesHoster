<script lang="ts">
    import { enhance } from '$app/forms';
    import TimeDisplay from './TimeDisplay.svelte';

    let { timer } = $props<{
        timer: {
            isRunning: boolean;
            totalSeconds: number;
            remainingSeconds: number;
        }
    }>();

    const isTimerFinished = $derived(timer.remainingSeconds === 0);
    const timerAction = $derived(timer.isRunning ? '?/pauseTimer' : '?/resumeTimer');
</script>

<form 
    method="POST" 
    use:enhance={() => {
        if (isTimerFinished) return; 
        return async ({ update }) => {
            await update({ reset: false });
        };
    }}
>
    <div class="timer-interactive-wrapper">
        <button 
            type="submit" 
            formaction={!isTimerFinished ? timerAction : ''}
            class="timer-btn-reset"
            class:non-interactive={isTimerFinished}
        >
            <TimeDisplay 
                initialSeconds={timer.totalSeconds} 
                remainingTime={timer.remainingSeconds}
                isTimerRunning={timer.isRunning}
                onFinished={() => {}}
            />
            
            {#if !isTimerFinished}
                <div class="pause-overlay">
                    {#if timer.isRunning}
                        <svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/></svg>
                    {:else}
                        <svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
                    {/if}
                </div>
            {/if}
        </button>                        
    </div>
</form>

<style>
    .timer-interactive-wrapper {
        position: relative;
        height: 75px;
        display: flex;
        align-items: center;
        max-width: fit-content;
    }

    .timer-btn-reset {
        background: none;
        border: none;
        padding: 0;
        cursor: pointer;
        position: relative;
        flex-shrink: 0;
        /* Reset button default styling to avoid "the box" */
        appearance: none;
        outline: none;
    }

    .non-interactive {
        pointer-events: none;
        cursor: default;
    }

    .pause-overlay {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(0, 0, 0, 0.4);
        border-radius: 50%;
        color: #f7e7c4;
        opacity: 0;
        transition: opacity 0.2s ease;
        pointer-events: none;
        z-index: 20;
    }    

    .timer-btn-reset:hover .pause-overlay {
        opacity: 1;
    }

    .pause-overlay svg {
        width: 35px;
        height: 35px;
    }
</style>