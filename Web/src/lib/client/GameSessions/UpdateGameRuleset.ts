import { createEndpoint } from "../api";
export type UpdateGameRulesetRequest = {
    readonly gameId: string;
    readonly rulesetId: number;
};
export type UpdateGameRulesetResponse = {
    readonly id: number;
    readonly rulesetId: number;
};
export const UpdateGameRuleset = createEndpoint<UpdateGameRulesetRequest, UpdateGameRulesetResponse>('/api/games/{gameId}/ruleset', 'PATCH');
