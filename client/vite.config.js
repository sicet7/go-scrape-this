import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import preprocess from 'svelte-preprocess';

let application_version = process.env.npm_config_application_version ?? process.env.npm_package_version ?? '0.0.0';

// https://vitejs.dev/config/
export default defineConfig({
  define: {
    '__APP_VERSION__': application_version
  },
  plugins: [svelte({ preprocess: preprocess() })],
  server: {
    strictPort: true,
    proxy: {
      '^/api/.*$': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  }
})
