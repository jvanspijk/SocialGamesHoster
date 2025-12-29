import * as signalR from '@microsoft/signalr';
import { browser } from '$app/environment';
import { untrack } from 'svelte';

export type RoundsHubEvents = {
	OnRoundStarted: { roundId: number; ts: number };
	OnRoundEnded: { roundId: number; ts: number };
};

class RoundsHub {
	private connection: signalR.HubConnection | null = null;

	#state = $state<{ [K in keyof RoundsHubEvents]: RoundsHubEvents[K] | null }>({
		OnRoundStarted: null,
		OnRoundEnded: null
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

		this.connection.on('OnRoundStarted', (roundId: number) => {
			this.#state.OnRoundStarted = { roundId, ts: Date.now() };
		});
		this.connection.on('OnRoundEnded', (roundId: number) => {
			this.#state.OnRoundEnded = { roundId, ts: Date.now() };
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

	onEvent<K extends keyof RoundsHubEvents>(
		key: K,
		callback: (payload: RoundsHubEvents[K]) => void
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
export const roundsHub = new RoundsHub(`${apiBase}/api/rounds/hub`);
