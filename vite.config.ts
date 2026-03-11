import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import glsl from 'vite-plugin-glsl';

export default defineConfig(({ mode }) => ({
	assetsInclude: ['**/*.hdr'],
	plugins: [tailwindcss(), sveltekit(), glsl()],
	define:
		mode === 'ssr'
			? {
					// Polyfill for packages that reference window at top level during SSR
					window: 'globalThis'
				}
			: {},
	optimizeDeps: {
		esbuildOptions: {
			loader: {
				'.glsl': 'text',
				'.hdr': 'file'
			}
		}
	},
	ssr: {
		noExternal: ['@tanstack/svelte-query', '@viamrobotics/svelte-sdk', '@viamrobotics/prime-core']
	}
}));
