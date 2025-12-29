import * as signalR from '@microsoft/signalr';
import { browser } from '$app/environment';
import { untrack } from 'svelte';

export type AuthenticationHubEvents = {
	PlayerLoggedIn: { gameId: number; playerId: number; ts: number };
};

class AuthenticationHub {
	private connection: signalR.HubConnection | null = null;

	#state = $state<{ [K in keyof AuthenticationHubEvents]: AuthenticationHubEvents[K] | null }>({
		PlayerLoggedIn: null
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

		this.connection.on('PlayerLoggedIn', (gameId: number, playerId: number) => {
			this.#state.PlayerLoggedIn = { gameId, playerId, ts: Date.now() };
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

	onEvent<K extends keyof AuthenticationHubEvents>(
		key: K,
		callback: (payload: AuthenticationHubEvents[K]) => void
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
export const authenticationHub = new AuthenticationHub(`${apiBase}/api/authentication/hub`);
