import { createEndpoint } from "../api";
export type GetRolesRequest = {
    readonly rulesetId: string;
};
export type GetRolesResponse = {
    readonly id: number;
    readonly name: string;
    readonly description: string;
};
export const GetRoles = createEndpoint<GetRolesRequest, GetRolesResponse[]>('/api/rulesets/{rulesetId}/roles', 'GET');
