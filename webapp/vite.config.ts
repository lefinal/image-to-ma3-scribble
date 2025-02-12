import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { fileURLToPath } from 'node:url';
import { readFileSync } from 'node:fs';

const path = fileURLToPath(new URL('package.json', import.meta.url));
const pkg = JSON.parse(readFileSync(path, 'utf8'));

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  define: {
    PKG: pkg,
  },
  base: "/apps/image-to-ma3-scribble",
  server: {
    allowedHosts: ["localhost", "localho.st"]
  }
})
