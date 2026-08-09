import '@testing-library/jest-dom/vitest';

// Node 26 exposes an unconfigured global localStorage accessor before Vitest
// creates jsdom, preventing jsdom from installing its browser-backed storage.
const localStorageEntries = new Map<string, string>();
Object.defineProperty(globalThis, 'localStorage', {
	configurable: true,
	value: {
		get length() {
			return localStorageEntries.size;
		},
		clear() {
			localStorageEntries.clear();
		},
		getItem(key: string) {
			return localStorageEntries.get(String(key)) ?? null;
		},
		key(index: number) {
			return [...localStorageEntries.keys()][index] ?? null;
		},
		removeItem(key: string) {
			localStorageEntries.delete(String(key));
		},
		setItem(key: string, value: string) {
			localStorageEntries.set(String(key), String(value));
		}
	} satisfies Storage
});

if (!HTMLDialogElement.prototype.showModal) {
	HTMLDialogElement.prototype.showModal = function () {
		this.setAttribute('open', '');
	};
}

if (!HTMLDialogElement.prototype.close) {
	HTMLDialogElement.prototype.close = function () {
		this.removeAttribute('open');
		this.dispatchEvent(new Event('close'));
	};
}
