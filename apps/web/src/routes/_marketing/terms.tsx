import { createFileRoute } from '@tanstack/react-router';
import { TermsPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/terms')({
  head: () =>
    createSeo({ title: 'Terms of service — Outpipe', path: '/terms' }),
  component: TermsPage,
});
