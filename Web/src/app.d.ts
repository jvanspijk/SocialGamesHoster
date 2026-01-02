// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
// interface Ability {
// 	id: number;
// 	name: string;
// 	description: string;
// }

declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			user: { id: number; name: string; roleId: number | null } | null;
			admin: { id: number; name: string; roleId: number | null } | null;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
