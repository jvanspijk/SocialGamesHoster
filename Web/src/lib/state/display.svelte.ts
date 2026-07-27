const storageKey = 'sgh.display-preferences.v1';

let largeText = $state(false);
let highContrast = $state(false);
let initialized = false;

function apply() {
	if (typeof document === 'undefined') return;
	document.documentElement.toggleAttribute('data-text-size', largeText);
	if (largeText) document.documentElement.setAttribute('data-text-size', 'large');
	else document.documentElement.removeAttribute('data-text-size');
	if (highContrast) document.documentElement.setAttribute('data-contrast', 'high');
	else document.documentElement.removeAttribute('data-contrast');
}

function save() {
	if (typeof localStorage !== 'undefined') {
		localStorage.setItem(storageKey, JSON.stringify({ largeText, highContrast }));
	}
	apply();
}

export const displayPreferences = {
	get largeText() {
		return largeText;
	},
	get highContrast() {
		return highContrast;
	},
	init() {
		if (initialized || typeof localStorage === 'undefined') return;
		initialized = true;
		try {
			const stored = JSON.parse(localStorage.getItem(storageKey) ?? '{}') as {
				largeText?: boolean;
				highContrast?: boolean;
			};
			largeText = stored.largeText === true;
			highContrast = stored.highContrast === true;
		} catch {
			largeText = false;
			highContrast = false;
		}
		apply();
	},
	setLargeText(value: boolean) {
		largeText = value;
		save();
	},
	setHighContrast(value: boolean) {
		highContrast = value;
		save();
	}
};
