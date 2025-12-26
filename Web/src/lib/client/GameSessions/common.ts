export type RoleInfo = {
    readonly id: number;
    readonly name: string;
};

export type Participant = {
    readonly id: number;
    readonly name: string;
    readonly role: RoleInfo | null;
};

