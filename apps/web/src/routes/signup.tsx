import { createFileRoute } from '@tanstack/react-router';
import { SignupPage } from '#/features/auth';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/signup')({
  head: () => createSeo({ title: 'Create an account', path: '/signup' }),
  component: SignupPage,
});
