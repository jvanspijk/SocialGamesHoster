const BASE_URL = 'http://localhost:9090';

export type ApiError = {
	status: number;
	title: string;
	errors?: Record<string, string[]>;
	detail?: string;
};

export type ApiResponse<T> = { ok: true; data: T } | { ok: false; error: ApiError };

type SvelteFetch = typeof fetch;

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

export interface ApiEndpoint<TReq, TRes> {
	(fetch: SvelteFetch, request: TReq): Promise<TRes>;
	readonly url: string;
	readonly method: HttpMethod;
}

export function createEndpoint<TReq, TRes>(
	url: string,
	method: HttpMethod = 'GET'
): ApiEndpoint<TReq, ApiResponse<TRes>> {
	const endpoint = (async (f: SvelteFetch, request: TReq): Promise<ApiResponse<TRes>> => {
		let finalUrl = BASE_URL + url;

		const requestData: Record<string, unknown> = { ...(request as Record<string, unknown>) };

		// extract and replace path params
		const placeholders = url.match(/\{([^}]+)\}/g) || [];

		for (const placeholder of placeholders) {
			const key = placeholder.slice(1, -1);
			const value = requestData[key];

			if (value !== undefined && value !== null) {
				finalUrl = finalUrl.replace(placeholder, encodeURIComponent(String(value)));
				delete requestData[key];
			} else {
				console.warn(`Missing path parameter: ${key} for URL: ${url}`);
			}
		}

		const options: RequestInit = {
			method,
			headers: { 'Content-Type': 'application/json' },
			credentials: 'include'
		};

		const remainingKeys = Object.keys(requestData);

		if (remainingKeys.length > 0) {
			if (method === 'GET') {
				const queryObj: Record<string, string> = {};
				for (const k of remainingKeys) {
					queryObj[k] = String(requestData[k]);
				}
				const queryString = new URLSearchParams(queryObj).toString();
				if (queryString) {
					finalUrl += (finalUrl.includes('?') ? '&' : '?') + queryString;
				}
			} else {
				options.body = JSON.stringify(requestData);
			}
		}

		//console.debug(`Sending request to ${finalUrl}`)
		const res = await f(finalUrl, options);

		if (res.ok) {
			return { ok: true, data: await res.json() };
		}
		let errorBody: ApiError;
		try {
			errorBody = await res.json();
		} catch {
			errorBody = {
				status: res.status,
				title: 'Unknown Error',
				detail: res.statusText
			};
			if (res.status === 404) {
				console.debug(`endpoint at ${finalUrl} called.`);
			}
		}

		return { ok: false, error: errorBody };
	}) as ApiEndpoint<TReq, ApiResponse<TRes>>;

	return endpoint;
}
