import { browser } from '$app/environment';

let enabled = $state(browser ? localStorage.getItem('sgh.sound') !== 'off' : true);

export const sound = {
	get enabled() {
		return enabled;
	},
	toggle() {
		enabled = !enabled;
		if (browser) localStorage.setItem('sgh.sound', enabled ? 'on' : 'off');
	},
	set(value: boolean) {
		enabled = value;
		if (browser) localStorage.setItem('sgh.sound', enabled ? 'on' : 'off');
	}
};
