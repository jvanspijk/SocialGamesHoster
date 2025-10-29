<script lang="ts">
	let {
		initialSeconds = 120,
		remainingTime = $bindable(120),
        onFinished
	}: {
		initialSeconds: number;
		remainingTime: number;
        onFinished?: () => void;
	} = $props();

	let startTime = $state(Date.now());
	let initialDuration = $state(initialSeconds);
    let lastUpdate = $state(0);
    let isRunning = $state(true);

	const R = 45;
	const C = 2 * Math.PI * R;

	const progressOffset = $derived(
		C - (Math.max(0, remainingTime) / initialDuration) * C
	);

	$effect(() => {
		initialDuration = initialSeconds;
		remainingTime = initialSeconds;
		startTime = Date.now();
	});

	$effect(() => {
        if (!isRunning) return;

        let frame: number;

        const tick = () => {
            const now = Date.now();
            const deltaTime = (now - lastUpdate) / 1000; 
            remainingTime = Math.max(0, remainingTime - deltaTime);         
            lastUpdate = now;

            if (remainingTime > 0) { 
                frame = requestAnimationFrame(tick);
            }
        };

        if (remainingTime > 0) {
            lastUpdate = Date.now();
            tick();
        } else {            
            isRunning = false;
            if(onFinished) onFinished();
        }
        
        return () => cancelAnimationFrame(frame);
    });

	function formatTime(totalSeconds: number): string {
		const safe = Math.max(0, Math.floor(totalSeconds));
		const m = Math.floor(safe / 60);
		const s = safe % 60;
		return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
	}

	export function addSeconds(delta: number): void {
        if(delta <= 0) {
            return;
        }
        
        remainingTime += delta; 
        isRunning = true;       

        if(remainingTime > initialDuration)	{
            initialDuration = remainingTime;
        }
        
	}

	export function removeSeconds(delta: number): void {
		if (delta <= 0) {
			return;
		}
        remainingTime = Math.max(0, remainingTime - delta);
	}

    export function setRemainingTime(seconds: number): void {
        if(seconds < 0) {
            return;
        } 
        else if(seconds > 0) {
            isRunning = true; 
        }

        remainingTime = seconds;       

        if(remainingTime > initialDuration)	{
            initialDuration = remainingTime;
        }
    }

    export function togglePause(): void {
        isRunning = !isRunning;
    }

    export function stop(): void {
        isRunning = false;
        remainingTime = 0;
        if(onFinished) onFinished();        
    }
</script>

<div class="timer-wrapper">
    <svg class="progress-ring" viewBox="0 0 100 100">
        <circle
            class="ring-track"
            cx="50"
            cy="50"
            r="{R}"
        />
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
</div>

<style>
    @import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@400;700&display=swap');

    .timer-wrapper {
        position: relative;
        width: 150px; 
        height: 150px;
        background-color: #3b332d;
        border-radius: 50%;
        box-shadow: 0 4px 8px rgba(0, 0, 0, 0.4), inset 0 0 10px rgba(0, 0, 0, 0.5);
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
        font-size: 1.8em;
        font-weight: 700;
        color: #f7e7c4;         
        user-select: none;
    }
</style>