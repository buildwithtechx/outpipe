import { Link } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { BrandLockup } from '#/components/layout';

type AuthPageShellProps = {
  title: string;
  description: string;
  footer: ReactNode;
  children: ReactNode;
};

export function AuthPageShell({
  title,
  description,
  footer,
  children,
}: AuthPageShellProps) {
  return (
    <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-black px-6 py-12 text-white">
      <div className="pointer-events-none absolute left-1/2 top-0 h-96 w-160 -translate-x-1/2 bg-[radial-gradient(ellipse_at_top,rgba(99,102,241,0.1),transparent_68%)]" />
      <section className="relative z-10 w-full max-w-sm text-center">
        <BrandLockup
          className="mx-auto"
          iconClassName="size-11 rounded-2xl"
          nameClassName="text-lg font-semibold tracking-tight"
        />
        <h1 className="mt-8 text-3xl font-semibold tracking-[-0.045em]">
          {title}
        </h1>
        <p className="mt-2 text-sm text-white/45">{description}</p>
        <div className="mt-9">{children}</div>
        {footer && <div className="mt-8 text-sm text-white/45">{footer}</div>}
        <Link
          to="/"
          className="mt-7 inline-block text-xs text-white/35 transition-colors hover:text-indigo-300"
        >
          Back to home
        </Link>
      </section>
    </main>
  );
}
