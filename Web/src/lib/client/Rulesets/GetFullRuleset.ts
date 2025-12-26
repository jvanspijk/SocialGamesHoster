import { createEndpoint } from "../api";
import type { AbilityInfo, RoleInfo } from './common';

export type GetFullRulesetRequest = {
    readonly rulesetId: string;
};
export type GetFullRulesetResponse = {
    readonly id: number;
    readonly name: string;
    readonly description: string;
    readonly abilities: AbilityInfo[];
    readonly roles: RoleInfo[];
};
export const GetFullRuleset = createEndpoint<GetFullRulesetRequest, GetFullRulesetResponse>('/api/rulesets/{rulesetId}/full', 'GET');
