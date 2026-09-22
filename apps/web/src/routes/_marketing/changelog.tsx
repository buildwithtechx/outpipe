import { createFileRoute } from '@tanstack/react-router';
import { ChangelogPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/changelog')({
  head: () =>
    createSeo({
      title: 'Changelog — Outpipe',
      description: 'Follow new releases and improvements across Outpipe.',
      path: '/changelog',
    }),
  component: ChangelogPage,
});
