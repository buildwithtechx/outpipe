import { createFileRoute } from '@tanstack/react-router';
import { PricingPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/pricing')({
  head: () =>
    createSeo({
      title: 'Pricing — Outpipe',
      description:
        'Simple plans for local development tunnels, previews, webhooks, and teams.',
      path: '/pricing',
    }),
  component: PricingPage,
});
