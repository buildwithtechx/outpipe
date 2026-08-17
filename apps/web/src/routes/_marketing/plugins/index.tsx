import { createFileRoute } from '@tanstack/react-router';
import { PluginsPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/plugins/')({
  head: () =>
    createSeo({
      title: 'Plugins and SDKs — Outpipe',
      description:
        'Connect Outpipe to React, Vite, Next.js, NestJS, and Express.',
      path: '/plugins',
    }),
  component: PluginsPage,
});
