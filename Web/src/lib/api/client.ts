import PocketBase, { ClientResponseError } from 'pocketbase';
import { browser } from '$app/environment';
import type { Actor, AppErrorBody, AuthResponse } from './types';

export class AppApiError extends Error {
	constructor(
		public readonly body: AppErrorBody,
		public readonly status: number
	) {
		super(body.message);
	}
}

export const pb = new PocketBase(browser ? window.location.origin : 'http://127.0.0.1');
pb.autoCancellation(false);

export type ApiOptions = Omit<RequestInit, 'body'> & { body?: unknown };

export async function api<T>(path: string, init: ApiOptions = {}): Promise<T> {
	const headers = new Headers(init.headers);
	headers.set('Accept', 'application/json');
	if (
		init.body &&
		!(init.body instanceof FormData) &&
		!(init.body instanceof Blob) &&
		!(init.body instanceof ArrayBuffer)
	) {
		headers.set('Content-Type', 'application/json');
	}
	try {
		return await pb.send<T>(`/api/app/v1${path}`, {
			...init,
			headers: Object.fromEntries(headers.entries()),
			requestKey: null
		});
	} catch (caught) {
		if (caught instanceof ClientResponseError) {
			if (caught.status === 401) pb.authStore.clear();
			const data = caught.data as Partial<AppErrorBody>;
			throw new AppApiError(
				{
					code: data.code ?? 'network.unexpected',
					message: data.message ?? 'The host returned an unexpected response.',
					fieldErrors: data.fieldErrors,
					traceId: caught.status === 422 || data.fieldErrors ? undefined : data.traceId
				},
				caught.status
			);
		}
		throw caught;
	}
}

export function saveAuth(response: AuthResponse): Actor {
	const model = response.actor ?? {
		id: response.profile?.id ?? '',
		type: 'player_profiles',
		displayName: response.profile?.displayName ?? ''
	};
	pb.authStore.save(response.token, {
		...model,
		collectionId: model.type,
		collectionName: model.type
	});
	return model as Actor;
}

export function clearAuth() {
	pb.authStore.clear();
}

export function jsonBody(value: unknown): Pick<ApiOptions, 'body'> {
	return { body: value };
}

export async function download(path: string, filename: string, method = 'GET'): Promise<void> {
	const blob = await fetchBlob(path, method);
	const url = URL.createObjectURL(blob);
	const anchor = document.createElement('a');
	anchor.href = url;
	anchor.download = filename;
	anchor.click();
	URL.revokeObjectURL(url);
}

export async function fetchBlob(path: string, method = 'GET'): Promise<Blob> {
	const response = await fetch(`/api/app/v1${path}`, {
		method,
		headers: { Authorization: pb.authStore.token }
	});
	if (!response.ok) {
		if (response.status === 401) pb.authStore.clear();
		let body: Partial<AppErrorBody> = {};
		try {
			body = (await response.json()) as Partial<AppErrorBody>;
		} catch {
			// The standard fallback below intentionally hides raw server responses.
		}
		throw new AppApiError(
			{
				code: body.code ?? 'download.failed',
				message: body.message ?? 'The requested file could not be prepared.',
				traceId: response.status === 422 || body.fieldErrors ? undefined : body.traceId
			},
			response.status
		);
	}
	return response.blob();
}
