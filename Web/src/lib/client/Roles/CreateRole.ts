import { createEndpoint } from "../api";
export type CreateRoleRequest = {
    readonly rulesetId: string;
    readonly name: string;
    readonly description: string | null;
};
export type CreateRoleResponse = {
    readonly id: number;
    readonly name: string;
    readonly description: string;
};
export const CreateRole = createEndpoint<CreateRoleRequest, CreateRoleResponse>('/api/rulesets/{rulesetId}/roles', 'POST');
