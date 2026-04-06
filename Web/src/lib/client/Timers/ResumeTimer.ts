// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type ResumeTimerRequest = void;
export type ResumeTimerResponse = {
	readonly remainingSeconds: number;
};
export const ResumeTimer = createEndpoint<ResumeTimerRequest, ResumeTimerResponse>(
	'/api/timers/resume',
	'POST'
);
