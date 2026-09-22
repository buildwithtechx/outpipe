import { Link } from '@tanstack/react-router';
import { CircleAlert, LoaderCircle } from 'lucide-react';

type TunnelPageStateProps = {
  label?: string;
  error?: string;
  compact?: boolean;
  action?: {
    label: string;
    to?: string;
    onClick?: () => void;
  };
};

export function TunnelPageState({
  label,
  error,
  compact = false,
  action,
}: TunnelPageStateProps) {
  const content = error ?? label;

  return (
    <div
      className={`${compact ? 'rounded-2xl border border-white/10 py-10' : 'min-h-[55vh]'} flex items-center justify-center px-6 text-center`}
    >
      <div className="max-w-md">
        {error ? (
          <CircleAlert className="mx-auto size-5 text-rose-200" />
        ) : (
          <LoaderCircle className="mx-auto size-5 animate-spin text-indigo-200" />
        )}
        <p className="mt-3 text-sm leading-6 text-white/55">{content}</p>
        {action &&
          (action.to ? (
            <Link
              to={action.to}
              className="mt-4 inline-block text-sm font-medium text-indigo-200 hover:text-indigo-100"
            >
              {action.label}
            </Link>
          ) : (
            <button
              type="button"
              onClick={action.onClick}
              className="mt-4 inline-block text-sm font-medium text-indigo-200 hover:text-indigo-100"
            >
              {action.label}
            </button>
          ))}
      </div>
    </div>
  );
}
