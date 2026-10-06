import type { ReactNode } from 'react';
import { Button } from '#/components/ui/button';

export function QueryFeedback({
  query,
  label,
  children,
  failClosed = false,
}: {
  query: {
    isLoading: boolean;
    isError: boolean;
    data?: unknown;
    refetch: () => unknown;
  };
  label: string;
  children: ReactNode;
  failClosed?: boolean;
}) {
  if (query.isLoading)
    return (
      <p role="status" className="p-6 text-sm text-muted-foreground">
        Loading {label}...
      </p>
    );
  if (query.isError && (failClosed || query.data === undefined))
    return (
      <div
        role="alert"
        className="space-y-3 rounded-xl border border-border p-6"
      >
        <p className="text-sm">Could not load {label}.</p>
        <Button variant="outline" onClick={() => void query.refetch()}>
          Try again
        </Button>
      </div>
    );
  return children;
}
