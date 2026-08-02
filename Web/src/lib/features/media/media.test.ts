import { describe, expect, it, vi } from 'vitest';

const { fetchBlob } = vi.hoisted(() => ({ fetchBlob: vi.fn().mockResolvedValue(new Blob()) }));
vi.mock('$lib/api/client', () => ({ fetchBlob }));

import { loadProtectedMedia, normalizeMediaEndpoint } from './media';

describe('protected media transport', () => {
	it('normalizes application endpoints before authenticated blob loading', async () => {
		expect(normalizeMediaEndpoint('/api/app/v1/assets/portrait')).toBe('/assets/portrait');
		expect(normalizeMediaEndpoint('/assets/portrait')).toBe('/assets/portrait');

		await loadProtectedMedia('/api/app/v1/assets/portrait');
		expect(fetchBlob).toHaveBeenCalledWith('/assets/portrait');
	});
});
