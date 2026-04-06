// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type CreateAbilityRequest = {
	readonly rulesetId: number;
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
