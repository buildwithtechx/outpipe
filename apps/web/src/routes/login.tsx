import { createFileRoute } from '@tanstack/react-router';
import { LoginPage } from '#/features/auth';
import { createSeo } from '#/lib/seo';

export const Route = createFileRoute('/login')({
  head: () => createSeo({ title: 'Sign in', path: '/login' }),
  component: LoginPage,
});
