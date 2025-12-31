import * as signalR from '@microsoft/signalr';
import { browser } from '$app/environment';
import { untrack } from 'svelte';

export type ChatHubEvents = {
	ChannelCreated: { channelId: number; memberIds: number[]; gameId: number; ts: number };
	MessageSent: { channelId: number; senderId: number; messageId: string; ts: number };
};

class ChatHub {
	private connection: signalR.HubConnection | null = null;

	#state = $state<{ [K in keyof ChatHubEvents]: ChatHubEvents[K] | null }>({
		ChannelCreated: null,
		MessageSent: null
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

		this.connection.on(
			'ChannelCreated',
			(channelId: number, memberIds: number[], gameId: number) => {
				this.#state.ChannelCreated = { channelId, memberIds, gameId, ts: Date.now() };
			}
		);
		this.connection.on('MessageSent', (channelId: number, senderId: number, messageId: string) => {
			this.#state.MessageSent = { channelId, senderId, messageId, ts: Date.now() };
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

	onEvent<K extends keyof ChatHubEvents>(key: K, callback: (payload: ChatHubEvents[K]) => void) {
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
export const chatHub = new ChatHub(`${apiBase}/api/chat/hub`);
