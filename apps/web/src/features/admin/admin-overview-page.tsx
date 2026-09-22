import { Link } from '@tanstack/react-router';
import type { ReactNode } from 'react';
import { useAdminOverview } from './hooks/use-admin-resources';

export function AdminOverviewPage() {
  const query = useAdminOverview();

  if (query.isLoading) {
    return <AdminState text="Loading platform overview…" />;
  }

  if (query.isError || !query.data) {
    return <AdminState text="We could not load the platform overview." error />;
  }

  const cards = [
    ['Users', query.data.users, '/admin/users'],
    ['Organizations', query.data.organizations, '/admin/organizations'],
    ['Tunnels', query.data.tunnels, '/admin/tunnels'],
    ['Subscriptions', query.data.subscriptions, '/admin/subscriptions'],
  ] as const;

  return (
    <AdminShell
      title="Platform overview"
      subtitle="A quick read on the Outpipe control plane."
    >
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map(([label, value, to]) => (
          <Link
            key={label}
            to={to}
            className="rounded-2xl border border-white/10 bg-white/[0.025] p-5 transition hover:border-indigo-300/40 hover:bg-indigo-300/[0.05]"
          >
            <p className="text-sm text-white/50">{label}</p>
            <p className="mt-3 text-3xl font-semibold">{value}</p>
            <span className="mt-5 block text-sm text-indigo-200">
              Open view →
            </span>
          </Link>
        ))}
      </div>
    </AdminShell>
  );
}

export function AdminShell({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
}) {
  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-purple-400">Platform administration</p>
        <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">{title}</h1>
        <p className="mt-1 text-sm text-white/55">{subtitle}</p>
      </header>
      <section className="pt-2">{children}</section>
    </div>
  );
}
function AdminState({
  text,
  error = false,
}: {
  text: string;
  error?: boolean;
}) {
  return (
    <p className={`p-8 text-sm ${error ? 'text-rose-200' : 'text-white/55'}`}>
      {text}
    </p>
  );
}
