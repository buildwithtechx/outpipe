import { createFileRoute } from '@tanstack/react-router';
import { LandingPage } from '#/features/marketing';
import { createSeo, siteName } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/')({
  head: () =>
    createSeo({
      title: `${siteName} — Secure tunnels for local development`,
      description:
        'Secure public access for local services, previews, webhooks, and private networks with one CLI, desktop app, and developer protocol.',
      path: '/',
    }),
  component: LandingPage,
});
