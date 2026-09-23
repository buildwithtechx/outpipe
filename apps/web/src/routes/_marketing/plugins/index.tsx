import { createFileRoute } from '@tanstack/react-router';
import { PluginsPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/plugins/')({
  head: () =>
    createSeo({
      title: 'Plugins and SDKs — Outpipe',
      description:
        'Outpipe plugins and SDKs for TypeScript, React, Vite, Next.js, NestJS, Express, Go, Rust, PHP, and Angular.',
      path: '/plugins',
    }),
  component: PluginsPage,
});
