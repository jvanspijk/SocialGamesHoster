import type { Profile } from '$lib/api/types';

const accents = ['crimson', 'forest', 'navy', 'gold', 'plum'] as const;
type Accent = (typeof accents)[number];

let accent = $state<Accent>('crimson');

function isAccent(value: string): value is Accent {
	return accents.includes(value as Accent);
}

function apply() {
	if (typeof document === 'undefined') return;
	if (accent === 'crimson') document.documentElement.removeAttribute('data-accent');
	else document.documentElement.setAttribute('data-accent', accent);
}

export const profilePreferences = {
	get accent() {
		return accent;
	},
	applyProfile(profile: Pick<Profile, 'accent'>) {
		accent = isAccent(profile.accent) ? profile.accent : 'crimson';
		apply();
	},
	setAccent(value: string) {
		if (!isAccent(value)) return;
		accent = value;
		apply();
	}
};
