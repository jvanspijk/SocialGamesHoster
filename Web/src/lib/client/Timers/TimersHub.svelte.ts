import * as signalR from '@microsoft/signalr';
import { browser } from '$app/environment';
import { untrack } from 'svelte';

export type TimersHubEvents = {
	OnTimerStarted: { remainingSeconds: number; totalSeconds: number; ts: number };
	OnTimerChanged: { remainingSeconds: number; totalSeconds: number; ts: number };
	OnTimerPaused: { remainingSeconds: number; totalSeconds: number; ts: number };
	OnTimerResumed: { remainingSeconds: number; totalSeconds: number; ts: number };
	OnTimerStopped: { ts: number };
	OnTimerFinished: { ts: number };
};

class TimersHub {
	private connection: signalR.HubConnection | null = null;

	#state = $state<{ [K in keyof TimersHubEvents]: TimersHubEvents[K] | null }>({
		OnTimerStarted: null,
		OnTimerChanged: null,
		OnTimerPaused: null,
		OnTimerResumed: null,
		OnTimerStopped: null,
		OnTimerFinished: null
	});

	#connectionState = $state<'Disconnected' | 'Connected' | 'Reconnecting' | 'Faulted'>(
		'Disconnected'
	);

	constructor(url: string) {
		if (!browser) return;

		this.connection = new signalR.HubConnectionBuilder()
			.withUrl(url)
			.withAutomaticReconnect()
			.configureLogging(signalR.LogLevel.Information)
			.build();

		this.registerListeners();
		this.start();
	}

	private async start() {
		try {
			await this.connection?.start();
			this.#connectionState = 'Connected';
		} catch (err) {
			console.error('SignalR Start Error: ', err);
			this.#connectionState = 'Faulted';
		}
	}

	private registerListeners() {
		if (!this.connection) return;

		this.connection.on('OnTimerStarted', (remainingSeconds: number, totalSeconds: number) => {
			this.#state.OnTimerStarted = { remainingSeconds, totalSeconds, ts: Date.now() };
		});
		this.connection.on('OnTimerChanged', (remainingSeconds: number, totalSeconds: number) => {
			this.#state.OnTimerChanged = { remainingSeconds, totalSeconds, ts: Date.now() };
		});
		this.connection.on('OnTimerPaused', (remainingSeconds: number, totalSeconds: number) => {
			this.#state.OnTimerPaused = { remainingSeconds, totalSeconds, ts: Date.now() };
		});
		this.connection.on('OnTimerResumed', (remainingSeconds: number, totalSeconds: number) => {
			this.#state.OnTimerResumed = { remainingSeconds, totalSeconds, ts: Date.now() };
		});
		this.connection.on('OnTimerStopped', () => {
			this.#state.OnTimerStopped = { ts: Date.now() };
		});
		this.connection.on('OnTimerFinished', () => {
			this.#state.OnTimerFinished = { ts: Date.now() };
		});

		this.connection.onreconnecting(() => (this.#connectionState = 'Reconnecting'));
		this.connection.onreconnected(() => (this.#connectionState = 'Connected'));
		this.connection.onclose(() => (this.#connectionState = 'Disconnected'));
	}

	get events() {
		return this.#state;
	}
	get status() {
		return this.#connectionState;
	}

	onEvent<K extends keyof TimersHubEvents>(
		key: K,
		callback: (payload: TimersHubEvents[K]) => void
	) {
		$effect.pre(() => {
			const value = this.#state[key];
			if (value) {
				untrack(() => callback(value));
				this.#state[key] = null;
			}
		});
	}

	async disconnect() {
		if (this.connection) {
			await this.connection.stop();
		}
	}
}

const apiBase = browser ? `${window.location.protocol}//${window.location.hostname}:9090` : '';
export const timersHub = new TimersHub(`${apiBase}/api/timers/hub`);
