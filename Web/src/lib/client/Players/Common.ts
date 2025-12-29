export type AbilityInfo = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};

export type RoleInfo = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
	readonly abilities: AbilityInfo[];
};
