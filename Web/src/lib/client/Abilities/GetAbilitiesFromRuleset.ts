import { createEndpoint } from '../api';
export type GetAbilitiesFromRulesetRequest = {
	readonly rulesetId: string;
};
export type GetAbilitiesFromRulesetResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const GetAbilitiesFromRuleset = createEndpoint<
	GetAbilitiesFromRulesetRequest,
	GetAbilitiesFromRulesetResponse[]
>('/api/rulesets/{rulesetId}/abilties', 'GET');
