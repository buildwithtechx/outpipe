import { createFileRoute } from '@tanstack/react-router';
import { LandingPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/')({
  head: () =>
    createSeo({ title: 'Secure tunnels for local development', path: '/' }),
  component: LandingPage,
});
