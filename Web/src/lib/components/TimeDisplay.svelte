<script lang="ts">
    import { timersHub } from "$lib/client/Timers/TimersHub.svelte";
    let lastProcessedTs = $state(0);
    let {
        initialSeconds = 0,
        remainingTime = $bindable(0),
        isTimerRunning = false,
        onFinished,
        onclick
    }: {
        initialSeconds: number;
        remainingTime: number;
        isTimerRunning?: boolean;
        onFinished?: () => void;
        onclick?: () => void;
    } = $props();


    let internalTotal = $state(0);
    let isRunning = $state(false); 
    let lastUpdate = $state(Date.now());

    $effect(() => {
        internalTotal = initialSeconds;
        isRunning = isTimerRunning;
    });

    const R = 45;
    const C = 2 * Math.PI * R;
    const progressOffset = $derived(
        C - (Math.max(0, remainingTime) / (internalTotal || 1)) * C
    );

    const handleSync = (payload: { remainingSeconds: number; totalSeconds: number }) => {
        remainingTime = payload.remainingSeconds;
        internalTotal = payload.totalSeconds;
        isRunning = true;
        lastUpdate = Date.now();
    };

    timersHub.onEvent('OnTimerStarted', handleSync);
    timersHub.onEvent('OnTimerResumed', handleSync);
    timersHub.onEvent('OnTimerChanged', handleSync);

    timersHub.onEvent('OnTimerPaused', (payload) => {
        remainingTime = payload.remainingSeconds;
        isRunning = false;
    });

    timersHub.onEvent('OnTimerStopped', () => {
        isRunning = false;
        remainingTime = 0;
        onFinished?.();
    });

    timersHub.onEvent('OnTimerFinished', () => {
        isRunning = false;
        remainingTime = 0;
        onFinished?.();
    });

    $effect(() => {
        if (!isRunning || remainingTime <= 0) return;

        let frame: number;
        const tick = () => {
            const now = Date.now();
            const delta = (now - lastUpdate) / 1000;
            lastUpdate = now;

            remainingTime = Math.max(0, remainingTime - delta);
            
            if (remainingTime > 0) {
                frame = requestAnimationFrame(tick);
            } else {
                isRunning = false;
                onFinished?.();
            }
        };

        frame = requestAnimationFrame(tick);
        return () => cancelAnimationFrame(frame);
    });

    function formatTime(seconds: number): string {
        const safe = Math.max(0, Math.floor(seconds));
        return `${String(Math.floor(safe / 60)).padStart(2, '0')}:${String(safe % 60).padStart(2, '0')}`;
    }

    const isUrgent = $derived(remainingTime < (internalTotal * 0.25));
</script>

<div class="timer-container" class:urgent={isUrgent} class:inactive={!isRunning}>
    <div class="timer-wrapper">
        <svg class="progress-ring" viewBox="0 0 100 100">
            <circle class="ring-track" cx="50" cy="50" r="{R}" />
            <circle
                class="ring-progress"
                cx="50"
                cy="50"
                r="{R}"
                stroke-dasharray="{C}"
                stroke-dashoffset="{progressOffset}"
            />
        </svg>
        <div class="time-text">
            {formatTime(remainingTime)}
        </div>

        {#if onclick}
            <button 
                type="button" 
                class="hitbox" 
                onclick={() => onclick?.()}
                aria-label="Timer action"
            ></button>
        {/if}
    </div>
</div>

<style>
    @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@400;700&display=swap');
    
    .timer-wrapper {
        position: relative;
        width: 75px; 
        height: 75px;
        background-color: #5b4a3c; 
        border-radius: 50%;
        box-shadow: 0 0 5px rgba(255, 255, 255, 0.2), 
                    inset 0 0 15px rgba(0, 0, 0, 0.9);
        border: 2px solid #5b4a3c;
        transition: transform 0.1s ease-out, box-shadow 0.1s ease-out;
        overflow: hidden;
    }

    .timer-wrapper:has(.hitbox:active) {
        transform: scale(0.92);
        box-shadow: 0 0 2px rgba(255, 255, 255, 0.1), 
                    inset 0 0 20px rgba(0, 0, 0, 1);
    }

    .hitbox {
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        z-index: 10;
        
        appearance: none;
        background: transparent;
        border: none;
        padding: 0;
        margin: 0;
        cursor: pointer;
        
        border-radius: 50%;
        
        outline: none;
    }

    .hitbox:hover {
        background: rgba(255, 255, 255, 0.025);
    }

    .hitbox:focus-visible {
        box-shadow: 0 0 0 3px rgba(212, 142, 21, 0.5);
    }

    .inactive {
        opacity: 0.5;
        filter: grayscale(0.6) brightness(0.8);
    }

    .inactive .ring-progress {
        stroke: #a62a2a; 
    }

    .progress-ring {
        position: absolute;
        width: 100%;
        height: 100%;
        transform: rotate(-90deg);
        transition: transform 0.3s ease; 
    }

    .ring-track {
        fill: none;
        stroke: #311e1c; 
        stroke-width: 6; 
    }

    .ring-progress {
        fill: none;
        stroke: #a62a2a; 
        stroke-width: 10;
        stroke-linecap: round;
        transition: stroke-dashoffset 0.05s linear;
    }

    .time-text {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        font-family: 'Cinzel', serif;
        font-size: 1em;
        font-weight: 700;
        color: #f7e7c4;         
        user-select: none;
    }

    .urgent .ring-progress {
        stroke: #d48e15;
        filter: drop-shadow(0 0 5px #d48e15);
    }

    .urgent .time-text {
        color: #fff;
        animation: pulse 1s infinite alternate;
    }

    @keyframes pulse {
        from { opacity: 1; transform: translate(-50%, -50%) scale(1); }
        to { opacity: 0.8; transform: translate(-50%, -50%) scale(1.05); }
    }
</style>