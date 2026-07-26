import { browser } from '$app/environment';

type ConnectionState = 'connected' | 'reconnecting' | 'offline';

let state = $state<ConnectionState>(browser && !navigator.onLine ? 'offline' : 'connected');

if (browser) {
	window.addEventListener('online', () => {
		state = 'reconnecting';
	});
	window.addEventListener('offline', () => {
		state = 'offline';
	});
}

export const connection = {
	get state() {
		return state;
	},
	set(value: ConnectionState) {
		state = value;
	}
};
