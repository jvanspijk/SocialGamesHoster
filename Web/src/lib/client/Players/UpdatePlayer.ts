import { createEndpoint } from "../api";
export type UpdatePlayerRequest = {
    readonly id: string;
    readonly newName: string | null;
    readonly newRoleId: number | null;
};
export type UpdatePlayerResponse = {
    readonly id: number;
    readonly name: string;
    readonly roleId: number | null;
};
export const UpdatePlayer = createEndpoint<UpdatePlayerRequest, UpdatePlayerResponse>('/api/players/{id}', 'PATCH');
