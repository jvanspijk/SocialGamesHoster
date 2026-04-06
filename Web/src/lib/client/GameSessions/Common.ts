// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
export type RoleInfo = {
	readonly id: number;
	readonly name: string;
};

export type Participant = {
	readonly id: number;
	readonly name: string;
	readonly role: RoleInfo | null;
};

export type Phase = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
