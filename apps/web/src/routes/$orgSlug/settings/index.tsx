import { createFileRoute, Link } from '@tanstack/react-router';

export const Route = createFileRoute('/$orgSlug/settings/')({
  component: SettingsPage,
});

function SettingsPage() {
  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <h1 className="text-3xl font-semibold tracking-tight">Settings</h1>
      <p className="mt-3 text-sm text-white/55">
        Manage your profile and workspace configuration.
      </p>
      <nav className="mt-8 grid gap-3">
        <Link
          to="/$orgSlug/settings/profile"
          params={{ orgSlug: Route.useParams().orgSlug }}
          className="rounded-2xl border border-white/10 bg-white/[0.025] p-5 hover:border-indigo-300/30"
        >
          <span className="font-medium">Profile</span>
          <span className="mt-1 block text-sm text-white/45">
            Your account identity and status.
          </span>
        </Link>
        <Link
          to="/$orgSlug/settings/organization"
          params={{ orgSlug: Route.useParams().orgSlug }}
          className="rounded-2xl border border-white/10 bg-white/[0.025] p-5 hover:border-indigo-300/30"
        >
          <span className="font-medium">Organization</span>
          <span className="mt-1 block text-sm text-white/45">
            Workspace identity and ownership.
          </span>
        </Link>
      </nav>
    </main>
  );
}
