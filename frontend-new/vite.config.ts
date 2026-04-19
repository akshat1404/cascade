/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
	plugins: [sveltekit()],
	test: {
		// Use the Jest-compatible globals (describe, it, expect…) without imports
		globals: true,
		// jsdom gives us a browser-like environment for DOM-touching tests
		environment: 'jsdom',
		// Where test files live
		include: ['src/tests/**/*.test.ts'],
		// Path aliases that mirror SvelteKit's $lib alias
		alias: {
			'$lib': path.resolve(__dirname, './src/lib'),
		},
		// HTML report + standard console reporter
		reporters: ['verbose'],
	},
});
