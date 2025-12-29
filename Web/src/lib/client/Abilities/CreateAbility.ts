import { createEndpoint } from '../api';
export type CreateAbilityRequest = {
	readonly rulesetId: string;
	readonly name: string;
	readonly description: string;
};
export type CreateAbilityResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const CreateAbility = createEndpoint<CreateAbilityRequest, CreateAbilityResponse>(
	'/api/rulesets/{rulesetId}/abilties',
	'POST'
);
