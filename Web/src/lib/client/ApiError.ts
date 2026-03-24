export type ApiError = {
	status: number;
	title: string;
	errors?: Record<string, string[]>;
	detail?: string;
};
