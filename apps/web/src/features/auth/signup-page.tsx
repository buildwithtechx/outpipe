import { Link } from '@tanstack/react-router';
import { AuthNotice } from '#/features/auth/components/auth-notice';
import { AuthPageShell } from '#/features/auth/components/auth-page-shell';
import { OAuthProviderButton } from '#/features/auth/components/oauth-provider-button';
import { useAuthNotice } from '#/features/auth/hooks/use-auth-notice';
import { useAuthRedirect } from '#/features/auth/hooks/use-auth-redirect';
import { useOAuthSignIn } from '#/features/auth/hooks/use-oauth-sign-in';

export function SignupPage() {
  const notice = useAuthNotice();
  useAuthRedirect();
  const { provider, signIn } = useOAuthSignIn();

  return (
    <AuthPageShell
      title="Create an account"
      description="Sign up to get started with Outpipe."
      footer={
        <>
          Already have an account?{' '}
          <Link
            to="/login"
            className="font-medium text-indigo-300 hover:text-indigo-200"
          >
            Sign in
          </Link>
        </>
      }
    >
      {notice && <AuthNotice>{notice}</AuthNotice>}
      <div className="space-y-3">
        <OAuthProviderButton
          provider="github"
          label="Continue with"
          loading={provider === 'github'}
          disabled={provider !== null}
          onClick={signIn}
        />
        <OAuthProviderButton
          provider="google"
          label="Continue with"
          loading={provider === 'google'}
          disabled={provider !== null}
          onClick={signIn}
        />
      </div>
      <p className="mt-6 text-center text-xs leading-5 text-white/35">
        By continuing, you agree to the{' '}
        <Link to="/terms" className="underline hover:text-white/60">
          Terms of Service
        </Link>{' '}
        and{' '}
        <Link to="/privacy" className="underline hover:text-white/60">
          Privacy Policy
        </Link>
        .
      </p>
    </AuthPageShell>
  );
}
