import { createEndpoint } from "../api";
export type GetRulesetRequest = {
    readonly rulesetId: string;
};
export type GetRulesetResponse = {
    readonly id: number;
    readonly name: string;
    readonly description: string;
};
export const GetRuleset = createEndpoint<GetRulesetRequest, GetRulesetResponse>('/api/rulesets/{rulesetId}', 'GET');
