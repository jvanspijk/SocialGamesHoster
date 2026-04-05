from pathlib import Path

api_str = r"""
import type { ApiError } from "./ApiError";

function getBaseUrl(): string {
    if (typeof window === 'undefined') {
        return 'http://localhost:9090';
    }
    const { protocol, hostname } = window.location;
    return `${protocol}//${hostname}:9090`;
}

const BASE_URL = getBaseUrl();

export type ApiResponse<T> = 
    | { ok: true; data: T } 
    | { ok: false; error: ApiError };

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
    const endpoint = (async (f: SvelteFetch, request: TReq, token: string | undefined = undefined): Promise<ApiResponse<TRes>> => {
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

		const headers: Record<string, string> = {
			'Content-Type': 'application/json'
		};

		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		const options: RequestInit = {
			method,
			headers,
			credentials: 'include'
		};	

        const remainingKeys = Object.keys(requestData);
        
        if (remainingKeys.length > 0) {
            if (method === 'GET') {
                const queryObj: Record<string, string> = {};
				for (const k of remainingKeys) {
					const value = requestData[k];
					if (value !== undefined && value !== null) {
						queryObj[k] = String(value);
					}
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
            const contentType = res.headers.get("content-type");
            if (res.status === 204 || !contentType || !contentType.includes("application/json")) {
                return { ok: true, data: {} as TRes };
            }
            const text = await res.text();
            const data = text ? JSON.parse(text) : {};
            return { ok: true, data };
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
            if(res.status === 404) {
                console.debug(`endpoint at ${finalUrl} called.`)
            }
        }

        return { ok: false, error: errorBody };
    }) as ApiEndpoint<TReq, ApiResponse<TRes>>;

    return endpoint;        
}
"""

def create_api_sdk(output_base_path: Path):
    output_file = output_base_path / "Api.ts"
    with open(output_file, "w") as f:
        f.write(api_str)
        f.close()