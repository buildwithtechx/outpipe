import { createFileRoute } from '@tanstack/react-router';
import { PrivacyPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/privacy')({
  head: () =>
    createSeo({ title: 'Privacy policy — Outpipe', path: '/privacy' }),
  component: PrivacyPage,
});
