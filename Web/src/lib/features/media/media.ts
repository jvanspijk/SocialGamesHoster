import { fetchBlob } from '$lib/api/client';
import type { MediaLoader } from '$lib/components/Media.svelte';

const applicationApiPrefix = '/api/app/v1';

export function loadProtectedMedia(endpoint: string): ReturnType<MediaLoader> {
	return fetchBlob(normalizeMediaEndpoint(endpoint));
}

export function normalizeMediaEndpoint(endpoint: string): string {
	return endpoint.startsWith(applicationApiPrefix)
		? endpoint.slice(applicationApiPrefix.length) || '/'
		: endpoint;
}
