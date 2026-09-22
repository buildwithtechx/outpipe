import { createFileRoute } from '@tanstack/react-router';
import { SdkPage } from '#/features/marketing';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/_marketing/sdks')({
  head: () =>
    createSeo({
      title: 'SDKs — Outpipe',
      description:
        'Official Outpipe SDKs for TypeScript, Go, Rust, PHP, and Angular.',
      path: '/sdks',
    }),
  component: SdkPage,
});
