import starlight from '@astrojs/starlight';
import { defineConfig } from 'astro/config';

export default defineConfig({
  integrations: [
    starlight({
      title: 'Outpipe Docs',
      description:
        'Secure public access for local and private services with one CLI, desktop app, and developer protocol.',
      logo: {
        src: './src/assets/logo.svg',
        alt: 'Outpipe',
      },
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/buildwithtechx/outpipe',
        },
      ],
      head: [
        {
          tag: 'link',
          attrs: {
            rel: 'icon',
            href: '/favicon.svg',
            type: 'image/svg+xml',
          },
        },
      ],
      customCss: ['./src/styles/custom.css'],
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            'getting-started/installation',
            'getting-started/authentication',
            'getting-started/first-tunnel',
            'getting-started/desktop',
          ],
        },
        {
          label: 'Core Concepts',
          items: [
            'concepts/tunnels',
            'concepts/protocol',
            'concepts/edge-routing',
          ],
        },
        {
          label: 'Platform Features',
          items: [
            'platform-features/custom-domains',
            'platform-features/webhooks',
            'platform-features/api-keys',
            'platform-features/organizations',
            'platform-features/password-protection',
            'platform-features/ci-cd',
            'platform-features/observability',
          ],
        },
        {
          label: 'Integrations',
          items: [
            'integrations/overview',
            'integrations/sdk',
            'integrations/go',
            'integrations/rust',
            'integrations/php',
            'integrations/angular',
            'integrations/react',
            'integrations/vite',
            'integrations/next',
            'integrations/nest',
            'integrations/express',
          ],
        },
        {
          label: 'Reference',
          items: ['reference/cli', 'reference/architecture'],
        },
      ],
    }),
  ],
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
