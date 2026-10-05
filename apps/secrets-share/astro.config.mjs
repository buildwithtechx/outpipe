import { defineConfig } from 'astro/config';
import react from '@astrojs/react';

export default defineConfig({
  integrations: [react()],
  vite: {
    build: {
      target: 'es2022',
    },
    oxc: {
      target: 'es2022',
    },
    optimizeDeps: {
      rolldownOptions: {
        transform: {
          target: 'es2022',
        },
      },
    },
  },
});
