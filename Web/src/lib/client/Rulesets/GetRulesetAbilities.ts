import { createEndpoint } from '../api';
export type GetRulesetAbilitiesRequest = {
	readonly rulesetId: string;
};
export type GetRulesetAbilitiesResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const GetRulesetAbilities = createEndpoint<
	GetRulesetAbilitiesRequest,
	GetRulesetAbilitiesResponse[]
>('/api/rulesets/{rulesetId}/abilties', 'GET');
