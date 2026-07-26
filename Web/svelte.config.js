import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({ fallback: 'index.html' }),
		csp: {
			mode: 'hash',
			directives: {
				'default-src': ["'self'"],
				'script-src': ["'self'"],
				'style-src': ["'self'", "'unsafe-inline'"],
				'img-src': ["'self'", 'data:', 'blob:'],
				'media-src': ["'self'", 'blob:'],
				'connect-src': ["'self'"],
				'font-src': ["'self'"],
				'object-src': ["'none'"],
				'base-uri': ["'none'"],
				'form-action': ["'self'"]
			}
		}
	}
};

export default config;
