import * as signalR from '@microsoft/signalr';
import { browser } from '$app/environment';
import { untrack } from 'svelte';

export type PlayersHubEvents = {
    PlayerUpdated: { playerId: number; ts: number; };
};

class PlayersHub {
    private connection: signalR.HubConnection | null = null;

    #state = $state<{ [K in keyof PlayersHubEvents]: PlayersHubEvents[K] | null }>({
        PlayerUpdated: null,
    });

    #connectionState = $state<"Disconnected" | "Connected" | "Reconnecting" | "Faulted">("Disconnected");

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
            this.#connectionState = "Connected";
        } catch (err) {
            console.error("SignalR Start Error: ", err);
            this.#connectionState = "Faulted";
        }
    }

    private registerListeners() {
        if (!this.connection) return;

        this.connection.on("PlayerUpdated", (playerId: number) => {
            this.#state.PlayerUpdated = { playerId, ts: Date.now() };
        });

        this.connection.onreconnecting(() => this.#connectionState = "Reconnecting");
        this.connection.onreconnected(() => this.#connectionState = "Connected");
        this.connection.onclose(() => this.#connectionState = "Disconnected");
    }

    get events() { return this.#state; }
    get status() { return this.#connectionState; }

    onEvent<K extends keyof PlayersHubEvents>(
        key: K, 
        callback: (payload: PlayersHubEvents[K]) => void
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
export const playersHub = new PlayersHub(`${apiBase}/api/players/hub`);
