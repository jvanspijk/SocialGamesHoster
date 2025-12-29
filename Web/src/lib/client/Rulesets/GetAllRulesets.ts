import { createEndpoint } from '../api';
export type GetAllRulesetsRequest = void;
export type GetAllRulesetsResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const GetAllRulesets = createEndpoint<GetAllRulesetsRequest, GetAllRulesetsResponse[]>(
	'/api',
	'GET'
);
