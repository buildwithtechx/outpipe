import { SiGithub, SiGoogle } from 'react-icons/si';
import type { OAuthProvider } from '#/interfaces/auth';

type OAuthProviderButtonProps = {
  provider: OAuthProvider;
  label: string;
  loading: boolean;
  disabled?: boolean;
  onClick: (provider: OAuthProvider) => void;
};

export function OAuthProviderButton({
  provider,
  label,
  loading,
  disabled = false,
  onClick,
}: OAuthProviderButtonProps) {
  const Icon = provider === 'google' ? SiGoogle : SiGithub;
  const providerName = provider === 'google' ? 'Google' : 'GitHub';

  return (
    <button
      type="button"
      aria-label={`${label} ${providerName}`}
      disabled={disabled || loading}
      className="flex h-13 w-full items-center justify-center gap-3 rounded-xl border border-white/10 bg-[#111] text-sm font-medium text-white transition-colors hover:border-indigo-300/45 hover:bg-indigo-300/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-300 disabled:opacity-60"
      onClick={() => onClick(provider)}
    >
      <Icon className="size-4" />
      {loading ? `Connecting to ${providerName}…` : `${label} ${providerName}`}
    </button>
  );
}
