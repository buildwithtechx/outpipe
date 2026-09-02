import { Button } from '#/components/ui/button';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { useLogout } from '#/features/auth/hooks/use-logout';

export function ProfileSettingsPage() {
  const sessionQuery = useAuthSession();
  const logout = useLogout();

  if (sessionQuery.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading profile…</p>;
  }

  if (sessionQuery.isError || !sessionQuery.user) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load your profile.
      </p>
    );
  }

  const { user } = sessionQuery;

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <h1 className="text-3xl font-semibold tracking-tight">
          Profile settings
        </h1>
        <p className="mt-3 text-sm text-white/55">
          Review the identity connected to your Outpipe account.
        </p>
      </header>
      <section className="mt-8 grid gap-5 rounded-2xl border border-white/10 bg-white/[0.025] p-6">
        <div>
          <p className="text-xs uppercase tracking-wider text-white/40">Name</p>
          <p className="mt-2 text-white/85">{user.name || 'Not provided'}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-wider text-white/40">
            Email
          </p>
          <p className="mt-2 text-white/85">{user.email}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-wider text-white/40">
            Account status
          </p>
          <p className="mt-2 text-emerald-200">{user.status}</p>
        </div>
      </section>
      <section className="mt-8 rounded-2xl border border-white/10 bg-white/[0.025] p-6">
        <h2 className="text-lg font-medium">Session</h2>
        <p className="mt-2 text-sm text-white/50">
          Sign out on shared or temporary machines.
        </p>
        <Button
          type="button"
          variant="outline"
          className="mt-5"
          disabled={logout.isPending}
          onClick={() => logout.mutate()}
        >
          {logout.isPending ? 'Signing out…' : 'Sign out'}
        </Button>
      </section>
    </main>
  );
}
